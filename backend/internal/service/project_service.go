package service

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

var keyRegex = regexp.MustCompile(`^[A-Z0-9]{2,10}$`)

type ProjectService interface {
	CreateProject(name, key, color string) (*model.Project, error)
	GetAllProjects() ([]model.Project, error)
	GetProjectByID(id uuid.UUID) (*model.Project, error)
	UpdateProject(id uuid.UUID, name, color string) (*model.Project, error)
	DeleteProject(id uuid.UUID) error
}

type projectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) CreateProject(name, key, color string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrProjectNameRequired
	}

	key = strings.ToUpper(strings.TrimSpace(key))
	if !keyRegex.MatchString(key) {
		return nil, ErrInvalidProjectKey
	}

	// Check if key is already taken
	existing, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrProjectKeyTaken
	}

	if color == "" {
		color = "#4F46E5"
	}

	project := &model.Project{
		Name:        name,
		Key:         key,
		Color:       color,
		TaskCounter: 0,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetAllProjects() ([]model.Project, error) {
	return s.repo.FindAll()
}

func (s *projectService) GetProjectByID(id uuid.UUID) (*model.Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

func (s *projectService) UpdateProject(id uuid.UUID, name, color string) (*model.Project, error) {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	name = strings.TrimSpace(name)
	if name != "" {
		project.Name = name
	}
	if color != "" {
		project.Color = color
	}

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) DeleteProject(id uuid.UUID) error {
	project, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	return s.repo.Delete(id)
}
