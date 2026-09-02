package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateProjectRequest struct {
	Name  string `json:"name" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Color string `json:"color"`
}

type UpdateProjectRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CreateTaskRequest struct {
	Type             model.TaskType     `json:"type"`
	Title            string             `json:"title" binding:"required"`
	Description      string             `json:"description"`
	Status           model.TaskStatus   `json:"status"`
	Priority         model.TaskPriority `json:"priority"`
	Labels           []string           `json:"labels"`
	StepsToReproduce *string            `json:"steps_to_reproduce"`
	Severity         *model.BugSeverity `json:"severity"`
	Environment      *string            `json:"environment"`
	InitialSubtasks  []string           `json:"initial_subtasks"`
}

type UpdateTaskRequest struct {
	Title            *string             `json:"title"`
	Description      *string             `json:"description"`
	Type             *model.TaskType     `json:"type"`
	Status           *model.TaskStatus   `json:"status"`
	Priority         *model.TaskPriority `json:"priority"`
	Labels           *[]string           `json:"labels"`
	StepsToReproduce *string             `json:"steps_to_reproduce"`
	Severity         *model.BugSeverity  `json:"severity"`
	Environment      *string             `json:"environment"`
}

type BlockTaskRequest struct {
	IsBlocked     bool   `json:"is_blocked"`
	BlockedReason string `json:"blocked_reason"`
}

type ArchiveTaskRequest struct {
	IsArchived *bool `json:"is_archived"`
}

type SaveNoteRequest struct {
	Content string `json:"content"`
}

type CreateDependencyRequest struct {
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id" binding:"required"`
}

type CreateSubtaskRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateSubtaskRequest struct {
	Title  *string `json:"title"`
	IsDone *bool   `json:"is_done"`
}

type ReorderSubtasksRequest struct {
	OrderedIDs []uuid.UUID `json:"ordered_ids" binding:"required"`
}

type StartTimerRequest struct {
	TaskID    uuid.UUID  `json:"task_id" binding:"required"`
	SubtaskID *uuid.UUID `json:"subtask_id"`
}

type AdjustTimerRequest struct {
	DeltaSeconds int64 `json:"delta_seconds" binding:"required"`
}

type UpdateTimerRequest struct {
	DurationSeconds *int64     `json:"duration_seconds"`
	StartedAt       *time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
}

type SuggestSubtasksRequest struct {
	Title     string     `json:"title"`
	ProjectID *uuid.UUID `json:"project_id"`
	Count     int        `json:"count"`
}

type AcceptSubtasksRequest struct {
	Subtasks []string `json:"subtasks" binding:"required"`
}

// ErrorResponse sends a consistent JSON error response
func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error":   true,
		"message": message,
	})
}
