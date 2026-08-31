package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	startTime   time.Time
	environment string
	version     string
}

func NewHealthHandler(env, version string) *HealthHandler {
	return &HealthHandler{
		startTime:   time.Now(),
		environment: env,
		version:     version,
	}
}

type HealthResponse struct {
	Status        string    `json:"status"`
	App           string    `json:"app"`
	Version       string    `json:"version"`
	Environment   string    `json:"environment"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
}

// Check handles the GET /health endpoint
func (h *HealthHandler) Check(c *gin.Context) {
	uptime := time.Since(h.startTime).Seconds()

	c.JSON(http.StatusOK, HealthResponse{
		Status:        "ok",
		App:           "taskflow-backend",
		Version:       h.version,
		Environment:   h.environment,
		UptimeSeconds: uptime,
		Timestamp:     time.Now().UTC(),
	})
}
