package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// Create handles POST /api/tasks/:id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Content is required")
		return
	}

	comment, err := h.commentService.CreateComment(taskID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			RespondWithError(c, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrCommentContentRequired):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to create comment")
		}
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// ListByTask handles GET /api/tasks/:id/comments
func (h *CommentHandler) ListByTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format (must be UUID)")
		return
	}

	comments, err := h.commentService.GetCommentsByTaskID(taskID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch comments")
		return
	}

	c.JSON(http.StatusOK, comments)
}

// Delete handles DELETE /api/tasks/:id/comments/:commentId
func (h *CommentHandler) Delete(c *gin.Context) {
	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid comment ID format (must be UUID)")
		return
	}

	if err := h.commentService.DeleteComment(commentID); err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			RespondWithError(c, http.StatusNotFound, "Comment not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	c.Status(http.StatusNoContent)
}
