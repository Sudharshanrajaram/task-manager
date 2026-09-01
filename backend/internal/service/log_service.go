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
	Date                 string `json:"date"`
	ProjectID            string `json:"project_id"`
	ProjectName          string `json:"project_name"`
	ProjectKey           string `json:"project_key"`
	ProjectColor         string `json:"project_color"`
	TaskID               string `json:"task_id"`
	TicketKey            string `json:"ticket_key"`
	TaskTitle            string `json:"task_title"`
	SubtaskTitle         string `json:"subtask_title"`
	TotalDurationSeconds int64  `json:"total_duration_seconds"`
	Status               string `json:"status"`
}

type LogService interface {
	GetDailyLogs(from, to *time.Time, projectID *uuid.UUID) ([]DailyLogItem, error)
	GenerateExcelExport(from, to *time.Time, projectID *uuid.UUID) ([]byte, error)
	TriggerAutoArchive() (int64, error)
}

type logService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) LogService {
	return &logService{db: db}
}

func (s *logService) GetDailyLogs(from, to *time.Time, projectID *uuid.UUID) ([]DailyLogItem, error) {
	// Build query on time_entries joined with tasks, subtasks, projects
	query := s.db.Table("time_entries").
		Select(`
			DATE(time_entries.started_at) as date,
			projects.id as project_id,
			projects.name as project_name,
			projects.key as project_key,
			projects.color as project_color,
			tasks.id as task_id,
			tasks.ticket_key as ticket_key,
			tasks.title as task_title,
			COALESCE(subtasks.title, '') as subtask_title,
			SUM(time_entries.duration_seconds) as total_duration_seconds,
			tasks.status as status
		`).
		Joins("LEFT JOIN tasks ON tasks.id = time_entries.task_id").
		Joins("LEFT JOIN projects ON projects.id = tasks.project_id").
		Joins("LEFT JOIN subtasks ON subtasks.id = time_entries.subtask_id").
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

	query = query.Group("date, projects.id, projects.name, projects.key, projects.color, tasks.id, tasks.ticket_key, tasks.title, subtasks.title, tasks.status").
		Order("date DESC, ticket_key ASC")

	var items []DailyLogItem
	if err := query.Scan(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *logService) GenerateExcelExport(from, to *time.Time, projectID *uuid.UUID) ([]byte, error) {
	items, err := s.GetDailyLogs(from, to, projectID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Daily Activity Logs"
	f.SetSheetName("Sheet1", sheet)

	// Set title header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	headers := []string{"Date", "Project Key", "Project Name", "Ticket", "Task Title", "Subtask Title", "Duration (HH:MM:SS)", "Seconds", "Status"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for rowIdx, item := range items {
		row := rowIdx + 2

		h := item.TotalDurationSeconds / 3600
		m := (item.TotalDurationSeconds % 3600) / 60
		sec := item.TotalDurationSeconds % 60
		hms := fmt.Sprintf("%02d:%02d:%02d", h, m, sec)

		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.Date)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.ProjectKey)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.ProjectName)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), item.TicketKey)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), item.TaskTitle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), item.SubtaskTitle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), hms)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), item.TotalDurationSeconds)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), item.Status)
	}

	// Auto fit column widths
	_ = f.SetColWidth(sheet, "A", "A", 12)
	_ = f.SetColWidth(sheet, "B", "B", 14)
	_ = f.SetColWidth(sheet, "C", "C", 20)
	_ = f.SetColWidth(sheet, "D", "D", 12)
	_ = f.SetColWidth(sheet, "E", "E", 35)
	_ = f.SetColWidth(sheet, "F", "F", 25)
	_ = f.SetColWidth(sheet, "G", "G", 18)
	_ = f.SetColWidth(sheet, "H", "H", 12)
	_ = f.SetColWidth(sheet, "I", "I", 14)

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
