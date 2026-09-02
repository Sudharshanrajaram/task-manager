package service_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/service"
	"github.com/xuri/excelize/v2"
)

func TestLogService_DailyLogsAndExcelExport(t *testing.T) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("InitTestDB failed: %v", err)
	}

	// Seed project, task, and time entry
	proj := model.Project{
		ID:    uuid.New(),
		Name:  "Core Platform",
		Key:   "PLAT",
		Color: "#4F46E5",
	}
	testDB.Create(&proj)

	task := model.Task{
		ID:           uuid.New(),
		ProjectID:    proj.ID,
		TicketNumber: 1,
		TicketKey:    "PLAT-1",
		Title:        "Setup Database Migrations",
		Status:       model.StatusDone,
		Priority:     model.PriorityP0,
	}
	testDB.Create(&task)

	now := time.Now().UTC()
	start := now.Add(-30 * time.Minute)
	entry := model.TimeEntry{
		ID:              uuid.New(),
		TaskID:          task.ID,
		StartedAt:       start,
		EndedAt:         &now,
		DurationSeconds: 1800,
		IsRunning:       false,
	}
	testDB.Create(&entry)

	logService := service.NewLogService(testDB)

	// 1. Get Daily Logs
	logs, err := logService.GetDailyLogs(nil, nil, nil)
	if err != nil {
		t.Fatalf("GetDailyLogs failed: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("Expected at least 1 log entry")
	}
	if logs[0].TicketKey != "PLAT-1" || logs[0].TotalDurationSeconds != 1800 {
		t.Errorf("Unexpected log item: %+v", logs[0])
	}

	// 2. Generate Excel Export
	excelBytes, err := logService.GenerateExcelExport(nil, nil, nil, "UTC")
	if err != nil {
		t.Fatalf("GenerateExcelExport failed: %v", err)
	}
	if len(excelBytes) == 0 {
		t.Fatal("Expected non-empty excel byte stream")
	}

	// 3. Verify valid Excel file structure with excelize
	reader := bytes.NewReader(excelBytes)
	file, err := excelize.OpenReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse generated Excel workbook: %v", err)
	}
	defer file.Close()

	sheet := "Daily Activity Logs"
	val, err := file.GetCellValue(sheet, "A1")
	if err != nil || val != "Date" {
		t.Errorf("Expected cell A1 to be 'Date', got '%s', err=%v", val, err)
	}

	ticketVal, _ := file.GetCellValue(sheet, "D2")
	if ticketVal != "PLAT-1" {
		t.Errorf("Expected cell D2 to be 'PLAT-1', got '%s'", ticketVal)
	}
}
