package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)

	projSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	subtaskSvc := service.NewSubtaskService(subtaskRepo, taskRepo)

	projHandler := handler.NewProjectHandler(projSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	subtaskHandler := handler.NewSubtaskHandler(subtaskSvc)

	r := gin.New()
	api := r.Group("/api")
	{
		api.POST("/projects", projHandler.Create)
		api.GET("/projects", projHandler.List)
		api.GET("/projects/:id", projHandler.GetByID)
		api.PATCH("/projects/:id", projHandler.Update)
		api.DELETE("/projects/:id", projHandler.Delete)

		api.POST("/projects/:id/tasks", taskHandler.Create)
		api.GET("/projects/:id/tasks", taskHandler.ListByProject)

		api.GET("/tasks/:id", taskHandler.GetByIDOrKey)
		api.PATCH("/tasks/:id", taskHandler.Update)
		api.DELETE("/tasks/:id", taskHandler.Delete)

		api.POST("/tasks/:id/subtasks", subtaskHandler.Create)
		api.PUT("/tasks/:id/subtasks/reorder", subtaskHandler.Reorder)
		api.PATCH("/subtasks/:id", subtaskHandler.Update)
		api.DELETE("/subtasks/:id", subtaskHandler.Delete)
	}

	return r
}

func TestE2ECRUD(t *testing.T) {
	router := setupTestRouter(t)

	// 1. Create Project
	projBody := map[string]string{
		"name":  "TaskFlow Web",
		"key":   "TFW",
		"color": "#4F46E5",
	}
	bodyJSON, _ := json.Marshal(projBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create project: %d - %s", w.Code, w.Body.String())
	}

	var createdProj model.Project
	json.Unmarshal(w.Body.Bytes(), &createdProj)
	if createdProj.Key != "TFW" {
		t.Errorf("Expected project key TFW, got %s", createdProj.Key)
	}

	// 2. Create Task
	taskBody := handler.CreateTaskRequest{
		Type:        model.TypeBug,
		Title:       "Fix CSS Layout Overflow",
		Description: "Horizontal scrollbar appears on 1280px screens",
		Priority:    model.PriorityP1,
		Status:      model.StatusInProgress,
		Labels:      []string{"frontend", "ui"},
	}
	bodyJSON, _ = json.Marshal(taskBody)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%s/tasks", createdProj.ID), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create task: %d - %s", w.Code, w.Body.String())
	}

	var createdTask model.Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)
	if createdTask.TicketKey != "TFW-1" {
		t.Errorf("Expected ticket key TFW-1, got %s", createdTask.TicketKey)
	}

	// 3. Create Subtask
	subtaskBody := handler.CreateSubtaskRequest{
		Title: "Inspect container max-width",
	}
	bodyJSON, _ = json.Marshal(subtaskBody)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%s/subtasks", createdTask.ID), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create subtask: %d - %s", w.Code, w.Body.String())
	}

	var createdSubtask model.Subtask
	json.Unmarshal(w.Body.Bytes(), &createdSubtask)

	// 4. Update Subtask (toggle done)
	isDone := true
	updateSubtaskBody := handler.UpdateSubtaskRequest{
		IsDone: &isDone,
	}
	bodyJSON, _ = json.Marshal(updateSubtaskBody)
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/subtasks/%s", createdSubtask.ID), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to update subtask: %d - %s", w.Code, w.Body.String())
	}

	// 5. Query Task by Ticket Key "TFW-1"
	req, _ = http.NewRequest(http.MethodGet, "/api/tasks/TFW-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to fetch task by ticket key: %d - %s", w.Code, w.Body.String())
	}

	var fetchedTask model.Task
	json.Unmarshal(w.Body.Bytes(), &fetchedTask)
	if len(fetchedTask.Subtasks) != 1 || !fetchedTask.Subtasks[0].IsDone {
		t.Errorf("Subtask state in task retrieval mismatch")
	}

	// 6. Delete Subtask
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/subtasks/%s", createdSubtask.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Failed to delete subtask: %d", w.Code)
	}

	// 7. Delete Task
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tasks/%s", createdTask.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Failed to delete task: %d", w.Code)
	}

	// 8. Delete Project
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%s", createdProj.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Failed to delete project: %d", w.Code)
	}
}

func TestValidationErrors(t *testing.T) {
	router := setupTestRouter(t)

	// Invalid Project Key (lowercase, too short)
	projBody := map[string]string{"name": "Bad Project", "key": "a"}
	bodyJSON, _ := json.Marshal(projBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for invalid key, got %d", w.Code)
	}

	// Task without Title
	taskBody := map[string]string{"type": "task"}
	bodyJSON, _ = json.Marshal(taskBody)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%s/tasks", uuid.New()), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for missing title, got %d", w.Code)
	}
}
