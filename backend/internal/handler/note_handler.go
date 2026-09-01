package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/service"
)

type NoteHandler struct {
	noteService service.NoteService
}

func NewNoteHandler(noteService service.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

// GetTaskNote handles GET /api/tasks/:id/notes
func (h *NoteHandler) GetTaskNote(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	note, err := h.noteService.GetTaskNote(taskID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch task note")
		return
	}

	c.JSON(http.StatusOK, note)
}

// SaveTaskNote handles PUT /api/tasks/:id/notes
func (h *NoteHandler) SaveTaskNote(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	var req SaveNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var userID *uuid.UUID
	if uID, exists := middleware.GetUserID(c); exists {
		userID = &uID
	}

	note, err := h.noteService.SaveTaskNote(taskID, userID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			RespondWithError(c, http.StatusNotFound, "Task not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to save task note")
		return
	}

	c.JSON(http.StatusOK, note)
}

// GetScratchpad handles GET /api/notes/scratchpad
func (h *NoteHandler) GetScratchpad(c *gin.Context) {
	var userID *uuid.UUID
	if uID, exists := middleware.GetUserID(c); exists {
		userID = &uID
	}

	note, err := h.noteService.GetGlobalScratchpad(userID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch scratchpad")
		return
	}

	c.JSON(http.StatusOK, note)
}

// SaveScratchpad handles PUT /api/notes/scratchpad
func (h *NoteHandler) SaveScratchpad(c *gin.Context) {
	var req SaveNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var userID *uuid.UUID
	if uID, exists := middleware.GetUserID(c); exists {
		userID = &uID
	}

	note, err := h.noteService.SaveGlobalScratchpad(userID, req.Content)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to save scratchpad")
		return
	}

	c.JSON(http.StatusOK, note)
}
