package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/project"
)

const (
	// ProjectKey is the context key for project.
	ProjectKey ContextKey = "project"
)

// ProjectAuthorizationMiddleware validates that the current user owns the project.
type ProjectAuthorizationMiddleware struct {
	projectStore project.Store
	logger       logger.Logger
}

// NewProjectAuthorizationMiddleware creates a new project authorization middleware.
func NewProjectAuthorizationMiddleware(projectStore project.Store, log logger.Logger) *ProjectAuthorizationMiddleware {
	return &ProjectAuthorizationMiddleware{
		projectStore: projectStore,
		logger:       log,
	}
}

// Handler wraps an HTTP handler with project authorization.
func (m *ProjectAuthorizationMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, ok := GetUserID(r.Context())
		if !ok {
			respondError(w, http.StatusUnauthorized, "user not authenticated")
			return
		}

		// Extract project ID from URL
		vars := mux.Vars(r)
		idStr := vars["id"]
		if idStr == "" {
			// For routes like /api/v1/projects (without ID), skip authorization
			next.ServeHTTP(w, r)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID: must be a valid UUID")
			return
		}

		// Get project
		proj, err := m.projectStore.GetByID(r.Context(), id)
		if err != nil {
			if err == project.ErrProjectNotFound {
				respondError(w, http.StatusNotFound, "project not found")
				return
			}
			m.logger.Error(r.Context(), "failed to get project for authorization", map[string]interface{}{
				"error":      err.Error(),
				"project_id": id,
			})
			respondError(w, http.StatusInternalServerError, "authorization check failed")
			return
		}

		// Check if user owns the project
		if proj.OwnerID != userID {
			m.logger.Warn(r.Context(), "unauthorized project access attempt", map[string]interface{}{
				"user_id":    userID,
				"project_id": id,
				"owner_id":   proj.OwnerID,
			})
			respondError(w, http.StatusForbidden, "you don't have access to this project")
			return
		}

		// Add project to context for use by handlers
		ctx := context.WithValue(r.Context(), ProjectKey, proj)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetProject extracts the project from the request context.
func GetProject(ctx context.Context) (*project.Project, bool) {
	proj, ok := ctx.Value(ProjectKey).(*project.Project)
	return proj, ok
}
