package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// TestProcedureHandler handles test procedure-related requests.
type TestProcedureHandler struct {
	testProcedureStore testprocedure.Store
	logger             logger.Logger
}

// NewTestProcedureHandler creates a new test procedure handler.
func NewTestProcedureHandler(testProcedureStore testprocedure.Store, log logger.Logger) *TestProcedureHandler {
	return &TestProcedureHandler{
		testProcedureStore: testProcedureStore,
		logger:             log,
	}
}

// CreateTestProcedureRequest represents a test procedure creation request.
type CreateTestProcedureRequest struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Steps       testprocedure.Steps          `json:"steps"`
}

// UpdateTestProcedureRequest represents a test procedure update request.
type UpdateTestProcedureRequest struct {
	Name        *string                      `json:"name,omitempty"`
	Description *string                      `json:"description,omitempty"`
	Steps       *testprocedure.Steps         `json:"steps,omitempty"`
}

// ListTestProceduresResponse represents a list test procedures response.
type ListTestProceduresResponse struct {
	TestProcedures []*testprocedure.TestProcedure `json:"test_procedures"`
	Total          int                             `json:"total"`
}

// Create handles creating a new test procedure.
func (h *TestProcedureHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Extract project ID from URL
	vars := mux.Vars(r)
	projectIDStr := vars["project_id"]
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Parse request body
	var req CreateTestProcedureRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create test procedure
	tp := &testprocedure.TestProcedure{
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		ProjectID:   uint(projectID),
		CreatedBy:   userID,
	}

	if err := h.testProcedureStore.Create(r.Context(), tp); err != nil {
		if errors.Is(err, testprocedure.ErrInvalidTestProcedureName) || errors.Is(err, testprocedure.ErrInvalidSteps) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to create test procedure", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
		})
		respondError(w, http.StatusInternalServerError, "failed to create test procedure")
		return
	}

	respondJSON(w, http.StatusCreated, tp)
}

// List handles listing test procedures for a project.
func (h *TestProcedureHandler) List(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from URL
	vars := mux.Vars(r)
	projectIDStr := vars["project_id"]
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID")
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

	// List test procedures
	procedures, err := h.testProcedureStore.ListByProject(r.Context(), uint(projectID), limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list test procedures", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
		})
		respondError(w, http.StatusInternalServerError, "failed to list test procedures")
		return
	}

	respondJSON(w, http.StatusOK, ListTestProceduresResponse{
		TestProcedures: procedures,
		Total:          len(procedures),
	})
}

// GetByID handles getting a single test procedure by ID.
func (h *TestProcedureHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test procedure ID")
		return
	}

	// Get test procedure
	tp, err := h.testProcedureStore.GetByID(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get test procedure")
		return
	}

	respondJSON(w, http.StatusOK, tp)
}

// Update handles updating a test procedure (in-place, doesn't create version).
func (h *TestProcedureHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test procedure ID")
		return
	}

	// Parse request body
	var req UpdateTestProcedureRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build setters
	var setters []testprocedure.UpdateSetter
	if req.Name != nil {
		setters = append(setters, testprocedure.SetName(*req.Name))
	}
	if req.Description != nil {
		setters = append(setters, testprocedure.SetDescription(*req.Description))
	}
	if req.Steps != nil {
		setters = append(setters, testprocedure.SetSteps(*req.Steps))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Update test procedure
	if err := h.testProcedureStore.Update(r.Context(), uint(id), setters...); err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		if errors.Is(err, testprocedure.ErrInvalidTestProcedureName) || errors.Is(err, testprocedure.ErrInvalidSteps) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to update test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update test procedure")
		return
	}

	respondSuccess(w, "test procedure updated successfully")
}

// Delete handles deleting a test procedure.
func (h *TestProcedureHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test procedure ID")
		return
	}

	// Delete test procedure
	if err := h.testProcedureStore.Delete(r.Context(), uint(id)); err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to delete test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to delete test procedure")
		return
	}

	respondSuccess(w, "test procedure deleted successfully")
}

// CreateVersion handles creating a new version of a test procedure.
func (h *TestProcedureHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test procedure ID")
		return
	}

	// Create version
	newVersion, err := h.testProcedureStore.CreateVersion(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to create version", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to create version")
		return
	}

	respondJSON(w, http.StatusCreated, newVersion)
}

// GetVersionHistory handles getting version history for a test procedure.
func (h *TestProcedureHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test procedure ID")
		return
	}

	// Get version history
	versions, err := h.testProcedureStore.GetVersionHistory(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get version history", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get version history")
		return
	}

	respondJSON(w, http.StatusOK, ListTestProceduresResponse{
		TestProcedures: versions,
		Total:          len(versions),
	})
}
