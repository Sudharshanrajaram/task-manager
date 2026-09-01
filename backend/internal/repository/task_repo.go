package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type TaskFilter struct {
	ProjectID  *uuid.UUID
	Status     *model.TaskStatus
	Type       *model.TaskType
	Priority   *model.TaskPriority
	Search     *string
	IsArchived *bool
}

type TaskRepository interface {
	Create(tx *gorm.DB, task *model.Task) error
	FindAll(filter TaskFilter) ([]model.Task, error)
	FindByID(id uuid.UUID) (*model.Task, error)
	FindByTicketKey(ticketKey string) (*model.Task, error)
	Update(task *model.Task) error
	Delete(id uuid.UUID) error
	GetTotalTimeSpent(taskID uuid.UUID) (int64, error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(tx *gorm.DB, task *model.Task) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(task).Error
}

func (r *taskRepository) FindAll(filter TaskFilter) ([]model.Task, error) {
	var tasks []model.Task
	q := r.db.Preload("Subtasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("subtasks.order_index asc, subtasks.created_at asc")
	})

	if filter.ProjectID != nil {
		q = q.Where("project_id = ?", *filter.ProjectID)
	}
	if filter.Status != nil && *filter.Status != "" {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.Type != nil && *filter.Type != "" {
		q = q.Where("type = ?", *filter.Type)
	}
	if filter.Priority != nil && *filter.Priority != "" {
		q = q.Where("priority = ?", *filter.Priority)
	}
	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		q = q.Where("title LIKE ? OR description LIKE ? OR ticket_key LIKE ?", pattern, pattern, pattern)
	}
	if filter.IsArchived != nil {
		q = q.Where("is_archived = ?", *filter.IsArchived)
	}

	err := q.Order("ticket_number desc").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) FindByID(id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.Preload("Subtasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("subtasks.order_index asc, subtasks.created_at asc")
	}).Preload("Project").First(&task, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindByTicketKey(ticketKey string) (*model.Task, error) {
	var task model.Task
	err := r.db.Preload("Subtasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("subtasks.order_index asc, subtasks.created_at asc")
	}).Preload("Project").First(&task, "ticket_key = ?", ticketKey).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *taskRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Task{}, "id = ?", id).Error
}

func (r *taskRepository) GetTotalTimeSpent(taskID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&model.TimeEntry{}).
		Where("task_id = ?", taskID).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Scan(&total).Error
	return total, err
}
