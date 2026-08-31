package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskType string

const (
	TypeTask        TaskType = "task"
	TypeBug         TaskType = "bug"
	TypeImprovement TaskType = "improvement"
	TypeSpike       TaskType = "spike"
)

type TaskStatus string

const (
	StatusBacklog    TaskStatus = "backlog"
	StatusInProgress TaskStatus = "in_progress"
	StatusBlocked    TaskStatus = "blocked"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityP0 TaskPriority = "p0" // Critical / Urgent
	PriorityP1 TaskPriority = "p1" // High
	PriorityP2 TaskPriority = "p2" // Medium
	PriorityP3 TaskPriority = "p3" // Low
)

type BugSeverity string

const (
	SeverityCritical BugSeverity = "critical"
	SeverityMajor    BugSeverity = "major"
	SeverityMinor    BugSeverity = "minor"
	SeverityTrivial  BugSeverity = "trivial"
)

type Task struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	ProjectID    uuid.UUID    `gorm:"type:uuid;index;not null" json:"project_id"`
	TicketNumber int          `gorm:"not null" json:"ticket_number"`
	TicketKey    string       `gorm:"type:varchar(20);index;not null" json:"ticket_key"` // e.g. "AUTH-142"
	Type         TaskType     `gorm:"type:varchar(20);not null;default:'task'" json:"type"`
	Title        string       `gorm:"type:varchar(255);not null" json:"title"`
	Description  string       `gorm:"type:text" json:"description"`
	Status       TaskStatus   `gorm:"type:varchar(20);not null;default:'backlog'" json:"status"`
	Priority     TaskPriority `gorm:"type:varchar(10);not null;default:'p2'" json:"priority"`
	Labels       StringArray  `gorm:"type:text" json:"labels"`

	// Bug-specific fields
	StepsToReproduce *string      `gorm:"type:text" json:"steps_to_reproduce,omitempty"`
	Severity         *BugSeverity `gorm:"type:varchar(20)" json:"severity,omitempty"`
	Environment      *string      `gorm:"type:varchar(100)" json:"environment,omitempty"`

	// Relations
	Project     *Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Subtasks    []Subtask   `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;" json:"subtasks"`
	TimeEntries []TimeEntry `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;" json:"time_entries,omitempty"`

	// Computed / rollup field (not stored directly in DB)
	TotalTimeSpentSeconds int64 `gorm:"-" json:"total_time_spent_seconds"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Labels == nil {
		t.Labels = StringArray{}
	}
	return nil
}
