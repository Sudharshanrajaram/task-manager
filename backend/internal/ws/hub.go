package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/taskflow/backend/internal/service"
)

const RedisTimerChannel = "taskflow:timers"

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages to broadcast to clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Lock for client access
	mu sync.RWMutex

	// TimerManager subscription
	timerManager service.TimerManager

	// Redis client for distributed pub/sub
	redisClient *redis.Client

	// WebSocket upgrader
	upgrader websocket.Upgrader
}

func NewHub(timerManager service.TimerManager, redisClient *redis.Client, allowedOrigins []string) *Hub {
	originMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		originMap[strings.TrimSpace(o)] = true
	}

	h := &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		timerManager: timerManager,
		redisClient:  redisClient,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // Allow non-browser requests
				}
				if originMap["*"] || originMap[origin] {
					return true
				}
				// Allow local dev origins by default
				if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
					return true
				}
				return false
			},
		},
	}

	return h
}

// Run executes the hub event loop
func (h *Hub) Run(ctx context.Context) {
	// 1. If Redis is available, listen to Redis Pub/Sub channel
	if h.redisClient != nil {
		go h.listenRedisPubSub(ctx)
	}

	// 2. Always bridge local TimerManager events
	// If Redis is running, publish local events to Redis.
	// If Redis is not running, broadcast directly to local WS clients.
	timerEvents, unsubscribe := h.timerManager.Subscribe()
	go func() {
		defer unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-timerEvents:
				if !ok {
					return
				}
				eventBytes, err := json.Marshal(event)
				if err != nil {
					continue
				}

				if h.redisClient != nil {
					// Publish to Redis for multi-instance distribution
					_ = h.redisClient.Publish(ctx, RedisTimerChannel, eventBytes).Err()
				} else {
					// In-memory direct broadcast
					h.broadcast <- eventBytes
				}
			}
		}
	}()

	// 3. Main client connection event loop
	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			slog.Debug("WebSocket client registered", slog.Int("active_clients", clientCount))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			clientCount := len(h.clients)
			h.mu.Unlock()
			slog.Debug("WebSocket client unregistered", slog.Int("active_clients", clientCount))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow client, drop and close
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// listenRedisPubSub subscribes to Redis channel and feeds the broadcast channel
func (h *Hub) listenRedisPubSub(ctx context.Context) {
	pubsub := h.redisClient.Subscribe(ctx, RedisTimerChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.broadcast <- []byte(msg.Payload)
		}
	}
}

// HandleWebSocket upgrades HTTP to WebSocket and registers the client
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Warn("Failed to upgrade WebSocket connection", slog.String("error", err.Error()))
		return
	}

	client := NewClient(h, conn)
	h.register <- client

	// Start read and write pumps in separate goroutines
	go client.WritePump()
	go client.ReadPump()
}

// BroadcastEvent allows programmatic event dispatching
func (h *Hub) BroadcastEvent(event interface{}) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	h.broadcast <- bytes
	return nil
}

// ActiveClientCount returns the number of connected clients
func (h *Hub) ActiveClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
