package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type TimerManager interface {
	StartTimer(taskID uuid.UUID, subtaskID *uuid.UUID) (*model.TimeEntry, error)
	PauseTimer(entryID uuid.UUID) (*model.TimeEntry, error)
	ResumeTimer(entryID uuid.UUID) (*model.TimeEntry, error)
	StopTimer(entryID uuid.UUID) (*model.TimeEntry, error)
	AdjustTime(entryID uuid.UUID, deltaSeconds int64) (*model.TimeEntry, error)
	UpdateEntryTime(entryID uuid.UUID, durationSeconds *int64, startedAt, endedAt *time.Time) (*model.TimeEntry, error)
	GetActiveTimers() []ActiveTimerInfo
	GetTimerByEntryID(entryID uuid.UUID) (*ActiveTimerInfo, error)
	RecoverInFlightTimers() error
	GetAnalyticsSummary(rangeType string) (*repository.AnalyticsSummary, error)
	Subscribe() (<-chan TimerEvent, func())
}

type timerManager struct {
	timeEntryRepo repository.TimeEntryRepository
	taskRepo      repository.TaskRepository
	subtaskRepo   repository.SubtaskRepository

	// Registry of currently active timers
	activeTimers map[uuid.UUID]*ActiveTimer
	timersMu     sync.RWMutex

	// Event broadcasting
	subscribers map[chan TimerEvent]struct{}
	subsMu      sync.RWMutex
}

func NewTimerManager(
	timeEntryRepo repository.TimeEntryRepository,
	taskRepo repository.TaskRepository,
	subtaskRepo repository.SubtaskRepository,
) TimerManager {
	return &timerManager{
		timeEntryRepo: timeEntryRepo,
		taskRepo:      taskRepo,
		subtaskRepo:   subtaskRepo,
		activeTimers:  make(map[uuid.UUID]*ActiveTimer),
		subscribers:   make(map[chan TimerEvent]struct{}),
	}
}

// StartTimer creates a new time entry and starts an in-memory active timer goroutine
func (tm *timerManager) StartTimer(taskID uuid.UUID, subtaskID *uuid.UUID) (*model.TimeEntry, error) {
	task, err := tm.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	var subtaskTitle *string
	if subtaskID != nil {
		subtask, err := tm.subtaskRepo.FindByID(*subtaskID)
		if err != nil {
			return nil, err
		}
		if subtask == nil {
			return nil, ErrSubtaskNotFound
		}
		titleCopy := subtask.Title
		subtaskTitle = &titleCopy
	}

	// Check if already running in memory or DB
	tm.timersMu.RLock()
	for _, t := range tm.activeTimers {
		if t.TaskID == taskID {
			if (subtaskID == nil && t.SubtaskID == nil) ||
				(subtaskID != nil && t.SubtaskID != nil && *subtaskID == *t.SubtaskID) {
				tm.timersMu.RUnlock()
				return nil, ErrTimerAlreadyRunning
			}
		}
	}
	tm.timersMu.RUnlock()

	now := time.Now().UTC()
	entry := &model.TimeEntry{
		TaskID:          taskID,
		SubtaskID:       subtaskID,
		StartedAt:       now,
		DurationSeconds: 0,
		IsRunning:       true,
	}

	if err := tm.timeEntryRepo.Create(entry); err != nil {
		return nil, err
	}

	// If task is in backlog, automatically transition it to in_progress
	if task.Status == model.StatusBacklog {
		task.Status = model.StatusInProgress
		_ = tm.taskRepo.Update(task)
	}

	var projectKey, projectName, projectColor string
	if task.Project != nil {
		projectKey = task.Project.Key
		projectName = task.Project.Name
		projectColor = task.Project.Color
	}

	// Create ActiveTimer in registry
	ctx, cancel := context.WithCancel(context.Background())
	activeTimer := &ActiveTimer{
		EntryID:             entry.ID,
		TaskID:              taskID,
		SubtaskID:           subtaskID,
		TaskTitle:           task.Title,
		TicketKey:           task.TicketKey,
		ProjectKey:          projectKey,
		ProjectName:         projectName,
		ProjectColor:        projectColor,
		SubtaskTitle:        subtaskTitle,
		StartedAt:           now,
		BaseDurationSeconds: 0,
		IsPaused:            false,
		ctx:                 ctx,
		cancel:              cancel,
	}

	tm.timersMu.Lock()
	tm.activeTimers[entry.ID] = activeTimer
	tm.timersMu.Unlock()

	// Spawn monitor goroutine
	go tm.monitorTimer(activeTimer)

	// Broadcast started event
	tm.broadcast(TimerEvent{
		Type:      TimerEventStarted,
		Timer:     activeTimer.Snapshot(),
		Timestamp: now,
	})

	return entry, nil
}

