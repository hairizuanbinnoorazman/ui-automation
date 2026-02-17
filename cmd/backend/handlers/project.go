package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/project"
)

// ProjectHandler handles project-related requests.
type ProjectHandler struct {
	projectStore project.Store
	logger       logger.Logger
}

// NewProjectHandler creates a new project handler.
func NewProjectHandler(projectStore project.Store, log logger.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectStore: projectStore,
		logger:       log,
	}
}

// CreateProjectRequest represents a project creation request.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents a project update request.
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Create handles creating a new project.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Parse request body
	var req CreateProjectRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create project
	proj := &project.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		IsActive:    true,
	}

	if err := h.projectStore.Create(r.Context(), proj); err != nil {
		if errors.Is(err, project.ErrInvalidProjectName) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to create project", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		respondError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	respondJSON(w, http.StatusCreated, proj)
}

// List handles listing user's projects with pagination.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// List projects for user
	projects, err := h.projectStore.ListByOwner(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list projects", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		respondError(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	respondJSON(w, http.StatusOK, NewPaginatedResponse(projects, len(projects), limit, offset))
}

// GetByID handles getting a single project by ID.
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "project")
	if !ok {
		return
	}

	// Get project
	proj, err := h.projectStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	respondJSON(w, http.StatusOK, proj)
}

// Update handles updating a project.
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "project")
	if !ok {
		return
	}

	// Parse request body
	var req UpdateProjectRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build setters
	var setters []project.UpdateSetter
	if req.Name != nil {
		setters = append(setters, project.SetName(*req.Name))
	}
	if req.Description != nil {
		setters = append(setters, project.SetDescription(*req.Description))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Update project
	if err := h.projectStore.Update(r.Context(), id, setters...); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
		if errors.Is(err, project.ErrInvalidProjectName) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to update project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	// Get updated project to return it
	updatedProject, err := h.projectStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get updated project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get updated project")
		return
	}

	respondJSON(w, http.StatusOK, updatedProject)
}

// Delete handles soft deleting a project.
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "project")
	if !ok {
		return
	}

	// Delete project
	if err := h.projectStore.Delete(r.Context(), id); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
		h.logger.Error(r.Context(), "failed to delete project", map[string]interface{}{
			"error":      err.Error(),
			"project_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	respondSuccess(w, "project deleted successfully")
}
