package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func setupTimerTestRouter(t *testing.T) (*gin.Engine, service.ProjectService, service.TaskService, service.SubtaskService) {
	gin.SetMode(gin.TestMode)

	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)
	timeEntryRepo := repository.NewTimeEntryRepository(testDB)

	projSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	subtaskSvc := service.NewSubtaskService(subtaskRepo, taskRepo)
	timerMgr := service.NewTimerManager(timeEntryRepo, taskRepo, subtaskRepo)

	projHandler := handler.NewProjectHandler(projSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	subtaskHandler := handler.NewSubtaskHandler(subtaskSvc)
	timerHandler := handler.NewTimerHandler(timerMgr)

	r := gin.New()
	api := r.Group("/api")
	{
		api.POST("/projects", projHandler.Create)
		api.POST("/projects/:id/tasks", taskHandler.Create)
		api.POST("/tasks/:id/subtasks", subtaskHandler.Create)

		api.POST("/timers/start", timerHandler.Start)
		api.POST("/timers/:id/pause", timerHandler.Pause)
		api.POST("/timers/:id/resume", timerHandler.Resume)
		api.POST("/timers/:id/stop", timerHandler.Stop)
		api.POST("/timers/:id/adjust", timerHandler.Adjust)
		api.GET("/timers/active", timerHandler.GetActive)
		api.GET("/analytics/summary", timerHandler.AnalyticsSummary)
	}

	return r, projSvc, taskSvc, subtaskSvc
}

func TestTimerAPI_FullFlow(t *testing.T) {
	router, projSvc, taskSvc, subtaskSvc := setupTimerTestRouter(t)

	proj, _ := projSvc.CreateProject("Timer API Proj", "TAPI", "#4F46E5")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "API Task for Timer",
	})
	subtask, _ := subtaskSvc.CreateSubtask(task.ID, "API Subtask")

	// 1. Start Task Timer
	startReq := handler.StartTimerRequest{TaskID: task.ID}
	bodyJSON, _ := json.Marshal(startReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to start task timer: %d - %s", w.Code, w.Body.String())
	}

	var taskTimerInfo service.ActiveTimerInfo
	json.Unmarshal(w.Body.Bytes(), &taskTimerInfo)
	if taskTimerInfo.TicketKey != "TAPI-1" {
		t.Errorf("Expected ticket key TAPI-1, got %s", taskTimerInfo.TicketKey)
	}

	// 2. Start Subtask Timer concurrently
	startSubReq := handler.StartTimerRequest{TaskID: task.ID, SubtaskID: &subtask.ID}
	bodyJSON, _ = json.Marshal(startSubReq)
	req, _ = http.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to start subtask timer: %d - %s", w.Code, w.Body.String())
	}

	var subtaskTimerInfo service.ActiveTimerInfo
	json.Unmarshal(w.Body.Bytes(), &subtaskTimerInfo)

	// 3. GET /api/timers/active (should return 2 active timers)
	req, _ = http.NewRequest(http.MethodGet, "/api/timers/active", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get active timers: %d", w.Code)
	}

	var activeList []service.ActiveTimerInfo
	json.Unmarshal(w.Body.Bytes(), &activeList)
	if len(activeList) != 2 {
		t.Errorf("Expected 2 active timers, got %d", len(activeList))
	}

	// 4. Pause subtask timer
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/timers/%s/pause", subtaskTimerInfo.EntryID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to pause subtask timer: %d", w.Code)
	}

	// 5. Resume subtask timer
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/timers/%s/resume", subtaskTimerInfo.EntryID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to resume subtask timer: %d", w.Code)
	}

	// 6. Stop task timer
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/timers/%s/stop", taskTimerInfo.EntryID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to stop task timer: %d", w.Code)
	}

	// 7. Adjust subtask timer (+120s)
	adjustReq := handler.AdjustTimerRequest{DeltaSeconds: 120}
	bodyJSON, _ = json.Marshal(adjustReq)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/timers/%s/adjust", subtaskTimerInfo.EntryID), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to adjust timer: %d", w.Code)
	}

	// 8. Stop subtask timer
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/timers/%s/stop", subtaskTimerInfo.EntryID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Failed to stop subtask timer: %d", w.Code)
	}

	// 9. Analytics Summary
	req, _ = http.NewRequest(http.MethodGet, "/api/analytics/summary?range=week", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to fetch analytics summary: %d", w.Code)
	}

	var summary repository.AnalyticsSummary
	json.Unmarshal(w.Body.Bytes(), &summary)
	if summary.TotalTimeSpentSeconds < 120 {
		t.Errorf("Expected summary total time >= 120s, got %d", summary.TotalTimeSpentSeconds)
	}
}

func TestTimerAPI_DuplicateProtection(t *testing.T) {
	router, projSvc, taskSvc, _ := setupTimerTestRouter(t)

	proj, _ := projSvc.CreateProject("Dup Test Proj", "DUP", "#10B981")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Dup Task",
	})

	// Start timer first time -> 201 Created
	startReq := handler.StartTimerRequest{TaskID: task.ID}
	bodyJSON, _ := json.Marshal(startReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on first timer start, got %d", w.Code)
	}

	// Start timer again on same task -> 409 Conflict
	req, _ = http.NewRequest(http.MethodPost, "/api/timers/start", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict on duplicate timer start, got %d", w.Code)
	}
}
