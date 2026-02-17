package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/scriptgen"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// ScriptGenHandler handles script generation requests.
type ScriptGenHandler struct {
	scriptStore    scriptgen.Store
	procedureStore testprocedure.Store
	generator      scriptgen.ScriptGenerator
	storage        storage.BlobStorage
	logger         logger.Logger
}

// NewScriptGenHandler creates a new script generation handler.
func NewScriptGenHandler(
	scriptStore scriptgen.Store,
	procedureStore testprocedure.Store,
	generator scriptgen.ScriptGenerator,
	storage storage.BlobStorage,
	log logger.Logger,
) *ScriptGenHandler {
	return &ScriptGenHandler{
		scriptStore:    scriptStore,
		procedureStore: procedureStore,
		generator:      generator,
		storage:        storage,
		logger:         log,
	}
}

// GenerateScriptRequest represents a script generation request.
type GenerateScriptRequest struct {
	Framework scriptgen.Framework `json:"framework"`
}

// ListScriptsResponse represents a list scripts response.
type ListScriptsResponse struct {
	Scripts []*scriptgen.GeneratedScript `json:"scripts"`
	Total   int                          `json:"total"`
}

// Generate handles generating a new automation script.
func (h *ScriptGenHandler) Generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := GetUserID(ctx)
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
	var req GenerateScriptRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate framework
	if !req.Framework.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid framework (must be 'selenium' or 'playwright')")
		return
	}

	// Check if script already exists
	existingScript, err := h.scriptStore.GetByProcedureAndFramework(ctx, procedureID, req.Framework)
	if err == nil {
		// Script already exists, return it
		h.logger.Info(ctx, "script already exists, returning existing script", map[string]interface{}{
			"script_id":         existingScript.ID.String(),
			"test_procedure_id": procedureID.String(),
			"framework":         req.Framework,
		})
		respondJSON(w, http.StatusOK, existingScript)
		return
	}

	// If error is not "not found", return error
	if !errors.Is(err, scriptgen.ErrScriptNotFound) {
		h.logger.Error(ctx, "failed to check existing script", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
			"framework":         req.Framework,
		})
		respondError(w, http.StatusInternalServerError, "failed to check existing script")
		return
	}

	// Fetch test procedure
	procedure, err := h.procedureStore.GetByID(ctx, procedureID)
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return
		}
		h.logger.Error(ctx, "failed to fetch test procedure", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to fetch test procedure")
		return
	}

	// Generate script using LLM
	h.logger.Info(ctx, "generating script", map[string]interface{}{
		"test_procedure_id": procedureID.String(),
		"framework":         req.Framework,
	})

	scriptContent, err := h.generator.Generate(ctx, procedure, req.Framework)
	if err != nil {
		h.logger.Error(ctx, "failed to generate script", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
			"framework":         req.Framework,
		})
		respondError(w, http.StatusInternalServerError, "failed to generate script")
		return
	}

	// Generate filename
	sanitizedName := sanitizeProcedureName(procedure.Name)
	filename := fmt.Sprintf("%s_v%d_%s.py", sanitizedName, procedure.Version, req.Framework)
	storagePath := fmt.Sprintf("generated-scripts/%s/%s/%s",
		procedureID.String(),
		req.Framework,
		filename,
	)

	// Upload to storage
	reader := bytes.NewReader(scriptContent)
	if err := h.storage.Upload(ctx, storagePath, reader); err != nil {
		h.logger.Error(ctx, "failed to upload script to storage", map[string]interface{}{
			"error": err.Error(),
			"path":  storagePath,
		})
		respondError(w, http.StatusInternalServerError, "failed to save script")
		return
	}

	// Create database record
	script := &scriptgen.GeneratedScript{
		TestProcedureID:  procedureID,
		Framework:        req.Framework,
		ScriptPath:       storagePath,
		FileName:         filename,
		FileSize:         int64(len(scriptContent)),
		GenerationStatus: scriptgen.StatusCompleted,
		GeneratedBy:      userID,
		GeneratedAt:      time.Now(),
	}

	if err := h.scriptStore.Create(ctx, script); err != nil {
		// Try to clean up uploaded file
		if delErr := h.storage.Delete(ctx, storagePath); delErr != nil {
			h.logger.Warn(ctx, "failed to cleanup uploaded script after db error", map[string]interface{}{
				"delete_error": delErr.Error(),
				"path":         storagePath,
			})
		}

		h.logger.Error(ctx, "failed to create script record", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to save script record")
		return
	}

	h.logger.Info(ctx, "script generated successfully", map[string]interface{}{
		"script_id":         script.ID.String(),
		"test_procedure_id": procedureID.String(),
		"framework":         req.Framework,
		"file_size":         script.FileSize,
	})

	respondJSON(w, http.StatusCreated, script)
}

