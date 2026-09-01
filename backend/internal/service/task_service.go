package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"gorm.io/gorm"
)

type CreateTaskInput struct {
	ProjectID        uuid.UUID
	Type             model.TaskType
	Title            string
	Description      string
	Status           model.TaskStatus
	Priority         model.TaskPriority
	Labels           []string
	StepsToReproduce *string
	Severity         *model.BugSeverity
	Environment      *string
	InitialSubtasks  []string
}

type UpdateTaskInput struct {
	Title            *string
	Description      *string
	Type             *model.TaskType
	Status           *model.TaskStatus
	Priority         *model.TaskPriority
	Labels           *[]string
	StepsToReproduce *string
	Severity         *model.BugSeverity
	Environment      *string
}

type TaskService interface {
	CreateTask(input CreateTaskInput) (*model.Task, error)
	GetTasks(filter repository.TaskFilter) ([]model.Task, error)
	GetTaskByID(id uuid.UUID) (*model.Task, error)
	GetTaskByTicketKey(ticketKey string) (*model.Task, error)
	UpdateTask(id uuid.UUID, input UpdateTaskInput) (*model.Task, error)
	DeleteTask(id uuid.UUID) error
	BlockTask(id uuid.UUID, isBlocked bool, reason string) (*model.Task, error)
	ArchiveTask(id uuid.UUID, isArchived bool) (*model.Task, error)
}

type taskService struct {
	taskRepo    repository.TaskRepository
	projectRepo repository.ProjectRepository
	subtaskRepo repository.SubtaskRepository
}

func NewTaskService(
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	subtaskRepo repository.SubtaskRepository,
) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		subtaskRepo: subtaskRepo,
	}
}

func (s *taskService) CreateTask(input CreateTaskInput) (*model.Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return nil, ErrTaskTitleRequired
	}

	// Validate Type
	if input.Type == "" {
		input.Type = model.TypeTask
	}
	switch input.Type {
	case model.TypeTask, model.TypeBug, model.TypeImprovement, model.TypeSpike:
	default:
		return nil, ErrInvalidTaskType
	}

	// Validate Status
	if input.Status == "" {
		input.Status = model.StatusBacklog
	}
	switch input.Status {
	case model.StatusBacklog, model.StatusInProgress, model.StatusReview, model.StatusDone:
	default:
		return nil, ErrInvalidTaskStatus
	}

	// Validate Priority
	if input.Priority == "" {
		input.Priority = model.PriorityP2
	}
	switch input.Priority {
	case model.PriorityP0, model.PriorityP1, model.PriorityP2, model.PriorityP3:
	default:
		return nil, ErrInvalidTaskPriority
	}

	// Validate Bug severity if provided
	if input.Severity != nil && *input.Severity != "" {
		switch *input.Severity {
		case model.SeverityCritical, model.SeverityMajor, model.SeverityMinor, model.SeverityTrivial:
		default:
			return nil, ErrInvalidBugSeverity
		}
	}

	// Verify project exists
	project, err := s.projectRepo.FindByID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	var createdTask *model.Task

	// Execute atomic ticket generation and task creation inside a transaction
	err = s.projectRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		// Increment project task counter and lock row
		ticketNumber, projectKey, err := s.projectRepo.IncrementTaskCounter(tx, input.ProjectID)
		if err != nil {
			return err
		}

		ticketKey := fmt.Sprintf("%s-%d", projectKey, ticketNumber)

		task := &model.Task{
			ProjectID:        input.ProjectID,
			TicketNumber:     ticketNumber,
			TicketKey:        ticketKey,
			Type:             input.Type,
			Title:            input.Title,
			Description:      input.Description,
			Status:           input.Status,
			Priority:         input.Priority,
			Labels:           input.Labels,
			StepsToReproduce: input.StepsToReproduce,
			Severity:         input.Severity,
			Environment:      input.Environment,
		}

		if err := s.taskRepo.Create(tx, task); err != nil {
			return err
		}

		// Insert initial subtasks if any
		if len(input.InitialSubtasks) > 0 {
			subtasks := make([]model.Subtask, 0, len(input.InitialSubtasks))
			for idx, title := range input.InitialSubtasks {
				title = strings.TrimSpace(title)
				if title != "" {
					subtasks = append(subtasks, model.Subtask{
						TaskID:     task.ID,
						Title:      title,
						IsDone:     false,
						OrderIndex: idx,
					})
				}
			}
			if len(subtasks) > 0 {
				if err := tx.Create(&subtasks).Error; err != nil {
					return err
				}
				task.Subtasks = subtasks
			}
		}

		createdTask = task
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdTask, nil
}

