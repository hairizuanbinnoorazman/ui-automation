package handlers

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
	"github.com/hairizuan-noorazman/ui-automation/testrun"
)

const (
	// MaxUploadSize is the maximum file upload size (100MB)
	MaxUploadSize = 100 * 1024 * 1024
)

// TestRunHandler handles test run-related requests.
type TestRunHandler struct {
	testRunStore       testrun.Store
	assetStore         testrun.AssetStore
	testProcedureStore testprocedure.Store
	storage            storage.BlobStorage
	logger             logger.Logger
}

// NewTestRunHandler creates a new test run handler.
func NewTestRunHandler(testRunStore testrun.Store, assetStore testrun.AssetStore, testProcedureStore testprocedure.Store, storage storage.BlobStorage, log logger.Logger) *TestRunHandler {
	return &TestRunHandler{
		testRunStore:       testRunStore,
		assetStore:         assetStore,
		testProcedureStore: testProcedureStore,
		storage:            storage,
		logger:             log,
	}
}

// CreateTestRunRequest represents a test run creation request.
type CreateTestRunRequest struct {
	Notes string `json:"notes"`
}

// UpdateTestRunRequest represents a test run update request.
type UpdateTestRunRequest struct {
	Notes *string `json:"notes,omitempty"`
}

// CompleteTestRunRequest represents a test run completion request.
type CompleteTestRunRequest struct {
	Status testrun.Status `json:"status"`
	Notes  string         `json:"notes"`
}

// Create handles creating a new test run.
func (h *TestRunHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Extract test procedure ID from URL
	procedureID, ok := parseUUIDOrRespond(w, r, "procedure_id", "test procedure")
	if !ok {
		return
	}

	// Parse request body
	var req CreateTestRunRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create test run
	tr := &testrun.TestRun{
		TestProcedureID: procedureID,
		ExecutedBy:      userID,
		Notes:           req.Notes,
		Status:          testrun.StatusPending,
	}

	if err := h.testRunStore.Create(r.Context(), tr); err != nil {
		h.logger.Error(r.Context(), "failed to create test run", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID,
		})
		respondError(w, http.StatusInternalServerError, "failed to create test run")
		return
	}

	respondJSON(w, http.StatusCreated, tr)
}

// List handles listing test runs for a test procedure.
func (h *TestRunHandler) List(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	procedureID, ok := parseUUIDOrRespond(w, r, "procedure_id", "test procedure")
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

	// Get total count of test runs
	total, err := h.testRunStore.CountByTestProcedure(r.Context(), procedureID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to count test runs", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID,
		})
		respondError(w, http.StatusInternalServerError, "failed to count test runs")
		return
	}

	// List test runs
	runs, err := h.testRunStore.ListByTestProcedure(r.Context(), procedureID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list test runs", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID,
		})
		respondError(w, http.StatusInternalServerError, "failed to list test runs")
		return
	}

	respondJSON(w, http.StatusOK, NewPaginatedResponse(runs, total, limit, offset))
}

// GetByID handles getting a single test run by ID.
func (h *TestRunHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// Get test run
	tr, err := h.testRunStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get test run")
		return
	}

	respondJSON(w, http.StatusOK, tr)
}

// Update handles updating a test run.
func (h *TestRunHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// Parse request body
	var req UpdateTestRunRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build setters
	var setters []testrun.UpdateSetter
	if req.Notes != nil {
		setters = append(setters, testrun.SetNotes(*req.Notes))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Update test run
	if err := h.testRunStore.Update(r.Context(), id, setters...); err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		h.logger.Error(r.Context(), "failed to update test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to update test run")
		return
	}

	// Get updated test run to return it
	updatedRun, err := h.testRunStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get updated test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get updated test run")
		return
	}

	respondJSON(w, http.StatusOK, updatedRun)
}

// Start handles starting a test run.
func (h *TestRunHandler) Start(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// Start test run
	if err := h.testRunStore.Start(r.Context(), id); err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		if errors.Is(err, testrun.ErrTestRunAlreadyStarted) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to start test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to start test run")
		return
	}

	// Get the started test run to return it
	startedRun, err := h.testRunStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get started test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get started test run")
		return
	}

	respondJSON(w, http.StatusOK, startedRun)
}

// Complete handles completing a test run.
func (h *TestRunHandler) Complete(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// Parse request body
	var req CompleteTestRunRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Complete test run
	if err := h.testRunStore.Complete(r.Context(), id, req.Status, req.Notes); err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		if errors.Is(err, testrun.ErrTestRunNotRunning) || errors.Is(err, testrun.ErrInvalidStatus) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to complete test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to complete test run")
		return
	}

	// Get the completed test run to return it
	completedRun, err := h.testRunStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get completed test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get completed test run")
		return
	}

	respondJSON(w, http.StatusOK, completedRun)
}

