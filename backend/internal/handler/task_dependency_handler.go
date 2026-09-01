package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type TaskDependencyHandler struct {
	depService service.TaskDependencyService
}

func NewTaskDependencyHandler(depService service.TaskDependencyService) *TaskDependencyHandler {
	return &TaskDependencyHandler{depService: depService}
}

// AddDependency handles POST /api/tasks/:id/dependencies
func (h *TaskDependencyHandler) AddDependency(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	var req CreateDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: depends_on_task_id is required")
		return
	}

	dep, err := h.depService.AddDependency(taskID, req.DependsOnTaskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSelfDependency),
			errors.Is(err, service.ErrCircularDependency),
			errors.Is(err, service.ErrDependencyExists):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrTaskNotFound):
			RespondWithError(c, http.StatusNotFound, "Task not found")
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to add task dependency")
		}
		return
	}

	c.JSON(http.StatusCreated, dep)
}

// GetDependencies handles GET /api/tasks/:id/dependencies
func (h *TaskDependencyHandler) GetDependencies(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	blockedBy, blocks, err := h.depService.GetDependencies(taskID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch dependencies")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blocked_by": blockedBy,
		"blocks":     blocks,
	})
}

// RemoveDependency handles DELETE /api/tasks/:id/dependencies/:depId
func (h *TaskDependencyHandler) RemoveDependency(c *gin.Context) {
	depIDStr := c.Param("depId")
	depID, err := uuid.Parse(depIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid dependency ID format")
		return
	}

	if err := h.depService.RemoveDependency(depID); err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to remove task dependency")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dependency removed successfully"})
}
