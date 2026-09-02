package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

var (
	ErrCommentContentRequired = errors.New("comment content cannot be empty")
	ErrCommentNotFound        = errors.New("comment not found")
)

type CommentService interface {
	CreateComment(taskID uuid.UUID, content string) (*model.Comment, error)
	GetCommentsByTaskID(taskID uuid.UUID) ([]model.Comment, error)
	DeleteComment(id uuid.UUID) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
}

func NewCommentService(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
	}
}

func (s *commentService) CreateComment(taskID uuid.UUID, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrCommentContentRequired
	}

	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	comment := &model.Comment{
		TaskID:  taskID,
		Content: content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *commentService) GetCommentsByTaskID(taskID uuid.UUID) ([]model.Comment, error) {
	return s.commentRepo.FindByTaskID(taskID)
}

func (s *commentService) DeleteComment(id uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	return s.commentRepo.Delete(id)
}