// PauseTimer pauses an active timer, freezing its elapsed duration
func (tm *timerManager) PauseTimer(entryID uuid.UUID) (*model.TimeEntry, error) {
	tm.timersMu.RLock()
	activeTimer, exists := tm.activeTimers[entryID]
	tm.timersMu.RUnlock()

	if !exists {
		return nil, ErrTimerNotFound
	}

	activeTimer.mu.Lock()
	if activeTimer.IsPaused {
		activeTimer.mu.Unlock()
		return nil, ErrTimerAlreadyPaused
	}

	now := time.Now().UTC()
	delta := int64(now.Sub(activeTimer.StartedAt).Seconds())
	if delta < 0 {
		delta = 0
	}
	activeTimer.BaseDurationSeconds += delta
	activeTimer.IsPaused = true
	snapshot := activeTimer.SnapshotLocked()
	activeTimer.mu.Unlock()

	// Persist accumulated duration in DB
	entry, err := tm.timeEntryRepo.FindByID(entryID)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		entry.DurationSeconds = activeTimer.BaseDurationSeconds
		_ = tm.timeEntryRepo.Update(entry)
	}

	tm.broadcast(TimerEvent{
		Type:      TimerEventPaused,
		Timer:     snapshot,
		Timestamp: now,
	})

	return entry, nil
}

// ResumeTimer resumes a paused timer
func (tm *timerManager) ResumeTimer(entryID uuid.UUID) (*model.TimeEntry, error) {
	tm.timersMu.RLock()
	activeTimer, exists := tm.activeTimers[entryID]
	tm.timersMu.RUnlock()

	if !exists {
		return nil, ErrTimerNotFound
	}

	activeTimer.mu.Lock()
	if !activeTimer.IsPaused {
		activeTimer.mu.Unlock()
		return nil, ErrTimerNotPaused
	}

	now := time.Now().UTC()
	activeTimer.StartedAt = now
	activeTimer.IsPaused = false
	snapshot := activeTimer.SnapshotLocked()
	activeTimer.mu.Unlock()

	entry, _ := tm.timeEntryRepo.FindByID(entryID)

	tm.broadcast(TimerEvent{
		Type:      TimerEventResumed,
		Timer:     snapshot,
		Timestamp: now,
	})

	return entry, nil
}

// StopTimer stops the timer, commits final elapsed time to DB, and cleans up goroutine
func (tm *timerManager) StopTimer(entryID uuid.UUID) (*model.TimeEntry, error) {
	tm.timersMu.Lock()
	activeTimer, exists := tm.activeTimers[entryID]
	if exists {
		delete(tm.activeTimers, entryID)
	}
	tm.timersMu.Unlock()

	if !exists {
		return nil, ErrTimerNotFound
	}

	// Cancel background monitor goroutine
	activeTimer.cancel()

	finalDuration := activeTimer.ElapsedSeconds()
	now := time.Now().UTC()
	snapshot := activeTimer.Snapshot()
	snapshot.ElapsedSeconds = finalDuration

	// Persist completion in DB
	entry, err := tm.timeEntryRepo.FindByID(entryID)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		entry.DurationSeconds = finalDuration
		entry.IsRunning = false
		entry.EndedAt = &now
		if err := tm.timeEntryRepo.Update(entry); err != nil {
			return nil, err
		}
	}

	tm.broadcast(TimerEvent{
		Type:      TimerEventStopped,
		Timer:     snapshot,
		Timestamp: now,
	})

	return entry, nil
}

