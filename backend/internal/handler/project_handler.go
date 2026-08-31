package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/service"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: svc}
}

// Create handles POST /api/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: name and key are required")
		return
	}

	project, err := h.service.CreateProject(req.Name, req.Key, req.Color)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProjectNameRequired), errors.Is(err, service.ErrInvalidProjectKey):
			RespondWithError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProjectKeyTaken):
			RespondWithError(c, http.StatusConflict, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Failed to create project")
		}
		return
	}

	c.JSON(http.StatusCreated, project)
}

// List handles GET /api/projects
func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.service.GetAllProjects()
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch projects")
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetByID handles GET /api/projects/:id
func (h *ProjectHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid project ID format (must be UUID)")
		return
	}

	project, err := h.service.GetProjectByID(id)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			RespondWithError(c, http.StatusNotFound, "Project not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to fetch project")
		return
	}

	c.JSON(http.StatusOK, project)
}

// Update handles PATCH /api/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid project ID format (must be UUID)")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	project, err := h.service.UpdateProject(id, req.Name, req.Color)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			RespondWithError(c, http.StatusNotFound, "Project not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	c.JSON(http.StatusOK, project)
}

// Delete handles DELETE /api/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid project ID format (must be UUID)")
		return
	}

	if err := h.service.DeleteProject(id); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			RespondWithError(c, http.StatusNotFound, "Project not found")
			return
		}
		RespondWithError(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}
