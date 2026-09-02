package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	FindByTaskID(taskID uuid.UUID) ([]model.Comment, error)
	FindByID(id uuid.UUID) (*model.Comment, error)
	Update(comment *model.Comment) error
	Delete(id uuid.UUID) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) FindByTaskID(taskID uuid.UUID) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Where("task_id = ?", taskID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *commentRepository) FindByID(id uuid.UUID) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.First(&comment, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Comment{}, "id = ?", id).Error
}
