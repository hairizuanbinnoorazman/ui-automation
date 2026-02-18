package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/project"
	"github.com/hairizuan-noorazman/ui-automation/scriptgen"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// generatingTimeout is the maximum time a script may remain in StatusGenerating
// before it is considered stuck and eligible for regeneration.
const generatingTimeout = 10 * time.Minute

// ScriptGenHandler handles script generation requests.
type ScriptGenHandler struct {
	scriptStore    scriptgen.Store
	procedureStore testprocedure.Store
	projectStore   project.Store
	generator      scriptgen.ScriptGenerator
	storage        storage.BlobStorage
	logger         logger.Logger
}

// NewScriptGenHandler creates a new script generation handler.
func NewScriptGenHandler(
	scriptStore scriptgen.Store,
	procedureStore testprocedure.Store,
	projectStore project.Store,
	generator scriptgen.ScriptGenerator,
	storage storage.BlobStorage,
	log logger.Logger,
) *ScriptGenHandler {
	return &ScriptGenHandler{
		scriptStore:    scriptStore,
		procedureStore: procedureStore,
		projectStore:   projectStore,
		generator:      generator,
		storage:        storage,
		logger:         log,
	}
}

// verifyProcedureOwnership checks if the authenticated user owns the project
// containing the specified test procedure. Returns the procedure if authorized.
func (h *ScriptGenHandler) verifyProcedureOwnership(
	w http.ResponseWriter,
	ctx context.Context,
	procedureID uuid.UUID,
	userID uuid.UUID,
) (*testprocedure.TestProcedure, bool) {
	// Fetch the test procedure
	procedure, err := h.procedureStore.GetByID(ctx, procedureID)
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return nil, false
		}
		h.logger.Error(ctx, "failed to fetch test procedure for authorization", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return nil, false
	}

	// Fetch the project to verify ownership
	proj, err := h.projectStore.GetByID(ctx, procedure.ProjectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return nil, false
		}
		h.logger.Error(ctx, "failed to fetch project for authorization", map[string]interface{}{
			"error":      err.Error(),
			"project_id": procedure.ProjectID.String(),
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return nil, false
	}

	// Verify ownership
	if proj.OwnerID != userID {
		h.logger.Warn(ctx, "unauthorized procedure access attempt", map[string]interface{}{
			"user_id":           userID.String(),
			"test_procedure_id": procedureID.String(),
			"project_id":        procedure.ProjectID.String(),
			"owner_id":          proj.OwnerID.String(),
		})
		respondError(w, http.StatusForbidden, "you don't have access to this test procedure")
		return nil, false
	}

	return procedure, true
}

// GenerateScriptRequest represents a script generation request.
type GenerateScriptRequest struct {
	Framework scriptgen.Framework `json:"framework"`
}

// ListScriptsResponse represents a list scripts response.
type ListScriptsResponse = PaginatedResponse