// AdjustTime manually adds or removes seconds from a timer entry
func (tm *timerManager) AdjustTime(entryID uuid.UUID, deltaSeconds int64) (*model.TimeEntry, error) {
	entry, err := tm.timeEntryRepo.FindByID(entryID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrTimerNotFound
	}

	entry.DurationSeconds += deltaSeconds
	if entry.DurationSeconds < 0 {
		entry.DurationSeconds = 0
	}

	if err := tm.timeEntryRepo.Update(entry); err != nil {
		return nil, err
	}

	// If currently active in-memory, adjust base duration as well
	tm.timersMu.RLock()
	activeTimer, exists := tm.activeTimers[entryID]
	tm.timersMu.RUnlock()

	if exists {
		activeTimer.mu.Lock()
		activeTimer.BaseDurationSeconds += deltaSeconds
		if activeTimer.BaseDurationSeconds < 0 {
			activeTimer.BaseDurationSeconds = 0
		}
		snapshot := activeTimer.SnapshotLocked()
		activeTimer.mu.Unlock()

		tm.broadcast(TimerEvent{
			Type:      TimerEventAdjusted,
			Timer:     snapshot,
			Timestamp: time.Now().UTC(),
		})
	}

	return entry, nil
}

// UpdateEntryTime updates a time entry's logged duration or started/ended bounds post-facto
func (tm *timerManager) UpdateEntryTime(entryID uuid.UUID, durationSeconds *int64, startedAt, endedAt *time.Time) (*model.TimeEntry, error) {
	entry, err := tm.timeEntryRepo.FindByID(entryID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrTimerNotFound
	}

	var newDuration int64
	if startedAt != nil && endedAt != nil {
		if endedAt.Before(*startedAt) {
			return nil, fmt.Errorf("ended_at cannot be before started_at")
		}
		newDuration = int64(endedAt.Sub(*startedAt).Seconds())
		entry.StartedAt = *startedAt
		entry.EndedAt = endedAt
	} else if durationSeconds != nil {
		if *durationSeconds < 0 {
			return nil, fmt.Errorf("duration_seconds cannot be negative")
		}
		newDuration = *durationSeconds
		if startedAt != nil {
			entry.StartedAt = *startedAt
			end := startedAt.Add(time.Duration(newDuration) * time.Second)
			entry.EndedAt = &end
		} else if entry.EndedAt != nil {
			start := entry.EndedAt.Add(-time.Duration(newDuration) * time.Second)
			entry.StartedAt = start
		}
	} else {
		return nil, fmt.Errorf("either duration_seconds or started_at and ended_at must be provided")
	}

	entry.DurationSeconds = newDuration
	if err := tm.timeEntryRepo.Update(entry); err != nil {
		return nil, err
	}

	// If currently active in-memory, adjust base duration as well
	tm.timersMu.RLock()
	activeTimer, exists := tm.activeTimers[entryID]
	tm.timersMu.RUnlock()

	if exists {
		activeTimer.mu.Lock()
		activeTimer.BaseDurationSeconds = newDuration
		snapshot := activeTimer.SnapshotLocked()
		activeTimer.mu.Unlock()

		tm.broadcast(TimerEvent{
			Type:      TimerEventAdjusted,
			Timer:     snapshot,
			Timestamp: time.Now().UTC(),
		})
	}

	return entry, nil
}

// GetActiveTimers returns snapshots of all currently running timers across all projects
func (tm *timerManager) GetActiveTimers() []ActiveTimerInfo {
	tm.timersMu.RLock()
	defer tm.timersMu.RUnlock()

	list := make([]ActiveTimerInfo, 0, len(tm.activeTimers))
	for _, t := range tm.activeTimers {
		list = append(list, t.Snapshot())
	}
	return list
}

// GetTimerByEntryID returns snapshot of a single active timer
func (tm *timerManager) GetTimerByEntryID(entryID uuid.UUID) (*ActiveTimerInfo, error) {
	tm.timersMu.RLock()
	defer tm.timersMu.RUnlock()

	timer, exists := tm.activeTimers[entryID]
	if !exists {
		return nil, ErrTimerNotFound
	}
	snapshot := timer.Snapshot()
	return &snapshot, nil
}