func (s *taskService) GetTasks(filter repository.TaskFilter) ([]model.Task, error) {
	tasks, err := s.taskRepo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	// Calculate total time spent for each task
	for i := range tasks {
		timeSpent, _ := s.taskRepo.GetTotalTimeSpent(tasks[i].ID)
		tasks[i].TotalTimeSpentSeconds = timeSpent
	}

	return tasks, nil
}

func (s *taskService) GetTaskByID(id uuid.UUID) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	timeSpent, _ := s.taskRepo.GetTotalTimeSpent(task.ID)
	task.TotalTimeSpentSeconds = timeSpent

	for i := range task.Subtasks {
		subTime, _ := s.subtaskRepo.GetTotalTimeSpent(task.Subtasks[i].ID)
		task.Subtasks[i].TotalTimeSpentSeconds = subTime
	}

	return task, nil
}

func (s *taskService) GetTaskByTicketKey(ticketKey string) (*model.Task, error) {
	task, err := s.taskRepo.FindByTicketKey(ticketKey)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	timeSpent, _ := s.taskRepo.GetTotalTimeSpent(task.ID)
	task.TotalTimeSpentSeconds = timeSpent

	return task, nil
}

func (s *taskService) UpdateTask(id uuid.UUID, input UpdateTaskInput) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, ErrTaskTitleRequired
		}
		task.Title = title
	}

	if input.Description != nil {
		task.Description = *input.Description
	}

	if input.Type != nil {
		switch *input.Type {
		case model.TypeTask, model.TypeBug, model.TypeImprovement, model.TypeSpike:
			task.Type = *input.Type
		default:
			return nil, ErrInvalidTaskType
		}
	}

	if input.Status != nil {
		switch *input.Status {
		case model.StatusBacklog, model.StatusInProgress, model.StatusReview, model.StatusDone:
			task.Status = *input.Status
		default:
			return nil, ErrInvalidTaskStatus
		}
	}

	if input.Priority != nil {
		switch *input.Priority {
		case model.PriorityP0, model.PriorityP1, model.PriorityP2, model.PriorityP3:
			task.Priority = *input.Priority
		default:
			return nil, ErrInvalidTaskPriority
		}
	}

	if input.Labels != nil {
		task.Labels = *input.Labels
	}

	if input.StepsToReproduce != nil {
		task.StepsToReproduce = input.StepsToReproduce
	}

	if input.Severity != nil {
		if *input.Severity != "" {
			switch *input.Severity {
			case model.SeverityCritical, model.SeverityMajor, model.SeverityMinor, model.SeverityTrivial:
				task.Severity = input.Severity
			default:
				return nil, ErrInvalidBugSeverity
			}
		} else {
			task.Severity = nil
		}
	}

	if input.Environment != nil {
		task.Environment = input.Environment
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) DeleteTask(id uuid.UUID) error {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	return s.taskRepo.Delete(id)
}

func (s *taskService) BlockTask(id uuid.UUID, isBlocked bool, reason string) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	task.IsBlocked = isBlocked
	if isBlocked {
		reason = strings.TrimSpace(reason)
		if reason != "" {
			task.BlockedReason = &reason
		}
	} else {
		task.BlockedReason = nil
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) ArchiveTask(id uuid.UUID, isArchived bool) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	task.IsArchived = isArchived
	if isArchived {
		now := time.Now().UTC()
		task.ArchivedAt = &now
	} else {
		task.ArchivedAt = nil
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}
