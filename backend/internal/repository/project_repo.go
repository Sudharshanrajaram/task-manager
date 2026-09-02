package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProjectRepository interface {
	Create(project *model.Project) error
	FindAll() ([]model.Project, error)
	FindByID(id uuid.UUID) (*model.Project, error)
	FindByKey(key string) (*model.Project, error)
	Update(project *model.Project) error
	Delete(id uuid.UUID) error
	Restore(id uuid.UUID) error
	IncrementTaskCounter(tx *gorm.DB, projectID uuid.UUID) (int, string, error)
	GetDB() *gorm.DB
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *projectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) FindAll() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Where("is_deleted = ?", false).Order("created_at asc").Find(&projects).Error
	return projects, err
}

func (r *projectRepository) FindByID(id uuid.UUID) (*model.Project, error) {
	var project model.Project
	err := r.db.First(&project, "id = ? AND is_deleted = ?", id, false).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindByKey(key string) (*model.Project, error) {
	var project model.Project
	err := r.db.First(&project, "key = ? AND is_deleted = ?", key, false).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) Update(project *model.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&model.Project{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_deleted": true,
	}).Error
}

func (r *projectRepository) Restore(id uuid.UUID) error {
	return r.db.Model(&model.Project{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	}).Error
}

// IncrementTaskCounter locks the project row, increments the counter, and returns the new count and project key
func (r *projectRepository) IncrementTaskCounter(tx *gorm.DB, projectID uuid.UUID) (int, string, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var project model.Project
	// Lock the row for update to prevent race conditions during ticket creation
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&project, "id = ?", projectID).Error; err != nil {
		return 0, "", fmt.Errorf("failed to lock project: %w", err)
	}

	project.TaskCounter++
	if err := db.Model(&project).Update("task_counter", project.TaskCounter).Error; err != nil {
		return 0, "", fmt.Errorf("failed to increment task counter: %w", err)
	}

	return project.TaskCounter, project.Key, nil
}
