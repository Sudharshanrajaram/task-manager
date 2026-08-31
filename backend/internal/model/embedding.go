package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskTitleEmbedding struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;" json:"id"`
	ProjectID     *uuid.UUID  `gorm:"type:uuid;index" json:"project_id,omitempty"`
	SourceTitle   string      `gorm:"type:text;not null" json:"source_title"`
	FinalSubtasks StringArray `gorm:"type:text;not null" json:"final_subtasks"`
	EmbeddingJSON string      `gorm:"type:text" json:"embedding_json,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (e *TaskTitleEmbedding) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
