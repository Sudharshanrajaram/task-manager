package service

import "errors"

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectKeyTaken      = errors.New("project key is already in use")
	ErrInvalidProjectKey    = errors.New("project key must be 2-10 uppercase alphanumeric characters")
	ErrProjectNameRequired  = errors.New("project name is required")
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskTitleRequired    = errors.New("task title is required")
	ErrInvalidTaskType      = errors.New("invalid task type (must be task, bug, improvement, or spike)")
	ErrInvalidTaskStatus    = errors.New("invalid task status (must be backlog, in_progress, blocked, review, or done)")
	ErrInvalidTaskPriority  = errors.New("invalid task priority (must be p0, p1, p2, or p3)")
	ErrInvalidBugSeverity   = errors.New("invalid bug severity (must be critical, major, minor, or trivial)")
	ErrSubtaskNotFound      = errors.New("subtask not found")
	ErrSubtaskTitleRequired = errors.New("subtask title cannot be empty")
)
