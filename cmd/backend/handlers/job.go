package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/agent"
	"github.com/hairizuan-noorazman/ui-automation/endpoint"
	"github.com/hairizuan-noorazman/ui-automation/job"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/project"
)

// JobHandler handles job-related requests.
type JobHandler struct {
	jobStore      job.Store
	endpointStore endpoint.Store
	projectStore  project.Store
	workerPool    *agent.WorkerPool
	pipeline      *agent.Pipeline
	logger        logger.Logger
}

// NewJobHandler creates a new job handler.
func NewJobHandler(jobStore job.Store, endpointStore endpoint.Store, projectStore project.Store, pool *agent.WorkerPool, pipeline *agent.Pipeline, log logger.Logger) *JobHandler {
	return &JobHandler{
		jobStore:      jobStore,
		endpointStore: endpointStore,
		projectStore:  projectStore,
		workerPool:    pool,
		pipeline:      pipeline,
		logger:        log,
	}
}

// checkJobOwnership verifies that the authenticated user created the job.
// Returns false if the check fails (response already written).
func (h *JobHandler) checkJobOwnership(w http.ResponseWriter, r *http.Request, jobID uuid.UUID) bool {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return false
	}

	j, err := h.jobStore.GetByID(r.Context(), jobID)
	if err != nil {
		if errors.Is(err, job.ErrJobNotFound) {
			respondError(w, http.StatusNotFound, "job not found")
			return false
		}
		h.logger.Error(r.Context(), "failed to get job for authorization", map[string]interface{}{
			"error":  err.Error(),
			"job_id": jobID,
		})
		respondError(w, http.StatusInternalServerError, "authorization check failed")
		return false
	}

	if j.CreatedBy != userID {
		h.logger.Warn(r.Context(), "unauthorized job access attempt", map[string]interface{}{
			"user_id":    userID,
			"job_id":     jobID,
			"created_by": j.CreatedBy,
		})
		respondError(w, http.StatusForbidden, "you don't have access to this job")
		return false
	}

	return true
}

// CreateJobRequest represents a job creation request.
type CreateJobRequest struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Create handles creating a new job.
func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req CreateJobRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	jobType := job.JobType(req.Type)
	if !jobType.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid job type")
		return
	}

	// For ui_exploration jobs, validate required config fields
	if jobType == job.JobTypeUIExploration {
		endpointIDStr, ok := req.Config["endpoint_id"].(string)
		if !ok || endpointIDStr == "" {
			respondError(w, http.StatusBadRequest, "endpoint_id is required in config for ui_exploration jobs")
			return
		}
		endpointID, err := uuid.Parse(endpointIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "endpoint_id must be a valid UUID")
			return
		}

		projectIDStr, ok := req.Config["project_id"].(string)
		if !ok || projectIDStr == "" {
			respondError(w, http.StatusBadRequest, "project_id is required in config for ui_exploration jobs")
			return
		}
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "project_id must be a valid UUID")
			return
		}

		// Verify user owns the endpoint
		ep, err := h.endpointStore.GetByID(r.Context(), endpointID)
		if err != nil {
			if errors.Is(err, endpoint.ErrEndpointNotFound) {
				respondError(w, http.StatusNotFound, "endpoint not found")
				return
			}
			h.logger.Error(r.Context(), "failed to verify endpoint", map[string]interface{}{
				"error":       err.Error(),
				"endpoint_id": endpointID,
			})
			respondError(w, http.StatusInternalServerError, "failed to verify endpoint")
			return
		}
		if ep.CreatedBy != userID {
			respondError(w, http.StatusForbidden, "you don't have access to this endpoint")
			return
		}

		// Verify user owns the project
		proj, err := h.projectStore.GetByID(r.Context(), projectID)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				respondError(w, http.StatusNotFound, "project not found")
				return
			}
			h.logger.Error(r.Context(), "failed to verify project", map[string]interface{}{
				"error":      err.Error(),
				"project_id": projectID,
			})
			respondError(w, http.StatusInternalServerError, "failed to verify project")
			return
		}
		if proj.OwnerID != userID {
			respondError(w, http.StatusForbidden, "you don't have access to this project")
			return
		}
	}

	j := &job.Job{
		Type:      jobType,
		Status:    job.StatusCreated,
		Config:    job.JSONMap(req.Config),
		CreatedBy: userID,
	}

	if err := h.jobStore.Create(r.Context(), j); err != nil {
		h.logger.Error(r.Context(), "failed to create job", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	// Notify worker pool that a new job is available
	if jobType == job.JobTypeUIExploration && h.workerPool != nil {
		select {
		case h.workerPool.Work <- struct{}{}:
		default:
			// All workers busy; job stays in DB as 'created' until a worker is free
		}
	}

	respondJSON(w, http.StatusCreated, j)
}

// List handles listing jobs for the authenticated user.
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.jobStore.CountByCreator(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to count jobs", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to count jobs")
		return
	}

	jobs, err := h.jobStore.ListByCreator(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list jobs", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	respondJSON(w, http.StatusOK, NewPaginatedResponse(jobs, total, limit, offset))
}

// GetByID handles getting a single job by ID.
func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDOrRespond(w, r, "id", "job")
	if !ok {
		return
	}

	if !h.checkJobOwnership(w, r, id) {
		return
	}

	j, err := h.jobStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, job.ErrJobNotFound) {
			respondError(w, http.StatusNotFound, "job not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get job", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get job")
		return
	}

	respondJSON(w, http.StatusOK, j)
}

// Stop handles stopping a running job.
func (h *JobHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUIDOrRespond(w, r, "id", "job")
	if !ok {
		return
	}

	if !h.checkJobOwnership(w, r, id) {
		return
	}

	// Check if job is running
	j, err := h.jobStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, job.ErrJobNotFound) {
			respondError(w, http.StatusNotFound, "job not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get job", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get job")
		return
	}

	if j.Status != job.StatusRunning {
		respondError(w, http.StatusBadRequest, "job is not running")
		return
	}

	// Cancel the agent subprocess if running
	if h.pipeline != nil {
		h.pipeline.Stop(id)
	}

	result := job.JSONMap{"reason": "stopped by user"}
	if err := h.jobStore.Complete(r.Context(), id, job.StatusStopped, result); err != nil {
		if errors.Is(err, job.ErrJobNotRunning) {
			respondError(w, http.StatusBadRequest, "job is not running")
			return
		}
		h.logger.Error(r.Context(), "failed to stop job", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to stop job")
		return
	}

	stopped, err := h.jobStore.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error(r.Context(), "failed to get stopped job", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id,
		})
		respondError(w, http.StatusInternalServerError, "failed to get stopped job")
		return
	}

	respondJSON(w, http.StatusOK, stopped)
}
