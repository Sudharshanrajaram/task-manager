package service

import (
	"bytes"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type DailyLogItem struct {
	Date                 string  `json:"date"`
	FirstStartedAt       string  `json:"first_started_at,omitempty"`
	ProjectID            string  `json:"project_id"`
	ProjectName          string  `json:"project_name"`
	ProjectKey           string  `json:"project_key"`
	ProjectColor         string  `json:"project_color"`
	TaskID               string  `json:"task_id"`
	TicketKey            string  `json:"ticket_key"`
	TaskTitle            string  `json:"task_title"`
	AISummary            *string `json:"ai_summary,omitempty"`
	TotalDurationSeconds int64   `json:"total_duration_seconds"`
	DurationFormatted    string  `json:"duration_formatted"`
	Status               string  `json:"status"`
	LatestEntryID        *string `json:"latest_entry_id,omitempty"`
}

type LogService interface {
	GetDailyLogs(from, to *time.Time, projectID *uuid.UUID) ([]DailyLogItem, error)
	GenerateExcelExport(from, to *time.Time, projectID *uuid.UUID, tz string) ([]byte, error)
	TriggerAutoArchive() (int64, error)
}

type logService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) LogService {
	return &logService{db: db}
}

func (s *logService) GetDailyLogs(from, to *time.Time, projectID *uuid.UUID) ([]DailyLogItem, error) {
	// Group at task level per day
	query := s.db.Table("time_entries").
		Select(`
			DATE(time_entries.started_at) as date,
			MIN(time_entries.started_at) as first_started_at,
			MAX(time_entries.id) as latest_entry_id,
			projects.id as project_id,
			projects.name as project_name,
			projects.key as project_key,
			projects.color as project_color,
			tasks.id as task_id,
			tasks.ticket_key as ticket_key,
			tasks.title as task_title,
			tasks.ai_summary as ai_summary,
			SUM(time_entries.duration_seconds) as total_duration_seconds,
			tasks.status as status
		`).
		Joins("LEFT JOIN tasks ON tasks.id = time_entries.task_id").
		Joins("LEFT JOIN projects ON projects.id = tasks.project_id").
		Where("time_entries.is_running = ?", false)

	if from != nil {
		query = query.Where("time_entries.started_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("time_entries.started_at <= ?", *to)
	}
	if projectID != nil && *projectID != uuid.Nil {
		query = query.Where("tasks.project_id = ?", *projectID)
	}

	query = query.Group("date, projects.id, projects.name, projects.key, projects.color, tasks.id, tasks.ticket_key, tasks.title, tasks.ai_summary, tasks.status").
		Order("date DESC, ticket_key ASC")

	var items []DailyLogItem
	if err := query.Scan(&items).Error; err != nil {
		return nil, err
	}

	for i := range items {
		h := items[i].TotalDurationSeconds / 3600
		m := (items[i].TotalDurationSeconds % 3600) / 60
		items[i].DurationFormatted = fmt.Sprintf("%dh %02dm", h, m)
	}

	return items, nil
}

func (s *logService) GenerateExcelExport(from, to *time.Time, projectID *uuid.UUID, tz string) ([]byte, error) {
	items, err := s.GetDailyLogs(from, to, projectID)
	if err != nil {
		return nil, err
	}

	// Resolve local timezone
	loc := time.Local
	if tz != "" {
		if parsedLoc, err := time.LoadLocation(tz); err == nil {
			loc = parsedLoc
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Daily Activity Logs"
	f.SetSheetName("Sheet1", sheet)

	// Set title header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	// Phase 18: No subtasks, no seconds, include AI Summary column
	headers := []string{"Date", "Time (Local)", "Project", "Ticket", "Task Title", "Duration", "AI Summary", "Status"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for rowIdx, item := range items {
		row := rowIdx + 2

		// Duration formatted as Xh Ym (zero seconds per Phase 18)
		h := item.TotalDurationSeconds / 3600
		m := (item.TotalDurationSeconds % 3600) / 60
		durationClean := fmt.Sprintf("%dh %02dm", h, m)

		// Parse time and convert to local location without trailing Z
		localTimeStr := ""
		if item.FirstStartedAt != "" {
			for _, layout := range []string{"2006-01-02 15:04:05.999999999-07:00", time.RFC3339, "2006-01-02 15:04:05"} {
				if t, err := time.Parse(layout, item.FirstStartedAt); err == nil {
					localTimeStr = t.In(loc).Format("15:04")
					break
				}
			}
		}

		summaryText := "—"
		if item.AISummary != nil && *item.AISummary != "" {
			summaryText = *item.AISummary
		}

		projectLabel := fmt.Sprintf("[%s] %s", item.ProjectKey, item.ProjectName)

		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.Date)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), localTimeStr)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), projectLabel)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.TicketKey)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.TaskTitle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), durationClean)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), summaryText)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.Status)
	}

	// Auto fit column widths
	_ = f.SetColWidth(sheet, "A", "A", 14) // Date
	_ = f.SetColWidth(sheet, "B", "B", 14) // Time
	_ = f.SetColWidth(sheet, "C", "C", 22) // Project
	_ = f.SetColWidth(sheet, "D", "D", 12) // Ticket
	_ = f.SetColWidth(sheet, "E", "E", 34) // Task Title
	_ = f.SetColWidth(sheet, "F", "F", 14) // Duration
	_ = f.SetColWidth(sheet, "G", "G", 45) // AI Summary
	_ = f.SetColWidth(sheet, "H", "H", 14) // Status

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *logService) TriggerAutoArchive() (int64, error) {
	threshold := time.Now().UTC().AddDate(0, 0, -14)
	now := time.Now().UTC()

	result := s.db.Exec(
		"UPDATE tasks SET is_archived = true, archived_at = ? WHERE status = 'done' AND is_archived = false AND updated_at < ?",
		now, threshold,
	)
	return result.RowsAffected, result.Error
}
