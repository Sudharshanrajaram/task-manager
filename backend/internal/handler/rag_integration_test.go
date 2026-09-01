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
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
	"github.com/taskflow/backend/pkg/groq"
)

func setupRAGTestRouter(t *testing.T) (*gin.Engine, service.ProjectService, service.TaskService) {
	gin.SetMode(gin.TestMode)

	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)
	embeddingRepo := repository.NewEmbeddingRepository(testDB)

	groqClient := groq.NewClient("", "", "")

	projSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	ragSvc := service.NewRAGService(groqClient, embeddingRepo)

	taskHandler := handler.NewTaskHandler(taskSvc)
	projHandler := handler.NewProjectHandler(projSvc)
	ragHandler := handler.NewRAGHandler(ragSvc, taskSvc, subtaskRepo)

	r := gin.New()
	api := r.Group("/api")
	{
		api.POST("/projects", projHandler.Create)
		api.POST("/projects/:id/tasks", taskHandler.Create)
		api.GET("/tasks/:id", taskHandler.GetByIDOrKey)

		api.POST("/tasks/suggest-subtasks", ragHandler.SuggestSubtasks)
		api.POST("/tasks/:id/suggest-subtasks", ragHandler.SuggestSubtasksForTask)
		api.POST("/tasks/:id/accept-subtasks", ragHandler.AcceptSubtasks)
	}

	return r, projSvc, taskSvc
}

func TestRAG_HTTPFlow(t *testing.T) {
	router, projSvc, taskSvc := setupRAGTestRouter(t)

	proj, _ := projSvc.CreateProject("RAG API Proj", "RAPI", "#4F46E5")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Setup Stripe Payment Flow",
	})

	// 1. POST /api/tasks/suggest-subtasks (Freeform)
	freeReq := handler.SuggestSubtasksRequest{
		Title: "Setup Stripe Webhooks",
		Count: 4,
	}
	bodyJSON, _ := json.Marshal(freeReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/tasks/suggest-subtasks", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SuggestSubtasks failed: %d - %s", w.Code, w.Body.String())
	}

	var freeResult service.SubtaskSuggestionResult
	json.Unmarshal(w.Body.Bytes(), &freeResult)
	if len(freeResult.SuggestedSubtasks) == 0 {
		t.Errorf("Expected non-empty suggestions")
	}

	// 2. POST /api/tasks/:id/suggest-subtasks (Existing Task)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%s/suggest-subtasks", task.ID), bytes.NewReader([]byte(`{"count": 3}`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("SuggestSubtasksForTask failed: %d", w.Code)
	}

	// 3. POST /api/tasks/:id/accept-subtasks (Accept and Train RAG)
	acceptReq := handler.AcceptSubtasksRequest{
		Subtasks: []string{
			"Generate Stripe API keys",
			"Create checkout session endpoint",
			"Test webhook events",
		},
	}
	bodyJSON, _ = json.Marshal(acceptReq)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%s/accept-subtasks", task.ID), bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("AcceptSubtasks failed: %d - %s", w.Code, w.Body.String())
	}

	var updatedTask model.Task
	json.Unmarshal(w.Body.Bytes(), &updatedTask)
	if len(updatedTask.Subtasks) != 3 {
		t.Errorf("Expected 3 subtasks attached to task, got %d", len(updatedTask.Subtasks))
	}
}
