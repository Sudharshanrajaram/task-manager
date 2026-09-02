package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitDB initializes a GORM database connection based on configuration
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	isSqlite := cfg.Database.Host == "sqlite" || cfg.Database.Host == ":memory:"
	if isSqlite {
		dialector = sqlite.Open(cfg.Database.DBName)
	} else {
		dialector = postgres.Open(cfg.Database.DSN())
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}
	if cfg.Server.Env == "development" {
		gormConfig.Logger = gormlogger.Default.LogMode(gormlogger.Warn)
	}

	database, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		if !isSqlite && cfg.Server.Env == "development" {
			slog.Warn("Could not connect to PostgreSQL (Docker may not be running). Falling back to local SQLite database 'taskflow_dev.db' for development.",
				slog.String("error", err.Error()),
			)
			isSqlite = true
			database, err = gorm.Open(sqlite.Open("taskflow_dev.db"), gormConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize fallback sqlite database: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Configure connection pool
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic database object: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations
	if err := AutoMigrate(database, !isSqlite); err != nil {
		return nil, fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	dbType := "PostgreSQL"
	if isSqlite {
		dbType = "SQLite"
	}
	slog.Info("Database connected and migrations completed successfully",
		slog.String("engine", dbType),
		slog.String("host", cfg.Database.Host),
	)

	return database, nil
}

// InitTestDB creates an in-memory SQLite database instance for unit/integration tests
func InitTestDB() (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=busy_timeout(5000)"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	if err := AutoMigrate(database, false); err != nil {
		return nil, err
	}

	return database, nil
}

// AutoMigrate applies schema updates to the database
func AutoMigrate(database *gorm.DB, isPostgres bool) error {
	if isPostgres {
		// Attempt to enable pgvector extension if PostgreSQL is used
		_ = database.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
	}

	if err := database.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.Task{},
		&model.Subtask{},
		&model.TimeEntry{},
		&model.TaskTitleEmbedding{},
		&model.Note{},
		&model.TaskDependency{},
		&model.Comment{},
	); err != nil {
		return err
	}

	// Data migration: Retrofit any legacy 'blocked' status tasks to 'in_progress' with is_blocked = true
	_ = database.Exec("UPDATE tasks SET status = 'in_progress', is_blocked = true, blocked_reason = COALESCE(blocked_reason, 'Migrated from blocked status') WHERE status = 'blocked'").Error

	return nil
}
