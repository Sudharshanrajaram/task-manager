package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func TestTaskDependencyService_FlowAndCyclePrevention(t *testing.T) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("InitTestDB failed: %v", err)
	}

	depRepo := repository.NewTaskDependencyRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	depService := service.NewTaskDependencyService(depRepo, taskRepo)

	proj := model.Project{ID: uuid.New(), Name: "Infra", Key: "INFRA"}
	testDB.Create(&proj)

	task1 := model.Task{ID: uuid.New(), ProjectID: proj.ID, TicketKey: "INFRA-1", Title: "Setup VPC"}
	task2 := model.Task{ID: uuid.New(), ProjectID: proj.ID, TicketKey: "INFRA-2", Title: "Deploy EKS Cluster"}
	testDB.Create(&task1)
	testDB.Create(&task2)

	// 1. Cannot depend on self
	_, err = depService.AddDependency(task1.ID, task1.ID)
	if err != service.ErrSelfDependency {
		t.Errorf("Expected ErrSelfDependency, got %v", err)
	}

	// 2. Add valid dependency: task2 depends on task1 (task2 is blocked by task1)
	dep, err := depService.AddDependency(task2.ID, task1.ID)
	if err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}
	if dep.TaskID != task2.ID || dep.DependsOnTaskID != task1.ID {
		t.Errorf("Unexpected dependency: %+v", dep)
	}

	// 3. Duplicate check
	_, err = depService.AddDependency(task2.ID, task1.ID)
	if err != service.ErrDependencyExists {
		t.Errorf("Expected ErrDependencyExists, got %v", err)
	}

	// 4. Circular dependency check: task1 depends on task2
	_, err = depService.AddDependency(task1.ID, task2.ID)
	if err != service.ErrCircularDependency {
		t.Errorf("Expected ErrCircularDependency, got %v", err)
	}

	// 5. Query dependencies
	blockedBy, blocks, err := depService.GetDependencies(task2.ID)
	if err != nil {
		t.Fatalf("GetDependencies failed: %v", err)
	}
	if len(blockedBy) != 1 || blockedBy[0].DependsOnTaskID != task1.ID {
		t.Errorf("Expected task2 to be blocked by task1, got %+v", blockedBy)
	}
	if len(blocks) != 0 {
		t.Errorf("Expected task2 blocks 0 tasks, got %d", len(blocks))
	}

	// 6. Remove dependency
	if err := depService.RemoveDependency(dep.ID); err != nil {
		t.Fatalf("RemoveDependency failed: %v", err)
	}

	blockedByAfter, _, _ := depService.GetDependencies(task2.ID)
	if len(blockedByAfter) != 0 {
		t.Errorf("Expected 0 dependencies after removal, got %d", len(blockedByAfter))
	}
}
