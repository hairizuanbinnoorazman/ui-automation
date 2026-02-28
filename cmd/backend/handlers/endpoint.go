package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/endpoint"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
)

// EndpointHandler handles endpoint-related requests.
type EndpointHandler struct {
	endpointStore endpoint.Store
	logger        logger.Logger
}

// NewEndpointHandler creates a new endpoint handler.
func NewEndpointHandler(endpointStore endpoint.Store, log logger.Logger) *EndpointHandler {
	return &EndpointHandler{
		endpointStore: endpointStore,
		logger:        log,
	}
}

// checkEndpointOwnership verifies that the authenticated user owns the endpoint.
// Returns false if the check fails (response already written).
func (h *EndpointHandler) checkEndpointOwnership(w http.ResponseWriter, r *http.Request, endpointID uuid.UUID) bool {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return false
	}

	ep, err := h.endpointStore.GetByID(r.Context(), endpointID)
	if err != nil {
		if errors.Is(err, endpoint.ErrEndpointNotFound) {
			respondError(w, http.StatusNotFound, "endpoint not found")
			return false
		}
		h.logger.Error(r.Context(), "failed to get endpoint for authorization", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": endpointID,
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return false
	}

	if ep.CreatedBy != userID {
		h.logger.Warn(r.Context(), "unauthorized endpoint access attempt", map[string]interface{}{
			"user_id":     userID,
			"endpoint_id": endpointID,
			"created_by":  ep.CreatedBy,
		})
		respondError(w, http.StatusForbidden, "you don't have access to this endpoint")
		return false
	}

	return true
}

// CreateEndpointRequest represents an endpoint creation request.
type CreateEndpointRequest struct {
	Name        string               `json:"name"`
	URL         string               `json:"url"`
	Credentials endpoint.Credentials `json:"credentials,omitempty"`
}

// UpdateEndpointRequest represents an endpoint update request.
type UpdateEndpointRequest struct {
	Name        *string               `json:"name,omitempty"`
	URL         *string               `json:"url,omitempty"`
	Credentials *endpoint.Credentials `json:"credentials,omitempty"`
}

// Create handles creating a new endpoint.
func (h *EndpointHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req CreateEndpointRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ep := &endpoint.Endpoint{
		Name:        req.Name,
		URL:         req.URL,
		Credentials: req.Credentials,
		CreatedBy:   userID,
	}

	if err := h.endpointStore.Create(r.Context(), ep); err != nil {
		if errors.Is(err, endpoint.ErrInvalidEndpointName) ||
			errors.Is(err, endpoint.ErrInvalidEndpointURL) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to create endpoint", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create endpoint")
		return
	}

	respondJSON(w, http.StatusCreated, ep)
}

// List handles listing endpoints for the authenticated user.
func (h *EndpointHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	total, err := h.endpointStore.CountByCreator(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to count endpoints", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to count endpoints")
		return
	}

	endpoints, err := h.endpointStore.ListByCreator(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list endpoints", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list endpoints")
		return
	}

	respondJSON(w, http.StatusOK, NewPaginatedResponse(endpoints, total, limit, offset))
}

// GetByID handles getting a single endpoint by ID.
func (h *EndpointHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDOrRespond(w, r, "id", "endpoint")
	if !ok {
		return
	}

	if !h.checkEndpointOwnership(w, r, id) {
		return
	}

	ep, err := h.endpointStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, endpoint.ErrEndpointNotFound) {
			respondError(w, http.StatusNotFound, "endpoint not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get endpoint", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get endpoint")
		return
	}

	respondJSON(w, http.StatusOK, ep)
}

// Update handles updating an endpoint.
func (h *EndpointHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDOrRespond(w, r, "id", "endpoint")
	if !ok {
		return
	}

	if !h.checkEndpointOwnership(w, r, id) {
		return
	}

	var req UpdateEndpointRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var setters []endpoint.UpdateSetter
	if req.Name != nil {
		setters = append(setters, endpoint.SetName(*req.Name))
	}
	if req.URL != nil {
		setters = append(setters, endpoint.SetURL(*req.URL))
	}
	if req.Credentials != nil {
		setters = append(setters, endpoint.SetCredentials(*req.Credentials))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := h.endpointStore.Update(r.Context(), id, setters...); err != nil {
		if errors.Is(err, endpoint.ErrEndpointNotFound) {
			respondError(w, http.StatusNotFound, "endpoint not found")
			return
		}
		if errors.Is(err, endpoint.ErrInvalidEndpointName) ||
			errors.Is(err, endpoint.ErrInvalidEndpointURL) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to update endpoint", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update endpoint")
		return
	}

	updated, err := h.endpointStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get updated endpoint", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get updated endpoint")
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

// Delete handles deleting an endpoint.
func (h *EndpointHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDOrRespond(w, r, "id", "endpoint")
	if !ok {
		return
	}

	if !h.checkEndpointOwnership(w, r, id) {
		return
	}

	if err := h.endpointStore.Delete(r.Context(), id); err != nil {
		if errors.Is(err, endpoint.ErrEndpointNotFound) {
			respondError(w, http.StatusNotFound, "endpoint not found")
			return
		}
		h.logger.Error(r.Context(), "failed to delete endpoint", map[string]interface{}{
			"error":       err.Error(),
			"endpoint_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to delete endpoint")
		return
	}

	respondSuccess(w, "endpoint deleted successfully")
}
