package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subtask struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TaskID     uuid.UUID `gorm:"type:uuid;index;not null" json:"task_id"`
	Title      string    `gorm:"type:varchar(255);not null" json:"title"`
	IsDone     bool      `gorm:"default:false;not null" json:"is_done"`
	OrderIndex int       `gorm:"default:0;not null" json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Computed / rollup field
	TotalTimeSpentSeconds int64 `gorm:"-" json:"total_time_spent_seconds"`
}

func (s *Subtask) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
