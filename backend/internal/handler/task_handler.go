package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

type TaskHandler struct {
	service service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{service: svc}
}

// Create handles POST /api/projects/:id/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid project ID format (must be UUID)")
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: title is required")
		return
	}

	task, err := h.service.CreateTask(service.CreateTaskInput{
		ProjectID:        projectID,
		Type:             req.Type,
		Title:            req.Title,
		Description:      req.Description,
		Status:           req.Status,
		Priority:         req.Priority,
		Labels:           req.Labels,
		StepsToReproduce: req.StepsToReproduce,
		Severity:         req.Severity,
		Environment:      req.Environment,
		InitialSubtasks:  req.InitialSubtasks,
	})

	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNotFound):
			RespondWithError(c, http.StatusNotFound, "Project not found")
		case errors.Is(err, service.ErrTaskTitleRequired),
			errors.Is(err, service.ErrInvalidTaskType),
			errors.Is(err, service.ErrInvalidTaskStatus),
			errors.Is(err, service.ErrInvalidTaskPriority),
			errors.Is(err, service.ErrInvalidBugSeverity):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to create task")
		}
		return
	}

	c.JSON(http.StatusCreated, task)
}

// ListByProject handles GET /api/projects/:id/tasks
func (h *TaskHandler) ListByProject(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid project ID format (must be UUID)")
		return
	}

	filter := repository.TaskFilter{
		ProjectID: &projectID,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := model.TaskStatus(statusStr)
		filter.Status = &status
	}
	if typeStr := c.Query("type"); typeStr != "" {
		taskType := model.TaskType(typeStr)
		filter.Type = &taskType
	}
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := model.TaskPriority(priorityStr)
		filter.Priority = &priority
	}
	if searchStr := c.Query("search"); searchStr != "" {
		filter.Search = &searchStr
	}

	tasks, err := h.service.GetTasks(filter)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetByIDOrKey handles GET /api/tasks/:id (supports both UUID and ticket key e.g. "AUTH-42")
func (h *TaskHandler) GetByIDOrKey(c *gin.Context) {
	idOrKey := c.Param("id")

	var task *model.Task
	var err error

	if id, parseErr := uuid.Parse(idOrKey); parseErr == nil {
		task, err = h.service.GetTaskByID(id)
	} else {
		task, err = h.service.GetTaskByTicketKey(idOrKey)
	}

	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch task")
		return
	}

	c.JSON(http.StatusOK, task)
}

// Update handles PATCH /api/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	task, err := h.service.UpdateTask(id, service.UpdateTaskInput{
		Title:            req.Title,
		Description:      req.Description,
		Type:             req.Type,
		Status:           req.Status,
		Priority:         req.Priority,
		Labels:           req.Labels,
		StepsToReproduce: req.StepsToReproduce,
		Severity:         req.Severity,
		Environment:      req.Environment,
	})

	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			RespondWithError(c, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrTaskTitleRequired),
			errors.Is(err, service.ErrInvalidTaskType),
			errors.Is(err, service.ErrInvalidTaskStatus),
			errors.Is(err, service.ErrInvalidTaskPriority),
			errors.Is(err, service.ErrInvalidBugSeverity):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to update task")
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// Delete handles DELETE /api/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	if err := h.service.DeleteTask(id); err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
