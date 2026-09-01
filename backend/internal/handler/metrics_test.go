package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/taskflow/backend/internal/handler"
)

func TestMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	metricsH := handler.NewMetricsHandler(nil, nil, nil)
	r.GET("/metrics", metricsH.Metrics)

	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "taskflow_uptime_seconds")
	assert.Contains(t, body, "taskflow_http_requests_total")
	assert.Contains(t, body, "taskflow_database_connected")
}

func TestDeepHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	healthH := handler.NewHealthHandler("test", "1.0.0", nil, nil, nil)
	r.GET("/health", healthH.Check)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "taskflow-backend")
	assert.Contains(t, body, "components")
}
