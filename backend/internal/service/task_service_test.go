package service_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func setupTestServices(t *testing.T) (service.ProjectService, service.TaskService, service.SubtaskService) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)

	projSvc := service.NewProjectService(projectRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	subtaskSvc := service.NewSubtaskService(subtaskRepo, taskRepo)

	return projSvc, taskSvc, subtaskSvc
}

func TestProjectAndTaskCreation(t *testing.T) {
	projSvc, taskSvc, _ := setupTestServices(t)

	// 1. Create project
	proj, err := projSvc.CreateProject("Auth Service", "AUTH", "#4F46E5")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if proj.Key != "AUTH" {
		t.Errorf("Expected key AUTH, got %s", proj.Key)
	}

	// 2. Create Task 1
	task1, err := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID:   proj.ID,
		Type:        model.TypeTask,
		Title:       "Setup OAuth2 provider",
		Description: "Configure Google and GitHub OAuth",
		Priority:    model.PriorityP1,
		Status:      model.StatusInProgress,
		Labels:      []string{"backend", "security"},
		InitialSubtasks: []string{
			"Register OAuth apps",
			"Implement callback handler",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task1.TicketKey != "AUTH-1" {
		t.Errorf("Expected ticket key AUTH-1, got %s", task1.TicketKey)
	}
	if task1.TicketNumber != 1 {
		t.Errorf("Expected ticket number 1, got %d", task1.TicketNumber)
	}
	if len(task1.Subtasks) != 2 {
		t.Errorf("Expected 2 initial subtasks, got %d", len(task1.Subtasks))
	}

	// 3. Create Task 2 (Bug)
	steps := "1. Login with invalid token\n2. Observe 500 instead of 401"
	sev := model.SeverityCritical
	env := "production"
	task2, err := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID:        proj.ID,
		Type:             model.TypeBug,
		Title:            "Invalid token returns 500 error",
		Priority:         model.PriorityP0,
		Status:           model.StatusBacklog,
		StepsToReproduce: &steps,
		Severity:         &sev,
		Environment:      &env,
	})
	if err != nil {
		t.Fatalf("Failed to create bug task: %v", err)
	}

	if task2.TicketKey != "AUTH-2" {
		t.Errorf("Expected ticket key AUTH-2, got %s", task2.TicketKey)
	}
	if task2.Severity == nil || *task2.Severity != model.SeverityCritical {
		t.Errorf("Expected critical severity, got %v", task2.Severity)
	}

	// 4. Retrieve by TicketKey
	found, err := taskSvc.GetTaskByTicketKey("AUTH-2")
	if err != nil {
		t.Fatalf("Failed to find task by key: %v", err)
	}
	if found.ID != task2.ID {
		t.Errorf("Found task ID %s does not match %s", found.ID, task2.ID)
	}
}

func TestConcurrentTaskCreationTicketNumbering(t *testing.T) {
	projSvc, taskSvc, _ := setupTestServices(t)

	proj, err := projSvc.CreateProject("Billing Engine", "BILL", "#10B981")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	concurrency := 15
	var wg sync.WaitGroup
	ticketKeys := make(chan string, concurrency)
	errorsChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task, err := taskSvc.CreateTask(service.CreateTaskInput{
				ProjectID: proj.ID,
				Type:      model.TypeTask,
				Title:     fmt.Sprintf("Concurrent Task %d", idx),
				Priority:  model.PriorityP2,
			})
			if err != nil {
				errorsChan <- err
				return
			}
			ticketKeys <- task.TicketKey
		}(i)
	}

	wg.Wait()
	close(ticketKeys)
	close(errorsChan)

	for err := range errorsChan {
		t.Fatalf("Concurrent task creation failed: %v", err)
	}

	seen := make(map[string]bool)
	for key := range ticketKeys {
		if seen[key] {
			t.Fatalf("Duplicate ticket key detected: %s", key)
		}
		seen[key] = true
	}

	if len(seen) != concurrency {
		t.Errorf("Expected %d unique ticket keys, got %d", concurrency, len(seen))
	}
}

func TestSubtasksManagement(t *testing.T) {
	projSvc, taskSvc, subtaskSvc := setupTestServices(t)

	proj, _ := projSvc.CreateProject("Core API", "CORE", "#6366F1")
	task, _ := taskSvc.CreateTask(service.CreateTaskInput{
		ProjectID: proj.ID,
		Title:     "Build Core Endpoints",
	})

	// Add subtasks
	sub1, err := subtaskSvc.CreateSubtask(task.ID, "Write unit tests")
	if err != nil {
		t.Fatalf("Failed to create subtask 1: %v", err)
	}
	sub2, err := subtaskSvc.CreateSubtask(task.ID, "Add OpenAPI docs")
	if err != nil {
		t.Fatalf("Failed to create subtask 2: %v", err)
	}

	if sub1.OrderIndex != 0 || sub2.OrderIndex != 1 {
		t.Errorf("Subtask order indices wrong: %d, %d", sub1.OrderIndex, sub2.OrderIndex)
	}

	// Toggle done
	isDone := true
	updated, err := subtaskSvc.UpdateSubtask(sub1.ID, nil, &isDone)
	if err != nil {
		t.Fatalf("Failed to update subtask: %v", err)
	}
	if !updated.IsDone {
		t.Errorf("Expected isDone to be true")
	}

	// Reorder subtasks: sub2 then sub1
	err = subtaskSvc.ReorderSubtasks(task.ID, []uuid.UUID{sub2.ID, sub1.ID})
	if err != nil {
		t.Fatalf("Failed to reorder subtasks: %v", err)
	}

	// Fetch task with subtasks
	refreshed, _ := taskSvc.GetTaskByID(task.ID)
	if len(refreshed.Subtasks) != 2 {
		t.Fatalf("Expected 2 subtasks, got %d", len(refreshed.Subtasks))
	}
	if refreshed.Subtasks[0].ID != sub2.ID {
		t.Errorf("Expected sub2 to be first after reordering")
	}
}
