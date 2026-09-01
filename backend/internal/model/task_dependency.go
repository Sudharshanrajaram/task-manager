package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskDependency struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TaskID          uuid.UUID `gorm:"type:uuid;index;not null" json:"task_id"`
	DependsOnTaskID uuid.UUID `gorm:"type:uuid;index;not null" json:"depends_on_task_id"`
	CreatedAt       time.Time `json:"created_at"`

	Task          *Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;" json:"task,omitempty"`
	DependsOnTask *Task `gorm:"foreignKey:DependsOnTaskID;constraint:OnDelete:CASCADE;" json:"depends_on_task,omitempty"`
}

func (td *TaskDependency) BeforeCreate(tx *gorm.DB) error {
	if td.ID == uuid.Nil {
		td.ID = uuid.New()
	}
	return nil
}
