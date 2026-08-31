package service

import (
	"strings"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type SubtaskService interface {
	CreateSubtask(taskID uuid.UUID, title string) (*model.Subtask, error)
	UpdateSubtask(id uuid.UUID, title *string, isDone *bool) (*model.Subtask, error)
	DeleteSubtask(id uuid.UUID) error
	ReorderSubtasks(taskID uuid.UUID, orderedIDs []uuid.UUID) error
}

type subtaskService struct {
	subtaskRepo repository.SubtaskRepository
	taskRepo    repository.TaskRepository
}

func NewSubtaskService(subtaskRepo repository.SubtaskRepository, taskRepo repository.TaskRepository) SubtaskService {
	return &subtaskService{
		subtaskRepo: subtaskRepo,
		taskRepo:    taskRepo,
	}
}

func (s *subtaskService) CreateSubtask(taskID uuid.UUID, title string) (*model.Subtask, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrSubtaskTitleRequired
	}

	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	nextOrder, err := s.subtaskRepo.GetNextOrderIndex(taskID)
	if err != nil {
		return nil, err
	}

	subtask := &model.Subtask{
		TaskID:     taskID,
		Title:      title,
		IsDone:     false,
		OrderIndex: nextOrder,
	}

	if err := s.subtaskRepo.Create(subtask); err != nil {
		return nil, err
	}

	return subtask, nil
}

func (s *subtaskService) UpdateSubtask(id uuid.UUID, title *string, isDone *bool) (*model.Subtask, error) {
	subtask, err := s.subtaskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if subtask == nil {
		return nil, ErrSubtaskNotFound
	}

	if title != nil {
		t := strings.TrimSpace(*title)
		if t == "" {
			return nil, ErrSubtaskTitleRequired
		}
		subtask.Title = t
	}

	if isDone != nil {
		subtask.IsDone = *isDone
	}

	if err := s.subtaskRepo.Update(subtask); err != nil {
		return nil, err
	}

	timeSpent, _ := s.subtaskRepo.GetTotalTimeSpent(subtask.ID)
	subtask.TotalTimeSpentSeconds = timeSpent

	return subtask, nil
}

func (s *subtaskService) DeleteSubtask(id uuid.UUID) error {
	subtask, err := s.subtaskRepo.FindByID(id)
	if err != nil {
		return err
	}
	if subtask == nil {
		return ErrSubtaskNotFound
	}

	return s.subtaskRepo.Delete(id)
}

func (s *subtaskService) ReorderSubtasks(taskID uuid.UUID, orderedIDs []uuid.UUID) error {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	return s.subtaskRepo.Reorder(taskID, orderedIDs)
}
