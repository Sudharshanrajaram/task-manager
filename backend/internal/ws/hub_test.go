package ws_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
	"github.com/taskflow/backend/internal/ws"
)

func TestWebSocketHub_ConnectAndBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("InitTestDB failed: %v", err)
	}

	taskRepo := repository.NewTaskRepository(testDB)
	subtaskRepo := repository.NewSubtaskRepository(testDB)
	timeEntryRepo := repository.NewTimeEntryRepository(testDB)
	timerManager := service.NewTimerManager(timeEntryRepo, taskRepo, subtaskRepo)

	// Create Hub with in-memory event bridge
	hub := ws.NewHub(timerManager, nil, []string{"*"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	r := gin.New()
	r.GET("/ws/timers", hub.HandleWebSocket)

	server := httptest.NewServer(r)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/timers"

	// Connect WebSocket client
	wsConn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v (resp: %v)", err, resp)
	}
	defer wsConn.Close()

	// Give registration a moment
	time.Sleep(50 * time.Millisecond)

	if hub.ActiveClientCount() != 1 {
		t.Errorf("Expected 1 active client, got %d", hub.ActiveClientCount())
	}

	// Broadcast test event
	testPayload := map[string]string{
		"type":    "timer.started",
		"task_id": "test-task-123",
	}
	if err := hub.BroadcastEvent(testPayload); err != nil {
		t.Fatalf("BroadcastEvent failed: %v", err)
	}

	// Read message from WebSocket client
	_ = wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read WebSocket message: %v", err)
	}

	var received map[string]string
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal received WebSocket message: %v", err)
	}

	if received["type"] != "timer.started" || received["task_id"] != "test-task-123" {
		t.Errorf("Unexpected payload received: %v", received)
	}
}
