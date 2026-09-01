package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
)

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
	Type             model.TaskType    `json:"type"`
	Title            string            `json:"title" binding:"required"`
	Description      string            `json:"description"`
	Status           model.TaskStatus  `json:"status"`
	Priority         model.TaskPriority`json:"priority"`
	Labels           []string          `json:"labels"`
	StepsToReproduce *string           `json:"steps_to_reproduce"`
	Severity         *model.BugSeverity`json:"severity"`
	Environment      *string           `json:"environment"`
	InitialSubtasks  []string          `json:"initial_subtasks"`
}

type UpdateTaskRequest struct {
	Title            *string            `json:"title"`
	Description      *string            `json:"description"`
	Type             *model.TaskType    `json:"type"`
	Status           *model.TaskStatus  `json:"status"`
	Priority         *model.TaskPriority`json:"priority"`
	Labels           *[]string          `json:"labels"`
	StepsToReproduce *string            `json:"steps_to_reproduce"`
	Severity         *model.BugSeverity `json:"severity"`
	Environment      *string            `json:"environment"`
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