// Generate handles generating a new automation script.
// It creates a DB record with StatusGenerating, returns 202 Accepted immediately,
// and performs the LLM call and storage upload in a background goroutine.
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

	// Verify user owns the procedure's project BEFORE checking for existing scripts
	procedure, ok := h.verifyProcedureOwnership(w, ctx, procedureID, userID)
	if !ok {
		// Helper already logged and responded with appropriate error
		return
	}

	// Check if script already exists (including any in-progress generation)
	existingScript, err := h.scriptStore.GetByProcedureAndFramework(ctx, procedureID, req.Framework)
	if err == nil {
		isStuckGenerating := existingScript.GenerationStatus == scriptgen.StatusGenerating &&
			time.Since(existingScript.GeneratedAt) > generatingTimeout
		isFailed := existingScript.GenerationStatus == scriptgen.StatusFailed

		if isStuckGenerating || isFailed {
			h.logger.Info(ctx, "deleting stale/failed script to allow regeneration", map[string]interface{}{
				"script_id": existingScript.ID.String(),
				"status":    existingScript.GenerationStatus,
				"age":       time.Since(existingScript.GeneratedAt).String(),
			})
			// Best-effort cleanup of any partially uploaded artifact.
			if delErr := h.storage.Delete(ctx, existingScript.ScriptPath); delErr != nil {
				h.logger.Warn(ctx, "failed to cleanup stale script from storage", map[string]interface{}{
					"delete_error": delErr.Error(),
					"path":         existingScript.ScriptPath,
				})
			}
			if deleteErr := h.scriptStore.Delete(ctx, existingScript.ID); deleteErr != nil {
				h.logger.Error(ctx, "failed to delete stale script record", map[string]interface{}{
					"error":     deleteErr.Error(),
					"script_id": existingScript.ID.String(),
				})
				respondError(w, http.StatusInternalServerError, "failed to cleanup stale script")
				return
			}
			// Mark err as not-found so the check below treats this as a fresh start.
			err = scriptgen.ErrScriptNotFound
			// Fall through to create a new record.
		} else {
			h.logger.Info(ctx, "script already exists, returning existing script", map[string]interface{}{
				"script_id":         existingScript.ID.String(),
				"test_procedure_id": procedureID.String(),
				"framework":         req.Framework,
				"status":            existingScript.GenerationStatus,
			})
			respondJSON(w, http.StatusOK, existingScript)
			return
		}
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

	// Compute filename and storage path upfront — these are deterministic and
	// do not require the LLM result.
	sanitizedName := sanitizeProcedureName(procedure.Name)
	filename := fmt.Sprintf("%s_v%d_%s.py", sanitizedName, procedure.Version, req.Framework)
	storagePath := fmt.Sprintf("generated-scripts/%s/%s/%s",
		procedureID.String(),
		req.Framework,
		filename,
	)

	// Create the DB record immediately so the client can track progress.
	script := &scriptgen.GeneratedScript{
		TestProcedureID:  procedureID,
		Framework:        req.Framework,
		ScriptPath:       storagePath,
		FileName:         filename,
		FileSize:         0,
		GenerationStatus: scriptgen.StatusGenerating,
		GeneratedBy:      userID,
		GeneratedAt:      time.Now(),
	}

	if err := h.scriptStore.Create(ctx, script); err != nil {
		h.logger.Error(ctx, "failed to create script record", map[string]interface{}{
			"error":             err.Error(),
			"test_procedure_id": procedureID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create script record")
		return
	}

	// Kick off background generation. A detached context is used so the goroutine
	// is not cancelled when the HTTP request context expires.
	go h.generateInBackground(context.Background(), script.ID, procedure, req.Framework, storagePath)

	h.logger.Info(ctx, "script generation started", map[string]interface{}{
		"script_id":         script.ID.String(),
		"test_procedure_id": procedureID.String(),
		"framework":         req.Framework,
	})

	respondJSON(w, http.StatusAccepted, script)
}

// generateInBackground performs the LLM call, storage upload, and final DB update
// for an async script generation request. It must be called in a goroutine and
// must use a context that is not tied to an HTTP request lifetime.
func (h *ScriptGenHandler) generateInBackground(
	ctx context.Context,
	scriptID uuid.UUID,
	procedure *testprocedure.TestProcedure,
	framework scriptgen.Framework,
	storagePath string,
) {
	markFailed := func(reason error) {
		if updateErr := h.scriptStore.Update(ctx, scriptID,
			scriptgen.SetStatus(scriptgen.StatusFailed),
			scriptgen.SetErrorMessage(reason.Error()),
		); updateErr != nil {
			h.logger.Error(ctx, "failed to mark script as failed", map[string]interface{}{
				"error":     updateErr.Error(),
				"script_id": scriptID.String(),
			})
		}
	}

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "panic in background script generation", map[string]interface{}{
				"panic":     fmt.Sprintf("%v", r),
				"script_id": scriptID.String(),
			})
			markFailed(fmt.Errorf("internal panic: %v", r))
		}
	}()

	scriptContent, err := h.generator.Generate(ctx, procedure, framework)
	if err != nil {
		h.logger.Error(ctx, "background script generation failed", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		markFailed(err)
		return
	}

	reader := bytes.NewReader(scriptContent)
	if err := h.storage.Upload(ctx, storagePath, reader); err != nil {
		h.logger.Error(ctx, "failed to upload script to storage", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
			"path":      storagePath,
		})
		markFailed(err)
		return
	}

	if err := h.scriptStore.Update(ctx, scriptID,
		scriptgen.SetStatus(scriptgen.StatusCompleted),
		scriptgen.SetScriptPath(storagePath, int64(len(scriptContent))),
	); err != nil {
		h.logger.Error(ctx, "failed to mark script as completed", map[string]interface{}{
			"error":     err.Error(),
			"script_id": scriptID.String(),
		})
		// Best-effort cleanup so the orphaned file does not linger.
		if delErr := h.storage.Delete(ctx, storagePath); delErr != nil {
			h.logger.Warn(ctx, "failed to cleanup script after db update error", map[string]interface{}{
				"delete_error": delErr.Error(),
				"path":         storagePath,
			})
		}
		return
	}

	h.logger.Info(ctx, "script generated successfully", map[string]interface{}{
		"script_id": scriptID.String(),
		"file_size": len(scriptContent),
	})
}

