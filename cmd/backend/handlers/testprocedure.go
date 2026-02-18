package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/project"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// TestProcedureHandler handles test procedure-related requests.
type TestProcedureHandler struct {
	testProcedureStore testprocedure.Store
	projectStore       project.Store
	storage            storage.BlobStorage
	logger             logger.Logger
}

// NewTestProcedureHandler creates a new test procedure handler.
func NewTestProcedureHandler(testProcedureStore testprocedure.Store, projectStore project.Store, storage storage.BlobStorage, log logger.Logger) *TestProcedureHandler {
	return &TestProcedureHandler{
		testProcedureStore: testProcedureStore,
		projectStore:       projectStore,
		storage:            storage,
		logger:             log,
	}
}

// checkProcedureOwnership verifies that the authenticated user owns the project
// associated with the given procedure. Returns false if the check fails (response
// already written).
func (h *TestProcedureHandler) checkProcedureOwnership(w http.ResponseWriter, r *http.Request, procedureID uuid.UUID) bool {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return false
	}

	tp, err := h.testProcedureStore.GetByID(r.Context(), procedureID)
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return false
		}
		h.logger.Error(r.Context(), "failed to get test procedure for authorization", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID,
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return false
	}

	proj, err := h.projectStore.GetByID(r.Context(), tp.ProjectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return false
		}
		h.logger.Error(r.Context(), "failed to get project for authorization", map[string]interface{}{
			"error":      err.Error(),
			"project_id": tp.ProjectID,
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return false
	}

	if proj.OwnerID != userID {
		h.logger.Warn(r.Context(), "unauthorized procedure access attempt", map[string]interface{}{
			"user_id":           userID,
			"project_id":        tp.ProjectID,
			"owner_id":          proj.OwnerID,
			"test_procedure_id": procedureID,
		})
		respondError(w, http.StatusForbidden, "you don't have access to this test procedure")
		return false
	}

	return true
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

// Create handles creating a new test procedure.
func (h *TestProcedureHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Extract project ID from URL
	projectID, ok := parseUUIDOrRespond(w, r, "project_id", "project")
	if !ok {
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
		ProjectID:   projectID,
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
	projectID, ok := parseUUIDOrRespond(w, r, "project_id", "project")
	if !ok {
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

	// Get total count of test procedures
	total, err := h.testProcedureStore.CountByProject(r.Context(), projectID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to count test procedures", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
		})
		respondError(w, http.StatusInternalServerError, "failed to count test procedures")
		return
	}

	// List test procedures
	procedures, err := h.testProcedureStore.ListByProject(r.Context(), projectID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list test procedures", map[string]interface{}{
			"error":      err.Error(),
			"project_id": projectID,
		})
		respondError(w, http.StatusInternalServerError, "failed to list test procedures")
		return
	}

	respondJSON(w, http.StatusOK, NewPaginatedResponse(procedures, total, limit, offset))
}

// GetByID handles getting a single test procedure by ID.
// Supports ?draft=true query parameter to retrieve draft version.
func (h *TestProcedureHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Check if draft version is requested
	isDraft := r.URL.Query().Get("draft") == "true"

	var tp *testprocedure.TestProcedure
	var err error

	if isDraft {
		tp, err = h.testProcedureStore.GetDraft(r.Context(), id)
		if err != nil {
			if errors.Is(err, testprocedure.ErrDraftNotFound) {
				respondError(w, http.StatusNotFound, "draft version not found")
				return
			}
			h.logger.Error(r.Context(), "failed to get draft", map[string]interface{}{
				"error":             err.Error(),
				"test_procedure_id": id,
			})
			respondError(w, http.StatusInternalServerError, "failed to get draft")
			return
		}
	} else {
		tp, err = h.testProcedureStore.GetLatestCommitted(r.Context(), id)
		if err != nil {
			if errors.Is(err, testprocedure.ErrNoCommittedVersion) {
				respondError(w, http.StatusNotFound, "no committed version exists")
				return
			}
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
	}

	respondJSON(w, http.StatusOK, tp)
}

// Update handles updating a test procedure draft.
// This always updates the draft (v0), not the committed version.
func (h *TestProcedureHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
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

	// Update draft
	if err := h.testProcedureStore.UpdateDraft(r.Context(), id, setters...); err != nil {
		if errors.Is(err, testprocedure.ErrDraftNotFound) {
			respondError(w, http.StatusNotFound, "draft not found")
			return
		}
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		if errors.Is(err, testprocedure.ErrInvalidTestProcedureName) || errors.Is(err, testprocedure.ErrInvalidSteps) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to update draft", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update draft")
		return
	}

	// Get updated draft to return it
	updatedDraft, err := h.testProcedureStore.GetDraft(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get updated draft", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get updated draft")
		return
	}

	respondJSON(w, http.StatusOK, updatedDraft)
}

