package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&model.Project{},
		&model.Task{},
		&model.Subtask{},
		&model.TimeEntry{},
		&model.Comment{},
	)
	assert.NoError(t, err)

	return db
}

func TestCommentService_Flow(t *testing.T) {
	db := setupTestDB(t)

	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo, taskRepo)

	proj := &model.Project{Name: "Test Project", Key: "TEST"}
	err := projectRepo.Create(proj)
	assert.NoError(t, err)

	task := &model.Task{
		ProjectID: proj.ID,
		TicketKey: "TEST-1",
		Title:     "Initial Task",
		Status:    model.StatusBacklog,
		Priority:  model.PriorityP1,
		Type:      model.TypeTask,
	}
	err = taskRepo.Create(nil, task)
	assert.NoError(t, err)

	// Test 1: Empty content validation
	_, err = commentService.CreateComment(task.ID, "   ")
	assert.ErrorIs(t, err, service.ErrCommentContentRequired)

	// Test 2: Successful comment creation
	c1, err := commentService.CreateComment(task.ID, "First comment on task")
	assert.NoError(t, err)
	assert.Equal(t, "First comment on task", c1.Content)
	assert.Equal(t, task.ID, c1.TaskID)

	c2, err := commentService.CreateComment(task.ID, "Second comment with **markdown**")
	assert.NoError(t, err)

	// Test 3: List comments chronologically
	comments, err := commentService.GetCommentsByTaskID(task.ID)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "First comment on task", comments[0].Content)
	assert.Equal(t, "Second comment with **markdown**", comments[1].Content)

	// Test 4: Delete comment
	err = commentService.DeleteComment(c1.ID)
	assert.NoError(t, err)

	commentsAfter, err := commentService.GetCommentsByTaskID(task.ID)
	assert.NoError(t, err)
	assert.Len(t, commentsAfter, 1)
	assert.Equal(t, c2.ID, commentsAfter[0].ID)
}

func TestProjectSoftDelete_Flow(t *testing.T) {
	db := setupTestDB(t)
	projectRepo := repository.NewProjectRepository(db)

	p := &model.Project{Name: "Soft Delete Project", Key: "SOFT"}
	err := projectRepo.Create(p)
	assert.NoError(t, err)

	// Find active
	projects, err := projectRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	// Soft delete
	err = projectRepo.Delete(p.ID)
	assert.NoError(t, err)

	// Query active: should be empty
	projectsAfter, err := projectRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, projectsAfter, 0)

	// Restore
	err = projectRepo.Restore(p.ID)
	assert.NoError(t, err)

	projectsRestored, err := projectRepo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, projectsRestored, 1)
}

func TestTimerManager_UpdateEntryTime(t *testing.T) {
	db := setupTestDB(t)

	projectRepo := repository.NewProjectRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	subtaskRepo := repository.NewSubtaskRepository(db)
	timeEntryRepo := repository.NewTimeEntryRepository(db)
	tm := service.NewTimerManager(timeEntryRepo, taskRepo, subtaskRepo)

	proj := &model.Project{Name: "Timer Project", Key: "TIME"}
	_ = projectRepo.Create(proj)
	task := &model.Task{ProjectID: proj.ID, TicketKey: "TIME-1", Title: "Timer Task"}
	_ = taskRepo.Create(nil, task)

	entry, err := tm.StartTimer(task.ID, nil)
	assert.NoError(t, err)

	_, err = tm.StopTimer(entry.ID)
	assert.NoError(t, err)

	// Update duration directly
	newDuration := int64(1800) // 30 minutes
	updated, err := tm.UpdateEntryTime(entry.ID, &newDuration, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1800), updated.DurationSeconds)

	// Verify total time spent on task was dynamically recomputed
	totalTime, err := taskRepo.GetTotalTimeSpent(task.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1800), totalTime)
}
