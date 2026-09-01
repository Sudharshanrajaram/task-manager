package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type TimerHandler struct {
	manager service.TimerManager
}

func NewTimerHandler(mgr service.TimerManager) *TimerHandler {
	return &TimerHandler{manager: mgr}
}

// Start handles POST /api/timers/start
func (h *TimerHandler) Start(c *gin.Context) {
	var req StartTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: task_id is required")
		return
	}

	entry, err := h.manager.StartTimer(req.TaskID, req.SubtaskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound), errors.Is(err, service.ErrSubtaskNotFound):
			RespondWithError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrTimerAlreadyRunning):
			RespondWithError(c, http.StatusConflict, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to start timer")
		}
		return
	}

	timerInfo, _ := h.manager.GetTimerByEntryID(entry.ID)
	if timerInfo != nil {
		c.JSON(http.StatusCreated, timerInfo)
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// Pause handles POST /api/timers/:id/pause
func (h *TimerHandler) Pause(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid timer ID format (must be UUID)")
		return
	}

	_, err = h.manager.PauseTimer(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTimerNotFound):
			RespondWithError(c, http.StatusNotFound, "Timer not found or not active")
		case errors.Is(err, service.ErrTimerAlreadyPaused):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to pause timer")
		}
		return
	}

	timerInfo, _ := h.manager.GetTimerByEntryID(id)
	c.JSON(http.StatusOK, timerInfo)
}

// Resume handles POST /api/timers/:id/resume
func (h *TimerHandler) Resume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid timer ID format (must be UUID)")
		return
	}

	_, err = h.manager.ResumeTimer(id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTimerNotFound):
			RespondWithError(c, http.StatusNotFound, "Timer not found or not active")
		case errors.Is(err, service.ErrTimerNotPaused):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to resume timer")
		}
		return
	}

	timerInfo, _ := h.manager.GetTimerByEntryID(id)
	c.JSON(http.StatusOK, timerInfo)
}

// Stop handles POST /api/timers/:id/stop
func (h *TimerHandler) Stop(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid timer ID format (must be UUID)")
		return
	}

	entry, err := h.manager.StopTimer(id)
	if err != nil {
		if errors.Is(err, service.ErrTimerNotFound) {
			RespondWithError(c, http.StatusNotFound, "Timer not found or not active")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to stop timer")
		return
	}

	c.JSON(http.StatusOK, entry)
}

// Adjust handles POST /api/timers/:id/adjust
func (h *TimerHandler) Adjust(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid timer ID format (must be UUID)")
		return
	}

	var req AdjustTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: delta_seconds is required")
		return
	}

	entry, err := h.manager.AdjustTime(id, req.DeltaSeconds)
	if err != nil {
		if errors.Is(err, service.ErrTimerNotFound) {
			RespondWithError(c, http.StatusNotFound, "Time entry not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to adjust timer")
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetActive handles GET /api/timers/active
func (h *TimerHandler) GetActive(c *gin.Context) {
	timers := h.manager.GetActiveTimers()
	c.JSON(http.StatusOK, timers)
}

// AnalyticsSummary handles GET /api/analytics/summary
func (h *TimerHandler) AnalyticsSummary(c *gin.Context) {
	rangeType := c.DefaultQuery("range", "week")
	summary, err := h.manager.GetAnalyticsSummary(rangeType)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to calculate analytics summary")
		return
	}

	c.JSON(http.StatusOK, summary)
}
