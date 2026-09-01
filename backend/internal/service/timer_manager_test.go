package service_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func setupTimerTestEnvironment(t *testing.T) (service.ProjectService, service.TaskService, service.SubtaskService, service.TimerManager, repository.TimeEntryRepository) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)
	timeEntryRepo := repository.NewTimeEntryRepository(testDB)

	projSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	subtaskSvc := service.NewSubtaskService(subtaskRepo, taskRepo)
	timerMgr := service.NewTimerManager(timeEntryRepo, taskRepo, subtaskRepo)

	return projSvc, taskSvc, subtaskSvc, timerMgr, timeEntryRepo
}

func TestTimerLifecycle_StartPauseResumeStop(t *testing.T) {
	projSvc, taskSvc, _, timerMgr, timeEntryRepo := setupTimerTestEnvironment(t)

	proj, err := projSvc.CreateProject("Timer Test Proj", "TIME", "#4F46E5")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	task, err := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Track this task",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 1. Start Timer
	entry, err := timerMgr.StartTimer(task.ID, nil)
	if err != nil {
		t.Fatalf("StartTimer failed: %v", err)
	}
	if !entry.IsRunning {
		t.Errorf("Expected timer entry to be running")
	}

	// Check active list
	activeList := timerMgr.GetActiveTimers()
	if len(activeList) != 1 {
		t.Fatalf("Expected 1 active timer, got %d", len(activeList))
	}
	if activeList[0].TicketKey != "TIME-1" {
		t.Errorf("Expected ticket key TIME-1, got %s", activeList[0].TicketKey)
	}

	// Wait 50ms
	time.Sleep(50 * time.Millisecond)

	// 2. Pause Timer
	pausedEntry, err := timerMgr.PauseTimer(entry.ID)
	if err != nil {
		t.Fatalf("PauseTimer failed: %v", err)
	}

	info, err := timerMgr.GetTimerByEntryID(entry.ID)
	if err != nil || !info.IsPaused {
		t.Errorf("Expected timer to be paused in memory")
	}

	// 3. Resume Timer
	time.Sleep(20 * time.Millisecond)
	_, err = timerMgr.ResumeTimer(entry.ID)
	if err != nil {
		t.Fatalf("ResumeTimer failed: %v", err)
	}

	info, _ = timerMgr.GetTimerByEntryID(entry.ID)
	if info.IsPaused {
		t.Errorf("Expected timer to be running after resume")
	}

	// 4. Stop Timer
	stoppedEntry, err := timerMgr.StopTimer(entry.ID)
	if err != nil {
		t.Fatalf("StopTimer failed: %v", err)
	}
	if stoppedEntry.IsRunning {
		t.Errorf("Expected stopped entry to not be running")
	}

	// Verify active list is now empty
	if len(timerMgr.GetActiveTimers()) != 0 {
		t.Errorf("Expected 0 active timers after stop, got %d", len(timerMgr.GetActiveTimers()))
	}

	// Check persistence in DB
	dbEntry, err := timeEntryRepo.FindByID(entry.ID)
	if err != nil || dbEntry == nil {
		t.Fatalf("Failed to fetch time entry from DB: %v", err)
	}
	if dbEntry.IsRunning {
		t.Errorf("Expected DB entry is_running to be false")
	}
	if dbEntry.EndedAt == nil {
		t.Errorf("Expected DB entry ended_at to be set")
	}

	_ = pausedEntry
}

func TestConcurrentTimers_RaceDetector(t *testing.T) {
	projSvc, taskSvc, subtaskSvc, timerMgr, _ := setupTimerTestEnvironment(t)

	proj, _ := projSvc.CreateProject("Concurrent Project", "CONC", "#10B981")

	// Pre-create 10 tasks, each with 1 subtask
	type taskSubPair struct {
		taskID    uuid.UUID
		subtaskID uuid.UUID
	}
	pairs := make([]taskSubPair, 10)
	for i := 0; i < 10; i++ {
		task, _ := taskSvc.CreateTask(service.CreateTaskInput{
			ProjectID: proj.ID,
			Title:     "Concurrent Task",
		})
		sub, _ := subtaskSvc.CreateSubtask(task.ID, "Concurrent Subtask")
		pairs[i] = taskSubPair{taskID: task.ID, subtaskID: sub.ID}
	}

	var wg sync.WaitGroup
	entries := make([]uuid.UUID, 0, 20)
	var entriesMu sync.Mutex

	// Launch 10 task timers and 10 subtask timers concurrently
	for i := 0; i < 10; i++ {
		// Task timer
		wg.Add(1)
		go func(tID uuid.UUID) {
			defer wg.Done()
			entry, err := timerMgr.StartTimer(tID, nil)
			if err == nil && entry != nil {
				entriesMu.Lock()
				entries = append(entries, entry.ID)
				entriesMu.Unlock()
			}
		}(pairs[i].taskID)

		// Subtask timer
		wg.Add(1)
		go func(tID, sID uuid.UUID) {
			defer wg.Done()
			entry, err := timerMgr.StartTimer(tID, &sID)
			if err == nil && entry != nil {
				entriesMu.Lock()
				entries = append(entries, entry.ID)
				entriesMu.Unlock()
			}
		}(pairs[i].taskID, pairs[i].subtaskID)
	}

	// Concurrently query active timers while they are starting
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = timerMgr.GetActiveTimers()
				time.Sleep(2 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Verify all 20 timers were registered
	if len(entries) != 20 {
		t.Fatalf("Expected 20 running timers, got %d", len(entries))
	}

	// Concurrently pause, resume, and stop all timers
	for _, entryID := range entries {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			_, _ = timerMgr.PauseTimer(id)
			_, _ = timerMgr.ResumeTimer(id)
			_, _ = timerMgr.StopTimer(id)
		}(entryID)
	}

	wg.Wait()

	// Verify all timers cleanly terminated
	activeRemaining := timerMgr.GetActiveTimers()
	if len(activeRemaining) != 0 {
		t.Errorf("Expected 0 active timers after stopping all, got %d", len(activeRemaining))
	}
}

