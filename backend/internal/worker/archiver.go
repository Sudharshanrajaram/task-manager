package worker

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// StartAutoArchiver runs a periodic background worker archiving Done tasks older than 14 days
func StartAutoArchiver(ctx context.Context, db *gorm.DB, interval time.Duration) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)
	go func() {
		// Run once immediately on start
		archiveOlderTasks(db)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				archiveOlderTasks(db)
			}
		}
	}()
}

func archiveOlderTasks(db *gorm.DB) {
	threshold := time.Now().UTC().AddDate(0, 0, -14)
	now := time.Now().UTC()

	result := db.Exec(
		"UPDATE tasks SET is_archived = true, archived_at = ? WHERE status = 'done' AND is_archived = false AND updated_at < ?",
		now, threshold,
	)

	if result.Error != nil {
		slog.Warn("Failed to execute auto-archive background worker", slog.String("error", result.Error.Error()))
		return
	}

	if result.RowsAffected > 0 {
		slog.Info("Auto-archived historical completed tasks", slog.Int64("archived_count", result.RowsAffected))
	}
}
