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

type RAGHandler struct {
	ragService  service.RAGService
	taskService service.TaskService
	subtaskRepo repository.SubtaskRepository
}

func NewRAGHandler(
	ragService service.RAGService,
	taskService service.TaskService,
	subtaskRepo repository.SubtaskRepository,
) *RAGHandler {
	return &RAGHandler{
		ragService:  ragService,
		taskService: taskService,
		subtaskRepo: subtaskRepo,
	}
}

// SuggestSubtasks handles POST /api/tasks/suggest-subtasks (freeform title)
func (h *RAGHandler) SuggestSubtasks(c *gin.Context) {
	var req SuggestSubtasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Title == "" {
		RespondWithError(c, http.StatusBadRequest, "Task title is required")
		return
	}

	result, err := h.ragService.SuggestSubtasks(c.Request.Context(), req.ProjectID, req.Title, req.Count)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// SuggestSubtasksForTask handles POST /api/tasks/:id/suggest-subtasks
func (h *RAGHandler) SuggestSubtasksForTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	task, err := h.taskService.GetTaskByID(id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	var req SuggestSubtasksRequest
	_ = c.ShouldBindJSON(&req) // Count is optional

	result, err := h.ragService.SuggestSubtasks(c.Request.Context(), &task.ProjectID, task.Title, req.Count)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// AcceptSubtasks handles POST /api/tasks/:id/accept-subtasks (creates subtasks & trains RAG memory)
func (h *RAGHandler) AcceptSubtasks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	task, err := h.taskService.GetTaskByID(id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	var req AcceptSubtasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: subtasks array is required")
		return
	}

	if len(req.Subtasks) == 0 {
		RespondWithError(c, http.StatusBadRequest, "Subtasks list cannot be empty")
		return
	}

	// 1. Get current max order index
	startOrder, _ := h.subtaskRepo.GetNextOrderIndex(task.ID)

	// 2. Insert subtasks in database
	subtaskModels := make([]model.Subtask, 0, len(req.Subtasks))
	for idx, title := range req.Subtasks {
		if title != "" {
			subtaskModels = append(subtaskModels, model.Subtask{
				TaskID:     task.ID,
				Title:      title,
				IsDone:     false,
				OrderIndex: startOrder + idx,
			})
		}
	}

	if err := h.subtaskRepo.CreateBatch(subtaskModels); err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to save subtasks")
		return
	}

	// 3. Trigger RAG Vector Store ingestion (the feedback loop)
	go func() {
		// Use fresh context for background task embedding
		_ = h.ragService.SaveAcceptedSubtasks(c.Request.Context(), &task.ProjectID, task.Title, req.Subtasks)
	}()

	// 4. Return refreshed task
	refreshedTask, _ := h.taskService.GetTaskByID(task.ID)
	c.JSON(http.StatusCreated, refreshedTask)
}