// List handles listing all scripts for a test procedure.
func (h *ScriptGenHandler) List(w http.ResponseWriter, r *http.Request) {
	// Extract test procedure ID from URL
	procedureID, ok := parseUUIDOrRespond(w, r, "procedure_id", "test procedure")
	if !ok {
		return
	}

	// List scripts
	scripts, err := h.scriptStore.ListByProcedure(r.Context(), procedureID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list scripts", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list scripts")
		return
	}

	respondJSON(w, http.StatusOK, ListScriptsResponse{
		Scripts: scripts,
		Total:   len(scripts),
	})
}

// GetByID handles retrieving a script by its ID.
func (h *ScriptGenHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract script ID from URL
	scriptID, ok := parseUUIDOrRespond(w, r, "script_id", "script")
	if !ok {
		return
	}

	// Get script
	script, err := h.scriptStore.GetByID(r.Context(), scriptID)
	if err != nil {
		if errors.Is(err, scriptgen.ErrScriptNotFound) {
			respondError(w, http.StatusNotFound, "script not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get script", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to get script")
		return
	}

	respondJSON(w, http.StatusOK, script)
}

// Download handles downloading a script file.
func (h *ScriptGenHandler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract script ID from URL
	scriptID, ok := parseUUIDOrRespond(w, r, "script_id", "script")
	if !ok {
		return
	}

	// Get script metadata
	script, err := h.scriptStore.GetByID(ctx, scriptID)
	if err != nil {
		if errors.Is(err, scriptgen.ErrScriptNotFound) {
			respondError(w, http.StatusNotFound, "script not found")
			return
		}
		h.logger.Error(ctx, "failed to get script", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to get script")
		return
	}

	// Download from storage
	reader, err := h.storage.Download(ctx, script.ScriptPath)
	if err != nil {
		h.logger.Error(ctx, "failed to download script from storage", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
			"path":      script.ScriptPath,
		})
		respondError(w, http.StatusInternalServerError, "failed to download script")
		return
	}
	defer reader.Close()

	// Set response headers
	w.Header().Set("Content-Type", "text/x-python")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", script.FileName))

	// Stream file to response
	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Error(ctx, "failed to stream script to response", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		return
	}

	h.logger.Info(ctx, "script downloaded", map[string]interface{}{
		"script_id": scriptID.String(),
		"filename":  script.FileName,
	})
}

// Delete handles deleting a script.
func (h *ScriptGenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract script ID from URL
	scriptID, ok := parseUUIDOrRespond(w, r, "script_id", "script")
	if !ok {
		return
	}

	// Get script metadata (to get storage path)
	script, err := h.scriptStore.GetByID(ctx, scriptID)
	if err != nil {
		if errors.Is(err, scriptgen.ErrScriptNotFound) {
			respondError(w, http.StatusNotFound, "script not found")
			return
		}
		h.logger.Error(ctx, "failed to get script", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to get script")
		return
	}

	// Delete from database first
	if err := h.scriptStore.Delete(ctx, scriptID); err != nil {
		h.logger.Error(ctx, "failed to delete script from database", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to delete script")
		return
	}

	// Delete from storage (best effort)
	if err := h.storage.Delete(ctx, script.ScriptPath); err != nil {
		h.logger.Warn(ctx, "failed to delete script from storage", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
			"path":      script.ScriptPath,
		})
		// Don't fail the request - DB record is already deleted
	}

	h.logger.Info(ctx, "script deleted", map[string]interface{}{
		"script_id": scriptID.String(),
	})

	w.WriteHeader(http.StatusNoContent)
}

// sanitizeProcedureName removes or replaces characters that are problematic in filenames.
func sanitizeProcedureName(name string) string {
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace other problematic characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)

	// Limit length
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}
