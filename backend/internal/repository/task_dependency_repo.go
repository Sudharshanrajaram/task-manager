package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type TaskDependencyRepository interface {
	Create(dep *model.TaskDependency) error
	Delete(id uuid.UUID) error
	FindByTaskID(taskID uuid.UUID) ([]model.TaskDependency, error)
	FindDependenciesForTask(taskID uuid.UUID) ([]model.TaskDependency, error)
	Exists(taskID, dependsOnTaskID uuid.UUID) (bool, error)
}

type taskDependencyRepository struct {
	db *gorm.DB
}

func NewTaskDependencyRepository(db *gorm.DB) TaskDependencyRepository {
	return &taskDependencyRepository{db: db}
}

func (r *taskDependencyRepository) Create(dep *model.TaskDependency) error {
	return r.db.Create(dep).Error
}

func (r *taskDependencyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.TaskDependency{}, "id = ?", id).Error
}

func (r *taskDependencyRepository) FindByTaskID(taskID uuid.UUID) ([]model.TaskDependency, error) {
	var deps []model.TaskDependency
	err := r.db.Preload("DependsOnTask").Where("task_id = ?", taskID).Find(&deps).Error
	return deps, err
}

func (r *taskDependencyRepository) FindDependenciesForTask(taskID uuid.UUID) ([]model.TaskDependency, error) {
	var deps []model.TaskDependency
	err := r.db.Preload("Task").Where("depends_on_task_id = ?", taskID).Find(&deps).Error
	return deps, err
}

func (r *taskDependencyRepository) Exists(taskID, dependsOnTaskID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.TaskDependency{}).
		Where("task_id = ? AND depends_on_task_id = ?", taskID, dependsOnTaskID).
		Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return count > 0, nil
}