// UploadAsset handles uploading an asset for a test run.
func (h *TestRunHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// Verify test run exists
	_, err := h.testRunStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to verify test run")
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		h.logger.Error(r.Context(), "failed to parse multipart form", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, "file too large or invalid form data")
		return
	}

	// Get asset_type parameter
	assetTypeStr := r.FormValue("asset_type")
	if assetTypeStr == "" {
		respondError(w, http.StatusBadRequest, "asset_type is required")
		return
	}
	assetType := testrun.AssetType(assetTypeStr)
	if !assetType.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid asset_type")
		return
	}

	// Get optional description
	description := r.FormValue("description")

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Sanitize filename
	filename := sanitizeFilename(header.Filename)
	if filename == "" {
		respondError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	// Generate storage path
	storagePath := fmt.Sprintf("test-runs/%d/%s/%s", id, assetType, filename)

	// Upload to storage
	if err := h.storage.Upload(r.Context(), storagePath, file); err != nil {
		h.logger.Error(r.Context(), "failed to upload file to storage", map[string]interface{}{
			"error": err.Error(),
			"path":  storagePath,
		})
		respondError(w, http.StatusInternalServerError, "failed to upload file")
		return
	}

	// Get file size
	fileSize := header.Size

	// Create asset record
	asset := &testrun.TestRunAsset{
		TestRunID:   id,
		AssetType:   assetType,
		AssetPath:   storagePath,
		FileName:    filename,
		FileSize:    fileSize,
		MimeType:    header.Header.Get("Content-Type"),
		Description: description,
	}

	if err := h.assetStore.Create(r.Context(), asset); err != nil {
		// Clean up uploaded file on database error
		h.storage.Delete(r.Context(), storagePath)
		h.logger.Error(r.Context(), "failed to create asset record", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create asset record")
		return
	}

	respondJSON(w, http.StatusCreated, asset)
}

// ListAssets handles listing assets for a test run.
func (h *TestRunHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	// List assets
	assets, err := h.assetStore.ListByTestRun(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list assets", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to list assets")
		return
	}

	respondJSON(w, http.StatusOK, assets)
}

