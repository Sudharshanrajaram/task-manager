package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type NoteRepository interface {
	FindByTaskID(taskID uuid.UUID) (*model.Note, error)
	FindGlobalScratchpad(userID *uuid.UUID) (*model.Note, error)
	UpsertTaskNote(taskID uuid.UUID, userID *uuid.UUID, content string) (*model.Note, error)
	UpsertGlobalScratchpad(userID *uuid.UUID, content string) (*model.Note, error)
}

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) FindByTaskID(taskID uuid.UUID) (*model.Note, error) {
	var note model.Note
	err := r.db.First(&note, "task_id = ?", taskID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) FindGlobalScratchpad(userID *uuid.UUID) (*model.Note, error) {
	var note model.Note
	query := r.db.Where("task_id IS NULL")
	if userID != nil && *userID != uuid.Nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) UpsertTaskNote(taskID uuid.UUID, userID *uuid.UUID, content string) (*model.Note, error) {
	var note model.Note
	err := r.db.First(&note, "task_id = ?", taskID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		note = model.Note{
			TaskID:  &taskID,
			UserID:  userID,
			Content: content,
		}
		if err := r.db.Create(&note).Error; err != nil {
			return nil, err
		}
	} else {
		note.Content = content
		if userID != nil && *userID != uuid.Nil {
			note.UserID = userID
		}
		if err := r.db.Save(&note).Error; err != nil {
			return nil, err
		}
	}

	return &note, nil
}

func (r *noteRepository) UpsertGlobalScratchpad(userID *uuid.UUID, content string) (*model.Note, error) {
	note, err := r.FindGlobalScratchpad(userID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		newNote := model.Note{
			TaskID:  nil,
			UserID:  userID,
			Content: content,
		}
		if err := r.db.Create(&newNote).Error; err != nil {
			return nil, err
		}
		return &newNote, nil
	}

	note.Content = content
	if err := r.db.Save(note).Error; err != nil {
		return nil, err
	}

	return note, nil
}
