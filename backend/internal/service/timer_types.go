package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TimerEventType string

const (
	TimerEventStarted   TimerEventType = "timer.started"
	TimerEventPaused    TimerEventType = "timer.paused"
	TimerEventResumed   TimerEventType = "timer.resumed"
	TimerEventStopped   TimerEventType = "timer.stopped"
	TimerEventAdjusted  TimerEventType = "timer.adjusted"
	TimerEventIdleAlert TimerEventType = "timer.idle_alert"
)

// ActiveTimer is a thread-safe in-memory representation of a running timer
type ActiveTimer struct {
	EntryID             uuid.UUID
	TaskID              uuid.UUID
	SubtaskID           *uuid.UUID
	TaskTitle           string
	TicketKey           string
	ProjectKey          string
	ProjectName         string
	ProjectColor        string
	SubtaskTitle        *string
	StartedAt           time.Time
	BaseDurationSeconds int64
	IsPaused            bool

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// ElapsedSeconds calculates current live elapsed time safely
func (t *ActiveTimer) ElapsedSeconds() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.IsPaused {
		return t.BaseDurationSeconds
	}
	runningDelta := int64(time.Since(t.StartedAt).Seconds())
	if runningDelta < 0 {
		runningDelta = 0
	}
	return t.BaseDurationSeconds + runningDelta
}

// Snapshot returns a point-in-time value copy for API responses
func (t *ActiveTimer) Snapshot() ActiveTimerInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.snapshotLocked()
}

// SnapshotLocked returns a point-in-time value copy when t.mu is already locked
func (t *ActiveTimer) SnapshotLocked() ActiveTimerInfo {
	return t.snapshotLocked()
}

func (t *ActiveTimer) snapshotLocked() ActiveTimerInfo {
	var subTitle *string
	if t.SubtaskTitle != nil {
		copyTitle := *t.SubtaskTitle
		subTitle = &copyTitle
	}

	var subID *uuid.UUID
	if t.SubtaskID != nil {
		copyID := *t.SubtaskID
		subID = &copyID
	}

	elapsed := t.BaseDurationSeconds
	if !t.IsPaused {
		runningDelta := int64(time.Since(t.StartedAt).Seconds())
		if runningDelta > 0 {
			elapsed += runningDelta
		}
	}

	return ActiveTimerInfo{
		EntryID:        t.EntryID,
		TaskID:         t.TaskID,
		SubtaskID:      subID,
		TaskTitle:      t.TaskTitle,
		TicketKey:      t.TicketKey,
		ProjectKey:     t.ProjectKey,
		ProjectName:    t.ProjectName,
		ProjectColor:   t.ProjectColor,
		SubtaskTitle:   subTitle,
		StartedAt:      t.StartedAt,
		ElapsedSeconds: elapsed,
		IsPaused:       t.IsPaused,
	}
}

// ActiveTimerInfo is a serializable snapshot of an active timer
type ActiveTimerInfo struct {
	EntryID        uuid.UUID  `json:"entry_id"`
	TaskID         uuid.UUID  `json:"task_id"`
	SubtaskID      *uuid.UUID `json:"subtask_id,omitempty"`
	TaskTitle      string     `json:"task_title"`
	TicketKey      string     `json:"ticket_key"`
	ProjectKey     string     `json:"project_key"`
	ProjectName    string     `json:"project_name"`
	ProjectColor   string     `json:"project_color"`
	SubtaskTitle   *string    `json:"subtask_title,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	ElapsedSeconds int64      `json:"elapsed_seconds"`
	IsPaused       bool       `json:"is_paused"`
}

// TimerEvent is published to subscribers and the WebSocket hub
type TimerEvent struct {
	Type      TimerEventType  `json:"type"`
	Timer     ActiveTimerInfo `json:"timer"`
	Timestamp time.Time       `json:"timestamp"`
	Message   string          `json:"message,omitempty"`
}
