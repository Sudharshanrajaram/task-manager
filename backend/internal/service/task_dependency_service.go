package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

var (
	ErrSelfDependency     = errors.New("a task cannot depend on itself")
	ErrDependencyExists   = errors.New("this dependency relationship already exists")
	ErrCircularDependency = errors.New("circular task dependency detected")
	ErrDependencyNotFound = errors.New("task dependency not found")
)

type TaskDependencyService interface {
	AddDependency(taskID, dependsOnTaskID uuid.UUID) (*model.TaskDependency, error)
	RemoveDependency(depID uuid.UUID) error
	GetDependencies(taskID uuid.UUID) (blockedBy []model.TaskDependency, blocks []model.TaskDependency, err error)
}

type taskDependencyService struct {
	depRepo  repository.TaskDependencyRepository
	taskRepo repository.TaskRepository
}

func NewTaskDependencyService(
	depRepo repository.TaskDependencyRepository,
	taskRepo repository.TaskRepository,
) TaskDependencyService {
	return &taskDependencyService{
		depRepo:  depRepo,
		taskRepo: taskRepo,
	}
}

func (s *taskDependencyService) AddDependency(taskID, dependsOnTaskID uuid.UUID) (*model.TaskDependency, error) {
	if taskID == dependsOnTaskID {
		return nil, ErrSelfDependency
	}

	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	depTask, err := s.taskRepo.FindByID(dependsOnTaskID)
	if err != nil {
		return nil, err
	}
	if depTask == nil {
		return nil, errors.New("dependency target task not found")
	}

	// Check for existing
	exists, err := s.depRepo.Exists(taskID, dependsOnTaskID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDependencyExists
	}

	// Check reverse (cycle of length 2)
	reverseExists, err := s.depRepo.Exists(dependsOnTaskID, taskID)
	if err != nil {
		return nil, err
	}
	if reverseExists {
		return nil, ErrCircularDependency
	}

	dep := &model.TaskDependency{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnTaskID,
	}

	if err := s.depRepo.Create(dep); err != nil {
		return nil, err
	}

	dep.DependsOnTask = depTask
	return dep, nil
}

func (s *taskDependencyService) RemoveDependency(depID uuid.UUID) error {
	return s.depRepo.Delete(depID)
}

func (s *taskDependencyService) GetDependencies(taskID uuid.UUID) ([]model.TaskDependency, []model.TaskDependency, error) {
	blockedBy, err := s.depRepo.FindByTaskID(taskID)
	if err != nil {
		return nil, nil, err
	}

	blocks, err := s.depRepo.FindDependenciesForTask(taskID)
	if err != nil {
		return nil, nil, err
	}

	return blockedBy, blocks, nil
}
