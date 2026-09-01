package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type LogHandler struct {
	logService service.LogService
}

func NewLogHandler(logService service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// GetDailyLogs handles GET /api/logs/daily?from=&to=&project_id=
func (h *LogHandler) GetDailyLogs(c *gin.Context) {
	from, to, projectID, err := parseLogFilters(c)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.logService.GetDailyLogs(from, to, projectID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve daily logs")
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": items})
}

// ExportExcel handles GET /api/logs/export?from=&to=&format=xlsx
func (h *LogHandler) ExportExcel(c *gin.Context) {
	from, to, projectID, err := parseLogFilters(c)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	excelBytes, err := h.logService.GenerateExcelExport(from, to, projectID)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to generate Excel export")
		return
	}

	fileName := fmt.Sprintf("taskflow-activity-%s.xlsx", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelBytes)
}

// TriggerArchive handles POST /api/logs/archive-trigger
func (h *LogHandler) TriggerArchive(c *gin.Context) {
	count, err := h.logService.TriggerAutoArchive()
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to execute auto-archive")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Auto-archive execution completed",
		"archived_count": count,
	})
}

func parseLogFilters(c *gin.Context) (*time.Time, *time.Time, *uuid.UUID, error) {
	var from *time.Time
	var to *time.Time
	var projectID *uuid.UUID

	if fromStr := c.Query("from"); fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid 'from' date format (expected YYYY-MM-DD)")
		}
		from = &t
	}

	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid 'to' date format (expected YYYY-MM-DD)")
		}
		// End of the day
		t = t.Add(24*time.Hour - time.Nanosecond)
		to = &t
	}

	if pIDStr := c.Query("project_id"); pIDStr != "" {
		pID, err := uuid.Parse(pIDStr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid 'project_id' UUID format")
		}
		projectID = &pID
	}

	return from, to, projectID, nil
}
