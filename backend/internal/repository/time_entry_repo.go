package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/gorm"
)

type ProjectTimeSummary struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	ProjectKey  string    `json:"project_key"`
	Color       string    `json:"color"`
	TotalTime   int64     `json:"total_time_seconds"`
}

type TypeTimeSummary struct {
	Type      model.TaskType `json:"type"`
	TotalTime int64          `json:"total_time_seconds"`
}

type AnalyticsSummary struct {
	TotalTimeSpentSeconds int64                `json:"total_time_spent_seconds"`
	ByProject             []ProjectTimeSummary `json:"by_project"`
	ByType                []TypeTimeSummary    `json:"by_type"`
	Range                 string               `json:"range"`
}

type TimeEntryRepository interface {
	Create(entry *model.TimeEntry) error
	FindByID(id uuid.UUID) (*model.TimeEntry, error)
	FindRunningByTaskOrSubtask(taskID uuid.UUID, subtaskID *uuid.UUID) (*model.TimeEntry, error)
	FindAllRunning() ([]model.TimeEntry, error)
	Update(entry *model.TimeEntry) error
	Delete(id uuid.UUID) error
	GetAnalyticsSummary(since time.Time, rangeLabel string) (*AnalyticsSummary, error)
}

type timeEntryRepository struct {
	db *gorm.DB
}

func NewTimeEntryRepository(db *gorm.DB) TimeEntryRepository {
	return &timeEntryRepository{db: db}
}

func (r *timeEntryRepository) Create(entry *model.TimeEntry) error {
	return r.db.Create(entry).Error
}

func (r *timeEntryRepository) FindByID(id uuid.UUID) (*model.TimeEntry, error) {
	var entry model.TimeEntry
	err := r.db.First(&entry, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (r *timeEntryRepository) FindRunningByTaskOrSubtask(taskID uuid.UUID, subtaskID *uuid.UUID) (*model.TimeEntry, error) {
	var entry model.TimeEntry
	q := r.db.Where("task_id = ? AND is_running = true", taskID)
	if subtaskID != nil {
		q = q.Where("subtask_id = ?", *subtaskID)
	} else {
		q = q.Where("subtask_id IS NULL")
	}

	err := q.First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (r *timeEntryRepository) FindAllRunning() ([]model.TimeEntry, error) {
	var entries []model.TimeEntry
	err := r.db.Where("is_running = true").Find(&entries).Error
	return entries, err
}

func (r *timeEntryRepository) Update(entry *model.TimeEntry) error {
	return r.db.Save(entry).Error
}

func (r *timeEntryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.TimeEntry{}, "id = ?", id).Error
}

func (r *timeEntryRepository) GetAnalyticsSummary(since time.Time, rangeLabel string) (*AnalyticsSummary, error) {
	summary := &AnalyticsSummary{
		Range:     rangeLabel,
		ByProject: []ProjectTimeSummary{},
		ByType:    []TypeTimeSummary{},
	}

	// 1. Total Time
	err := r.db.Model(&model.TimeEntry{}).
		Where("started_at >= ?", since).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Scan(&summary.TotalTimeSpentSeconds).Error
	if err != nil {
		return nil, err
	}

	// 2. Aggregate By Project
	type projResult struct {
		ProjectID   uuid.UUID
		ProjectName string
		ProjectKey  string
		Color       string
		TotalTime   int64
	}
	var projResults []projResult
	err = r.db.Table("time_entries").
		Select("projects.id as project_id, projects.name as project_name, projects.key as project_key, projects.color as color, COALESCE(SUM(time_entries.duration_seconds), 0) as total_time").
		Joins("JOIN tasks ON tasks.id = time_entries.task_id").
		Joins("JOIN projects ON projects.id = tasks.project_id").
		Where("time_entries.started_at >= ?", since).
		Group("projects.id, projects.name, projects.key, projects.color").
		Order("total_time desc").
		Scan(&projResults).Error

	if err != nil {
		return nil, err
	}
	for _, pr := range projResults {
		summary.ByProject = append(summary.ByProject, ProjectTimeSummary{
			ProjectID:   pr.ProjectID,
			ProjectName: pr.ProjectName,
			ProjectKey:  pr.ProjectKey,
			Color:       pr.Color,
			TotalTime:   pr.TotalTime,
		})
	}

	// 3. Aggregate By Ticket Type
	type typeResult struct {
		Type      model.TaskType
		TotalTime int64
	}
	var typeResults []typeResult
	err = r.db.Table("time_entries").
		Select("tasks.type as type, COALESCE(SUM(time_entries.duration_seconds), 0) as total_time").
		Joins("JOIN tasks ON tasks.id = time_entries.task_id").
		Where("time_entries.started_at >= ?", since).
		Group("tasks.type").
		Order("total_time desc").
		Scan(&typeResults).Error

	if err != nil {
		return nil, err
	}
	for _, tr := range typeResults {
		summary.ByType = append(summary.ByType, TypeTimeSummary{
			Type:      tr.Type,
			TotalTime: tr.TotalTime,
		})
	}

	return summary, nil
}