// List handles listing all scripts for a test procedure.
func (h *ScriptGenHandler) List(w http.ResponseWriter, r *http.Request) {
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

	// Verify user owns the procedure's project
	if _, ok := h.verifyProcedureOwnership(w, ctx, procedureID, userID); !ok {
		// Helper already logged and responded with appropriate error
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

	respondJSON(w, http.StatusOK, NewPaginatedResponse(scripts, len(scripts), 0, 0))
}

// GetByID handles retrieving a script by its ID.
func (h *ScriptGenHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := GetUserID(ctx)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Extract script ID from URL
	scriptID, ok := parseUUIDOrRespond(w, r, "script_id", "script")
	if !ok {
		return
	}

	// Get script
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

	// Verify user owns the procedure's project
	if _, ok := h.verifyProcedureOwnership(w, ctx, script.TestProcedureID, userID); !ok {
		// Helper already logged and responded with appropriate error
		return
	}

	respondJSON(w, http.StatusOK, script)
}

// Download handles downloading a script file.
func (h *ScriptGenHandler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := GetUserID(ctx)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

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

	// Verify user owns the procedure's project
	if _, ok := h.verifyProcedureOwnership(w, ctx, script.TestProcedureID, userID); !ok {
		// Helper already logged and responded with appropriate error
		return
	}

	// Download from storage
	reader, err := h.storage.Download(ctx, script.ScriptPath)
	if err != nil {
		if errors.Is(err, storage.ErrFileNotFound) {
			respondError(w, http.StatusNotFound, "script file not found in storage")
			return
		}
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", script.FileName))

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

	// Get user ID from context
	userID, ok := GetUserID(ctx)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

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

	// Verify user owns the procedure's project
	if _, ok := h.verifyProcedureOwnership(w, ctx, script.TestProcedureID, userID); !ok {
		// Helper already logged and responded with appropriate error
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

// filenameSanitizer replaces characters that are problematic in filenames or storage paths.
var filenameSanitizer = strings.NewReplacer(
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

// sanitizeProcedureName removes or replaces characters that are problematic in filenames.
func sanitizeProcedureName(name string) string {
	// Remove control characters (\n, \r, \x00, etc.) to prevent them from
	// reaching the storage path or database file_name column.
	var stripped strings.Builder
	for _, r := range name {
		if !unicode.IsControl(r) {
			stripped.WriteRune(r)
		}
	}
	name = stripped.String()

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove or replace other problematic characters
	name = filenameSanitizer.Replace(name)

	// Limit length (truncate at rune boundary to avoid splitting multi-byte UTF-8 characters)
	if runes := []rune(name); len(runes) > 100 {
		name = string(runes[:100])
	}

	return name
}
