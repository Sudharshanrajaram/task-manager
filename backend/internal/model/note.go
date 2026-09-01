package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Note struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	TaskID    *uuid.UUID `gorm:"type:uuid;index" json:"task_id,omitempty"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (n *Note) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
