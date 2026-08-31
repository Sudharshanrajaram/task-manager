package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type SubtaskHandler struct {
	service service.SubtaskService
}

func NewSubtaskHandler(svc service.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{service: svc}
}

// Create handles POST /api/tasks/:id/subtasks
func (h *SubtaskHandler) Create(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	var req CreateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: title is required")
		return
	}

	subtask, err := h.service.CreateSubtask(taskID, req.Title)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			RespondWithError(c, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrSubtaskTitleRequired):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to create subtask")
		}
		return
	}

	c.JSON(http.StatusCreated, subtask)
}

// Update handles PATCH /api/subtasks/:id
func (h *SubtaskHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid subtask ID format (must be UUID)")
		return
	}

	var req UpdateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	subtask, err := h.service.UpdateSubtask(id, req.Title, req.IsDone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubtaskNotFound):
			RespondWithError(c, http.StatusNotFound, "Subtask not found")
		case errors.Is(err, service.ErrSubtaskTitleRequired):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to update subtask")
		}
		return
	}

	c.JSON(http.StatusOK, subtask)
}

// Delete handles DELETE /api/subtasks/:id
func (h *SubtaskHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid subtask ID format (must be UUID)")
		return
	}

	if err := h.service.DeleteSubtask(id); err != nil {
		if errors.Is(err, service.ErrSubtaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Subtask not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete subtask")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subtask deleted successfully"})
}

// Reorder handles PUT /api/tasks/:id/subtasks/reorder
func (h *SubtaskHandler) Reorder(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	var req ReorderSubtasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: ordered_ids is required")
		return
	}

	if err := h.service.ReorderSubtasks(taskID, req.OrderedIDs); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to reorder subtasks")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subtasks reordered successfully"})
}
