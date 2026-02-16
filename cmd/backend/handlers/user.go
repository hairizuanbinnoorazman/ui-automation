package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/user"
)

// UserHandler handles user-related requests.
type UserHandler struct {
	userStore user.Store
	logger    logger.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userStore user.Store, log logger.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    log,
	}
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

// ListUsersResponse represents a list users response.
type ListUsersResponse struct {
	Users []*user.User `json:"users"`
	Total int          `json:"total"`
}

// List handles listing users with pagination.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
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

	// List users
	users, err := h.userStore.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list users", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	respondJSON(w, http.StatusOK, ListUsersResponse{
		Users: users,
		Total: len(users),
	})
}

// GetByID handles getting a single user by ID.
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Get user
	foundUser, err := h.userStore.GetByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, foundUser)
}

// Update handles updating a user.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build setters
	var setters []user.UpdateSetter

	if req.Email != nil {
		setters = append(setters, user.SetEmail(*req.Email))
	}

	if req.Username != nil {
		setters = append(setters, user.SetUsername(*req.Username))
	}

	if req.Password != nil {
		setters = append(setters, user.SetPassword(*req.Password))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Update user
	if err := h.userStore.Update(r.Context(), uint(id), setters...); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, user.ErrDuplicateEmail) {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		if errors.Is(err, user.ErrInvalidEmail) || errors.Is(err, user.ErrInvalidUsername) || errors.Is(err, user.ErrPasswordTooShort) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to update user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// Get updated user
	updatedUser, err := h.userStore.GetByID(r.Context(), uint(id))
	if err != nil {
		h.logger.Error(r.Context(), "failed to get updated user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get updated user")
		return
	}

	h.logger.Info(r.Context(), "user updated", map[string]interface{}{
		"user_id": id,
	})

	respondJSON(w, http.StatusOK, updatedUser)
}

// Delete handles soft deleting a user.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Delete user
	if err := h.userStore.Delete(r.Context(), uint(id)); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error(r.Context(), "failed to delete user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	h.logger.Info(r.Context(), "user deleted", map[string]interface{}{
		"user_id": id,
	})

	respondSuccess(w, "user deleted successfully")
}
