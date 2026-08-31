package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TimeEntry struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	TaskID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"task_id"`
	SubtaskID       *uuid.UUID `gorm:"type:uuid;index" json:"subtask_id,omitempty"`
	StartedAt       time.Time  `gorm:"not null" json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds int64      `gorm:"default:0;not null" json:"duration_seconds"`
	IsRunning       bool       `gorm:"default:false;not null" json:"is_running"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (te *TimeEntry) BeforeCreate(tx *gorm.DB) error {
	if te.ID == uuid.Nil {
		te.ID = uuid.New()
	}
	return nil
}