// Delete handles deleting a test procedure.
func (h *TestProcedureHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Delete test procedure
	if err := h.testProcedureStore.Delete(r.Context(), id); err != nil {
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
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Create version
	newVersion, err := h.testProcedureStore.CreateVersion(r.Context(), id)
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
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Get version history
	versions, err := h.testProcedureStore.GetVersionHistory(r.Context(), id)
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

	respondJSON(w, http.StatusOK, versions)
}

// UploadStepImage handles uploading an image for a test procedure step.
func (h *TestProcedureHandler) UploadStepImage(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Verify the authenticated user owns the project this procedure belongs to
	if !h.checkProcedureOwnership(w, r, id) {
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "image file is required")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !validExts[ext] {
		respondError(w, http.StatusBadRequest, "invalid file type, must be JPEG, PNG, GIF, or WebP")
		return
	}

	// Validate file content using magic bytes (not just the extension)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		respondError(w, http.StatusBadRequest, "failed to read file")
		return
	}
	contentType := http.DetectContentType(buf[:n])
	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !validMimeTypes[contentType] {
		respondError(w, http.StatusBadRequest, "invalid file content, must be JPEG, PNG, GIF, or WebP")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	path := fmt.Sprintf("test-procedures/%s/steps/%s", id.String(), filename)

	// Upload to storage
	if err := h.storage.Upload(r.Context(), path, file); err != nil {
		h.logger.Error(r.Context(), "failed to upload image", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id.String(),
			"path":              path,
		})
		respondError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	h.logger.Info(r.Context(), "image uploaded", map[string]interface{}{
		"test_procedure_id": id.String(),
		"path":              path,
	})

	// Return the image path
	respondJSON(w, http.StatusOK, map[string]string{
		"image_path": path,
	})
}

// DraftDiffResponse represents the response for GetDiff.
type DraftDiffResponse struct {
	Draft     *testprocedure.TestProcedure `json:"draft"`
	Committed *testprocedure.TestProcedure `json:"committed"`
}

// GetDiff handles getting both draft and committed versions for comparison.
func (h *TestProcedureHandler) GetDiff(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Verify the authenticated user owns the project this procedure belongs to
	if !h.checkProcedureOwnership(w, r, id) {
		return
	}

	var response DraftDiffResponse

	// Get draft version
	draft, err := h.testProcedureStore.GetDraft(r.Context(), id)
	if err != nil {
		if !errors.Is(err, testprocedure.ErrDraftNotFound) {
			h.logger.Error(r.Context(), "failed to get draft", map[string]interface{}{
				"error":             err.Error(),
				"test_procedure_id": id,
			})
			respondError(w, http.StatusInternalServerError, "failed to get draft")
			return
		}
		// Draft not found is acceptable, leave nil
	} else {
		response.Draft = draft
	}

	// Get latest committed version
	committed, err := h.testProcedureStore.GetLatestCommitted(r.Context(), id)
	if err != nil {
		if !errors.Is(err, testprocedure.ErrNoCommittedVersion) {
			h.logger.Error(r.Context(), "failed to get committed version", map[string]interface{}{
				"error":             err.Error(),
				"test_procedure_id": id,
			})
			respondError(w, http.StatusInternalServerError, "failed to get committed version")
			return
		}
		// No committed version is acceptable, leave nil
	} else {
		response.Committed = committed
	}

	respondJSON(w, http.StatusOK, response)
}

// ResetDraft handles resetting the draft to match the latest committed version.
func (h *TestProcedureHandler) ResetDraft(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Verify the authenticated user owns the project this procedure belongs to
	if !h.checkProcedureOwnership(w, r, id) {
		return
	}

	// Reset draft
	if err := h.testProcedureStore.ResetDraft(r.Context(), id); err != nil {
		if errors.Is(err, testprocedure.ErrNoCommittedVersion) {
			respondError(w, http.StatusBadRequest, "no committed version exists to reset from")
			return
		}
		if errors.Is(err, testprocedure.ErrDraftNotFound) {
			respondError(w, http.StatusNotFound, "draft not found")
			return
		}
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to reset draft", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to reset draft")
		return
	}

	respondSuccess(w, "draft reset successfully")
}

// CommitDraft handles committing the draft as a new version.
func (h *TestProcedureHandler) CommitDraft(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
	if !ok {
		return
	}

	// Verify the authenticated user owns the project this procedure belongs to
	if !h.checkProcedureOwnership(w, r, id) {
		return
	}

	// Commit draft
	newVersion, err := h.testProcedureStore.CommitDraft(r.Context(), id)
	if err != nil {
		if errors.Is(err, testprocedure.ErrDraftNotFound) {
			respondError(w, http.StatusNotFound, "draft not found")
			return
		}
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(r.Context(), "failed to commit draft", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to commit draft")
		return
	}

	respondJSON(w, http.StatusCreated, newVersion)
}