// DownloadAsset handles downloading an asset.
func (h *TestRunHandler) DownloadAsset(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from URL
	assetID, ok := parseUUIDOrRespond(w, r, "asset_id", "asset")
	if !ok {
		return
	}

	// Get asset
	asset, err := h.assetStore.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, testrun.ErrAssetNotFound) {
			respondError(w, http.StatusNotFound, "asset not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get asset", map[string]interface{}{
			"error":    err.Error(),
			"asset_id": assetID,
		})
		respondError(w, http.StatusInternalServerError, "failed to get asset")
		return
	}

	// Download from storage
	reader, err := h.storage.Download(r.Context(), asset.AssetPath)
	if err != nil {
		if errors.Is(err, storage.ErrFileNotFound) {
			respondError(w, http.StatusNotFound, "file not found in storage")
			return
		}
		h.logger.Error(r.Context(), "failed to download from storage", map[string]interface{}{
			"error": err.Error(),
			"path":  asset.AssetPath,
		})
		respondError(w, http.StatusInternalServerError, "failed to download file")
		return
	}
	defer reader.Close()

	// Set response headers
	w.Header().Set("Content-Type", asset.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", asset.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(asset.FileSize, 10))

	// Stream file to response
	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Error(r.Context(), "failed to stream file", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// DeleteAsset handles deleting an asset.
func (h *TestRunHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	// Extract asset ID from URL
	assetID, ok := parseUUIDOrRespond(w, r, "asset_id", "asset")
	if !ok {
		return
	}

	// Get asset to get storage path
	asset, err := h.assetStore.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, testrun.ErrAssetNotFound) {
			respondError(w, http.StatusNotFound, "asset not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get asset", map[string]interface{}{
			"error":    err.Error(),
			"asset_id": assetID,
		})
		respondError(w, http.StatusInternalServerError, "failed to get asset")
		return
	}

	// Delete from database first
	if err := h.assetStore.Delete(r.Context(), assetID); err != nil {
		h.logger.Error(r.Context(), "failed to delete asset record", map[string]interface{}{
			"error":    err.Error(),
			"asset_id": assetID,
		})
		respondError(w, http.StatusInternalServerError, "failed to delete asset")
		return
	}

	// Delete from storage (best effort - log error but don't fail request)
	if err := h.storage.Delete(r.Context(), asset.AssetPath); err != nil {
		h.logger.Warn(r.Context(), "failed to delete file from storage", map[string]interface{}{
			"error": err.Error(),
			"path":  asset.AssetPath,
		})
	}

	respondSuccess(w, "asset deleted successfully")
}

// GenerateGuide creates a ZIP archive containing a guide.md and all run assets.
func (h *TestRunHandler) GenerateGuide(w http.ResponseWriter, r *http.Request) {
	// Extract test run ID from URL
	id, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	ctx := r.Context()

	// Fetch test run
	tr, err := h.testRunStore.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return
		}
		h.logger.Error(ctx, "failed to get test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get test run")
		return
	}

	// Fetch test procedure
	proc, err := h.testProcedureStore.GetByID(ctx, tr.TestProcedureID)
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(ctx, "failed to get test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": tr.TestProcedureID,
		})
		respondError(w, http.StatusInternalServerError, "failed to get test procedure")
		return
	}

	// Fetch all assets
	assets, err := h.assetStore.ListByTestRun(ctx, id)
	if err != nil {
		h.logger.Error(ctx, "failed to list assets", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to list assets")
		return
	}

	// Build guide.md content
	var md strings.Builder
	fmt.Fprintf(&md, "# %s\n\n", proc.Name)
	if proc.Description != "" {
		fmt.Fprintf(&md, "%s\n\n", proc.Description)
	}
	fmt.Fprintf(&md, "## Overview\n\n")
	if tr.Notes != "" {
		fmt.Fprintf(&md, "%s\n\n", tr.Notes)
	}
	fmt.Fprintf(&md, "---\n\n")

	for i, asset := range assets {
		assetEntry := fmt.Sprintf("%s_%s", asset.ID.String(), asset.FileName)
		fmt.Fprintf(&md, "## Step %d\n\n", i+1)
		if asset.AssetType == testrun.AssetTypeImage {
			fmt.Fprintf(&md, "![Step %d](./assets/%s)\n\n", i+1, assetEntry)
		} else {
			fmt.Fprintf(&md, "[%s](./assets/%s)\n\n", asset.FileName, assetEntry)
		}
		if asset.Description != "" {
			fmt.Fprintf(&md, "%s\n\n", asset.Description)
		}
		fmt.Fprintf(&md, "---\n\n")
	}

	// Stream ZIP archive directly to the response writer
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", "guide-"+id.String()+".zip"))
	zw := zip.NewWriter(w)

	// Write guide.md
	guideWriter, err := zw.Create("guide.md")
	if err != nil {
		h.logger.Error(ctx, "failed to create guide.md in zip", map[string]interface{}{"error": err.Error()})
		return
	}
	if _, err := io.WriteString(guideWriter, md.String()); err != nil {
		h.logger.Error(ctx, "failed to write guide.md", map[string]interface{}{"error": err.Error()})
		return
	}

	// Write each asset into assets/ folder
	for _, asset := range assets {
		reader, err := h.storage.Download(ctx, asset.AssetPath)
		if err != nil {
			h.logger.Error(ctx, "failed to download asset for guide", map[string]interface{}{
				"error": err.Error(),
				"path":  asset.AssetPath,
			})
			return
		}

		assetEntry := fmt.Sprintf("%s_%s", asset.ID.String(), asset.FileName)
		assetWriter, err := zw.Create("assets/" + assetEntry)
		if err != nil {
			reader.Close()
			h.logger.Error(ctx, "failed to create asset entry in zip", map[string]interface{}{"error": err.Error()})
			return
		}

		if _, err := io.Copy(assetWriter, reader); err != nil {
			reader.Close()
			h.logger.Error(ctx, "failed to write asset to zip", map[string]interface{}{"error": err.Error()})
			return
		}
		reader.Close()
	}

	if err := zw.Close(); err != nil {
		h.logger.Error(ctx, "failed to close zip writer", map[string]interface{}{"error": err.Error()})
	}
}

// sanitizeFilename removes potentially dangerous characters from filenames.
func sanitizeFilename(filename string) string {
	// Get base name to remove any directory paths
	filename = filepath.Base(filename)

	// Remove any remaining path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Trim spaces
	filename = strings.TrimSpace(filename)

	return filename
}

// getFileFromMultipart extracts a file from multipart form data.
func getFileFromMultipart(file multipart.File) ([]byte, error) {
	return io.ReadAll(file)
}