func TestTimerManager_CrashRecovery(t *testing.T) {
	projSvc, taskSvc, _, _, timeEntryRepo := setupTimerTestEnvironment(t)

	proj, _ := projSvc.CreateProject("Recovery Project", "REC", "#F59E0B")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Recovery Task",
	})

	// Simulate an in-flight timer that was left running when server crashed 10 minutes ago
	tenMinutesAgo := time.Now().UTC().Add(-10 * time.Minute)
	orphanEntry := &model.TimeEntry{
		TaskID:          task.ID,
		StartedAt:       tenMinutesAgo,
		DurationSeconds: 600,
		IsRunning:       true,
	}
	if err := timeEntryRepo.Create(orphanEntry); err != nil {
		t.Fatalf("Failed to seed orphan time entry: %v", err)
	}

	// Create brand new TimerManager (simulating server reboot)
	testDB, _ := db.InitTestDB()
	freshRepo := repository.NewTimeEntryRepository(testDB)
	freshTaskRepo := repository.NewTaskRepository(testDB)
	freshSubRepo := repository.NewSubtaskRepository(testDB)
	freshMgr := service.NewTimerManager(freshRepo, freshTaskRepo, freshSubRepo)

	// Recover timers
	// First save the task in fresh DB so foreign key / task lookup succeeds
	_ = testDB.Create(proj)
	_ = testDB.Create(task)
	_ = freshRepo.Create(orphanEntry)

	if err := freshMgr.RecoverInFlightTimers(); err != nil {
		t.Fatalf("RecoverInFlightTimers failed: %v", err)
	}

	activeList := freshMgr.GetActiveTimers()
	if len(activeList) != 1 {
		t.Fatalf("Expected 1 recovered active timer, got %d", len(activeList))
	}

	recovered := activeList[0]
	if recovered.TicketKey != "REC-1" {
		t.Errorf("Expected ticket key REC-1, got %s", recovered.TicketKey)
	}
	if recovered.ElapsedSeconds < 600 {
		t.Errorf("Expected elapsed seconds >= 600, got %d", recovered.ElapsedSeconds)
	}

	// Clean up recovered timer
	_, err := freshMgr.StopTimer(orphanEntry.ID)
	if err != nil {
		t.Fatalf("Failed to stop recovered timer: %v", err)
	}
}

func TestTimerEvents_BroadcastSubscription(t *testing.T) {
	projSvc, taskSvc, _, timerMgr, _ := setupTimerTestEnvironment(t)

	proj, _ := projSvc.CreateProject("Events Project", "EVT", "#3B82F6")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Events Task",
	})

	eventChan, unsubscribe := timerMgr.Subscribe()
	defer unsubscribe()

	// Start timer
	entry, err := timerMgr.StartTimer(task.ID, nil)
	if err != nil {
		t.Fatalf("StartTimer failed: %v", err)
	}

	select {
	case event := <-eventChan:
		if event.Type != service.TimerEventStarted {
			t.Errorf("Expected TimerEventStarted, got %s", event.Type)
		}
		if event.Timer.TicketKey != "EVT-1" {
			t.Errorf("Expected ticket key EVT-1, got %s", event.Timer.TicketKey)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for TimerEventStarted")
	}

	// Stop timer
	_, _ = timerMgr.StopTimer(entry.ID)

	select {
	case event := <-eventChan:
		if event.Type != service.TimerEventStopped {
			t.Errorf("Expected TimerEventStopped, got %s", event.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for TimerEventStopped")
	}
}

func TestTimerAdjustment(t *testing.T) {
	projSvc, taskSvc, _, timerMgr, timeEntryRepo := setupTimerTestEnvironment(t)

	proj, _ := projSvc.CreateProject("Adjust Project", "ADJ", "#EC4899")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Adjust Task",
	})

	entry, _ := timerMgr.StartTimer(task.ID, nil)

	// Add 300 seconds (+5 minutes)
	updatedEntry, err := timerMgr.AdjustTime(entry.ID, 300)
	if err != nil {
		t.Fatalf("AdjustTime failed: %v", err)
	}
	if updatedEntry.DurationSeconds != 300 {
		t.Errorf("Expected duration 300, got %d", updatedEntry.DurationSeconds)
	}

	// Stop and verify in DB
	_, _ = timerMgr.StopTimer(entry.ID)
	dbEntry, _ := timeEntryRepo.FindByID(entry.ID)
	if dbEntry.DurationSeconds < 300 {
		t.Errorf("Expected persisted duration >= 300, got %d", dbEntry.DurationSeconds)
	}
}
