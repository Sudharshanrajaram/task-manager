package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type SubtaskRepository interface {
	Create(subtask *model.Subtask) error
	CreateBatch(subtasks []model.Subtask) error
	FindByID(id uuid.UUID) (*model.Subtask, error)
	FindByTaskID(taskID uuid.UUID) ([]model.Subtask, error)
	Update(subtask *model.Subtask) error
	Delete(id uuid.UUID) error
	GetNextOrderIndex(taskID uuid.UUID) (int, error)
	Reorder(taskID uuid.UUID, orderedIDs []uuid.UUID) error
	GetTotalTimeSpent(subtaskID uuid.UUID) (int64, error)
}

type subtaskRepository struct {
	db *gorm.DB
}

func NewSubtaskRepository(db *gorm.DB) SubtaskRepository {
	return &subtaskRepository{db: db}
}

func (r *subtaskRepository) Create(subtask *model.Subtask) error {
	return r.db.Create(subtask).Error
}

func (r *subtaskRepository) CreateBatch(subtasks []model.Subtask) error {
	if len(subtasks) == 0 {
		return nil
	}
	return r.db.Create(&subtasks).Error
}

func (r *subtaskRepository) FindByID(id uuid.UUID) (*model.Subtask, error) {
	var subtask model.Subtask
	err := r.db.First(&subtask, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subtask, nil
}

func (r *subtaskRepository) FindByTaskID(taskID uuid.UUID) ([]model.Subtask, error) {
	var subtasks []model.Subtask
	err := r.db.Where("task_id = ?", taskID).Order("order_index asc, created_at asc").Find(&subtasks).Error
	return subtasks, err
}

func (r *subtaskRepository) Update(subtask *model.Subtask) error {
	return r.db.Save(subtask).Error
}

func (r *subtaskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Subtask{}, "id = ?", id).Error
}

func (r *subtaskRepository) GetNextOrderIndex(taskID uuid.UUID) (int, error) {
	var maxOrder *int
	err := r.db.Model(&model.Subtask{}).
		Where("task_id = ?", taskID).
		Select("MAX(order_index)").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder + 1, nil
}

func (r *subtaskRepository) Reorder(taskID uuid.UUID, orderedIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for idx, id := range orderedIDs {
			if err := tx.Model(&model.Subtask{}).
				Where("id = ? AND task_id = ?", id, taskID).
				Update("order_index", idx).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *subtaskRepository) GetTotalTimeSpent(subtaskID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&model.TimeEntry{}).
		Where("subtask_id = ?", subtaskID).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Scan(&total).Error
	return total, err
}
