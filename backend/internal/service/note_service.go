package service

import (
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type NoteService interface {
	GetTaskNote(taskID uuid.UUID) (*model.Note, error)
	SaveTaskNote(taskID uuid.UUID, userID *uuid.UUID, content string) (*model.Note, error)
	GetGlobalScratchpad(userID *uuid.UUID) (*model.Note, error)
	SaveGlobalScratchpad(userID *uuid.UUID, content string) (*model.Note, error)
}

type noteService struct {
	noteRepo repository.NoteRepository
	taskRepo repository.TaskRepository
}

func NewNoteService(noteRepo repository.NoteRepository, taskRepo repository.TaskRepository) NoteService {
	return &noteService{
		noteRepo: noteRepo,
		taskRepo: taskRepo,
	}
}

func (s *noteService) GetTaskNote(taskID uuid.UUID) (*model.Note, error) {
	note, err := s.noteRepo.FindByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return &model.Note{TaskID: &taskID, Content: ""}, nil
	}
	return note, nil
}

func (s *noteService) SaveTaskNote(taskID uuid.UUID, userID *uuid.UUID, content string) (*model.Note, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	return s.noteRepo.UpsertTaskNote(taskID, userID, content)
}

func (s *noteService) GetGlobalScratchpad(userID *uuid.UUID) (*model.Note, error) {
	note, err := s.noteRepo.FindGlobalScratchpad(userID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return &model.Note{Content: ""}, nil
	}
	return note, nil
}

func (s *noteService) SaveGlobalScratchpad(userID *uuid.UUID, content string) (*model.Note, error) {
	return s.noteRepo.UpsertGlobalScratchpad(userID, content)
}
