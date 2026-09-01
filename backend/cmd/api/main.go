package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
	"github.com/taskflow/backend/pkg/groq"
	"github.com/taskflow/backend/pkg/logger"
)

const AppVersion = "1.0.0"

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize structured logger
	logger.InitLogger(cfg.Server.Env)
	slog.Info("Starting TaskFlow API server",
		slog.String("version", AppVersion),
		slog.String("environment", cfg.Server.Env),
		slog.String("port", cfg.Server.Port),
	)

	// 3. Initialize Database & run AutoMigrations
	database, err := db.InitDB(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Initialize Repositories
	projectRepo := repository.NewProjectRepository(database)
	taskRepo := repository.NewTaskRepository(database)
	subtaskRepo := repository.NewSubtaskRepository(database)
	timeEntryRepo := repository.NewTimeEntryRepository(database)
	embeddingRepo := repository.NewEmbeddingRepository(database)

	// 5. Initialize External Clients
	groqClient := groq.NewClient(cfg.Groq.APIKey, cfg.Groq.ChatModel, cfg.Groq.EmbeddingModel)

	// 6. Initialize Services
	projectService := service.NewProjectService(projectRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo, subtaskRepo)
	subtaskService := service.NewSubtaskService(subtaskRepo, taskRepo)
	timerManager := service.NewTimerManager(timeEntryRepo, taskRepo, subtaskRepo)
	ragService := service.NewRAGService(groqClient, embeddingRepo)

	// Recover in-flight active timers from DB on boot
	if err := timerManager.RecoverInFlightTimers(); err != nil {
		slog.Warn("Failed to recover in-flight timers", slog.String("error", err.Error()))
	}

	// 7. Initialize Handlers
	healthHandler := handler.NewHealthHandler(cfg.Server.Env, AppVersion)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	subtaskHandler := handler.NewSubtaskHandler(subtaskService)
	timerHandler := handler.NewTimerHandler(timerManager)
	ragHandler := handler.NewRAGHandler(ragService, taskService, subtaskRepo)

	// 8. Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 9. Setup Gin engine & middlewares
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	// 10. Register routes
	router.GET("/health", healthHandler.Check)

	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// Projects
		api.POST("/projects", projectHandler.Create)
		api.GET("/projects", projectHandler.List)
		api.GET("/projects/:id", projectHandler.GetByID)
		api.PATCH("/projects/:id", projectHandler.Update)
		api.DELETE("/projects/:id", projectHandler.Delete)

		// Tasks under Project
		api.POST("/projects/:id/tasks", taskHandler.Create)
		api.GET("/projects/:id/tasks", taskHandler.ListByProject)

		// Tasks
		api.GET("/tasks/:id", taskHandler.GetByIDOrKey)
		api.PATCH("/tasks/:id", taskHandler.Update)
		api.DELETE("/tasks/:id", taskHandler.Delete)

		// Subtasks
		api.POST("/tasks/:id/subtasks", subtaskHandler.Create)
		api.PUT("/tasks/:id/subtasks/reorder", subtaskHandler.Reorder)
		api.PATCH("/subtasks/:id", subtaskHandler.Update)
		api.DELETE("/subtasks/:id", subtaskHandler.Delete)

		// RAG AI Subtask Suggestions
		api.POST("/tasks/suggest-subtasks", ragHandler.SuggestSubtasks)
		api.POST("/tasks/:id/suggest-subtasks", ragHandler.SuggestSubtasksForTask)
		api.POST("/tasks/:id/accept-subtasks", ragHandler.AcceptSubtasks)

		// Timers
		api.POST("/timers/start", timerHandler.Start)
		api.POST("/timers/:id/pause", timerHandler.Pause)
		api.POST("/timers/:id/resume", timerHandler.Resume)
		api.POST("/timers/:id/stop", timerHandler.Stop)
		api.POST("/timers/:id/adjust", timerHandler.Adjust)
		api.GET("/timers/active", timerHandler.GetActive)

		// Analytics
		api.GET("/analytics/summary", timerHandler.AnalyticsSummary)
	}

	// 10. Setup HTTP server with sensible timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 11. Start server in a background goroutine
	go func() {
		slog.Info(fmt.Sprintf("Server listening on http://localhost:%s", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 12. Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down TaskFlow API server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server exited cleanly")
}
