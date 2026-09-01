package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/taskflow/backend/internal/service"
	"github.com/taskflow/backend/pkg/metrics"
	"gorm.io/gorm"
)

type MetricsHandler struct {
	db           *gorm.DB
	redisClient  *redis.Client
	timerManager service.TimerManager
	startTime    time.Time
}

func NewMetricsHandler(db *gorm.DB, redisClient *redis.Client, tm service.TimerManager) *MetricsHandler {
	return &MetricsHandler{
		db:           db,
		redisClient:  redisClient,
		timerManager: tm,
		startTime:    time.Now(),
	}
}

// Metrics returns Prometheus-formatted plain text metrics for scraping
func (h *MetricsHandler) Metrics(c *gin.Context) {
	uptimeSeconds := time.Since(h.startTime).Seconds()

	dbStatus := 0
	if h.db != nil {
		if sqlDB, err := h.db.DB(); err == nil && sqlDB.Ping() == nil {
			dbStatus = 1
		}
	}

	redisStatus := 0
	if h.redisClient != nil {
		if err := h.redisClient.Ping(c.Request.Context()).Err(); err == nil {
			redisStatus = 1
		}
	}

	activeTimersCount := 0
	if h.timerManager != nil {
		activeTimersCount = len(h.timerManager.GetActiveTimers())
	}

	metricsText := fmt.Sprintf(
		"# HELP taskflow_uptime_seconds Total uptime of TaskFlow API in seconds\n"+
			"# TYPE taskflow_uptime_seconds gauge\n"+
			"taskflow_uptime_seconds %.2f\n\n"+
			"# HELP taskflow_http_requests_total Total number of HTTP requests processed\n"+
			"# TYPE taskflow_http_requests_total counter\n"+
			"taskflow_http_requests_total %d\n\n"+
			"# HELP taskflow_database_connected Database connection health (1=up, 0=down)\n"+
			"# TYPE taskflow_database_connected gauge\n"+
			"taskflow_database_connected %d\n\n"+
			"# HELP taskflow_redis_connected Redis connection health (1=up, 0=down)\n"+
			"# TYPE taskflow_redis_connected gauge\n"+
			"taskflow_redis_connected %d\n\n"+
			"# HELP taskflow_active_timers Active running/paused timers count\n"+
			"# TYPE taskflow_active_timers gauge\n"+
			"taskflow_active_timers %d\n",
		uptimeSeconds,
		metrics.TotalHTTPRequests.Load(),
		dbStatus,
		redisStatus,
		activeTimersCount,
	)

	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(metricsText))
}
