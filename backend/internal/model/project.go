package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Key         string    `gorm:"type:varchar(10);uniqueIndex;not null" json:"key"`
	Color       string    `gorm:"type:varchar(20);default:'#4F46E5'" json:"color"`
	TaskCounter int       `gorm:"default:0" json:"task_counter"`
	Tasks       []Task    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"tasks,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