// RecoverInFlightTimers reconstitutes active in-flight timers from persistent database state on startup
func (tm *timerManager) RecoverInFlightTimers() error {
	runningEntries, err := tm.timeEntryRepo.FindAllRunning()
	if err != nil {
		return fmt.Errorf("failed to query running time entries: %w", err)
	}

	recoveredCount := 0
	for _, entry := range runningEntries {
		task, err := tm.taskRepo.FindByID(entry.TaskID)
		if err != nil || task == nil {
			continue
		}

		var subtaskTitle *string
		if entry.SubtaskID != nil {
			subtask, err := tm.subtaskRepo.FindByID(*entry.SubtaskID)
			if err == nil && subtask != nil {
				titleCopy := subtask.Title
				subtaskTitle = &titleCopy
			}
		}

		// Calculate already elapsed duration from started_at
		elapsed := int64(time.Since(entry.StartedAt).Seconds())
		if elapsed < 0 {
			elapsed = 0
		}

		var projectKey, projectName, projectColor string
		if task.Project != nil {
			projectKey = task.Project.Key
			projectName = task.Project.Name
			projectColor = task.Project.Color
		}

		ctx, cancel := context.WithCancel(context.Background())
		activeTimer := &ActiveTimer{
			EntryID:             entry.ID,
			TaskID:              entry.TaskID,
			SubtaskID:           entry.SubtaskID,
			TaskTitle:           task.Title,
			TicketKey:           task.TicketKey,
			ProjectKey:          projectKey,
			ProjectName:         projectName,
			ProjectColor:        projectColor,
			SubtaskTitle:        subtaskTitle,
			StartedAt:           entry.StartedAt,
			BaseDurationSeconds: entry.DurationSeconds,
			IsPaused:            false,
			ctx:                 ctx,
			cancel:              cancel,
		}

		tm.timersMu.Lock()
		tm.activeTimers[entry.ID] = activeTimer
		tm.timersMu.Unlock()

		go tm.monitorTimer(activeTimer)
		recoveredCount++
	}

	if recoveredCount > 0 {
		slog.Info("Hydrated in-flight timers from database", slog.Int("recovered_count", recoveredCount))
	}
	return nil
}

// monitorTimer runs a background goroutine per active timer for idle detection & lifecycle
func (tm *timerManager) monitorTimer(t *ActiveTimer) {
	// Idle warning triggers if timer runs uninterrupted for > 3 hours
	idleCheckDuration := 3 * time.Hour
	idleTimer := time.NewTimer(idleCheckDuration)
	defer idleTimer.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-idleTimer.C:
			tm.broadcast(TimerEvent{
				Type:      TimerEventIdleAlert,
				Timer:     t.Snapshot(),
				Timestamp: time.Now().UTC(),
				Message:   fmt.Sprintf("Timer on %s (%s) has been running for 3 hours. Still working on this?", t.TicketKey, t.TaskTitle),
			})
		}
	}
}

// Subscribe returns a read-only channel and an unsubscribe cleanup func
func (tm *timerManager) Subscribe() (<-chan TimerEvent, func()) {
	ch := make(chan TimerEvent, 50)
	tm.subsMu.Lock()
	tm.subscribers[ch] = struct{}{}
	tm.subsMu.Unlock()

	unsubscribe := func() {
		tm.subsMu.Lock()
		delete(tm.subscribers, ch)
		close(ch)
		tm.subsMu.Unlock()
	}

	return ch, unsubscribe
}

// broadcast dispatches an event non-blockingly to all active subscribers
func (tm *timerManager) broadcast(event TimerEvent) {
	tm.subsMu.RLock()
	defer tm.subsMu.RUnlock()

	for ch := range tm.subscribers {
		select {
		case ch <- event:
		default:
			// Buffer full, drop to avoid lagging the timer engine
		}
	}
}

func (tm *timerManager) GetAnalyticsSummary(rangeType string) (*repository.AnalyticsSummary, error) {
	now := time.Now().UTC()
	var since time.Time

	switch rangeType {
	case "today":
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "month":
		since = now.AddDate(0, -1, 0)
	case "week":
		fallthrough
	default:
		rangeType = "week"
		since = now.AddDate(0, 0, -7)
	}

	return tm.timeEntryRepo.GetAnalyticsSummary(since, rangeType)
}
