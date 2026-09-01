package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/taskflow/backend/internal/service"
	"gorm.io/gorm"
)

type HealthHandler struct {
	startTime    time.Time
	environment  string
	version      string
	db           *gorm.DB
	redisClient  *redis.Client
	timerManager service.TimerManager
}

func NewHealthHandler(env, version string, db *gorm.DB, redisClient *redis.Client, tm service.TimerManager) *HealthHandler {
	return &HealthHandler{
		startTime:    time.Now(),
		environment:  env,
		version:      version,
		db:           db,
		redisClient:  redisClient,
		timerManager: tm,
	}
}

type HealthResponse struct {
	Status        string            `json:"status"`
	App           string            `json:"app"`
	Version       string            `json:"version"`
	Environment   string            `json:"environment"`
	UptimeSeconds float64           `json:"uptime_seconds"`
	Timestamp     time.Time         `json:"timestamp"`
	Components    map[string]string `json:"components"`
	ActiveTimers  int               `json:"active_timers"`
}

// Check handles the GET /health endpoint with deep component inspection
func (h *HealthHandler) Check(c *gin.Context) {
	uptime := time.Since(h.startTime).Seconds()
	overallStatus := "ok"

	components := make(map[string]string)

	// 1. Inspect Database
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			components["database"] = "disconnected"
			overallStatus = "degraded"
		} else {
			components["database"] = "connected"
		}
	} else {
		components["database"] = "uninitialized"
	}

	// 2. Inspect Redis
	if h.redisClient != nil {
		if err := h.redisClient.Ping(c.Request.Context()).Err(); err != nil {
			components["redis"] = "in-memory-fallback"
		} else {
			components["redis"] = "connected"
		}
	} else {
		components["redis"] = "in-memory-fallback"
	}

	// 3. Inspect Timers
	activeCount := 0
	if h.timerManager != nil {
		activeCount = len(h.timerManager.GetActiveTimers())
	}

	statusCode := http.StatusOK
	if overallStatus != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:        overallStatus,
		App:           "taskflow-backend",
		Version:       h.version,
		Environment:   h.environment,
		UptimeSeconds: uptime,
		Timestamp:     time.Now().UTC(),
		Components:    components,
		ActiveTimers:  activeCount,
	})
}
