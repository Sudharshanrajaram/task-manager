package service_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

func TestNoteService_TaskNotesAndScratchpad(t *testing.T) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("InitTestDB failed: %v", err)
	}

	noteRepo := repository.NewNoteRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)
	noteService := service.NewNoteService(noteRepo, taskRepo)

	// Seed project and task
	proj := model.Project{ID: uuid.New(), Name: "Notes App", Key: "NOTE"}
	testDB.Create(&proj)

	task := model.Task{
		ID:           uuid.New(),
		ProjectID:    proj.ID,
		TicketNumber: 1,
		TicketKey:    "NOTE-1",
		Title:        "Implement Markdown Notes",
		Status:       model.StatusInProgress,
		Priority:     model.PriorityP1,
	}
	testDB.Create(&task)

	// 1. Initial get (empty note)
	note, err := noteService.GetTaskNote(task.ID)
	if err != nil {
		t.Fatalf("GetTaskNote failed: %v", err)
	}
	if note.Content != "" {
		t.Errorf("Expected empty content, got %s", note.Content)
	}

	// 2. Save note
	saved, err := noteService.SaveTaskNote(task.ID, nil, "# Design Thoughts\n- Use react-markdown\n- Support [[backlinks]]")
	if err != nil {
		t.Fatalf("SaveTaskNote failed: %v", err)
	}
	if saved.Content == "" {
		t.Fatal("Expected non-empty saved content")
	}

	// 3. Retrieve saved note
	fetched, err := noteService.GetTaskNote(task.ID)
	if err != nil || fetched.Content != saved.Content {
		t.Errorf("Mismatch in fetched note content: %v", err)
	}

	// 4. Test global scratchpad
	scratchpad, err := noteService.GetGlobalScratchpad(nil)
	if err != nil {
		t.Fatalf("GetGlobalScratchpad failed: %v", err)
	}
	if scratchpad.Content != "" {
		t.Errorf("Expected empty initial scratchpad")
	}

	savedScratch, err := noteService.SaveGlobalScratchpad(nil, "Quick scratchpad note: remember to test WebSockets")
	if err != nil {
		t.Fatalf("SaveGlobalScratchpad failed: %v", err)
	}
	if savedScratch.Content == "" {
		t.Fatal("Expected non-empty scratchpad")
	}
}
