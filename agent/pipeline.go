package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/endpoint"
	"github.com/hairizuan-noorazman/ui-automation/job"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// Pipeline orchestrates the three-agent exploration pipeline.
type Pipeline struct {
	config             Config
	jobStore           job.Store
	endpointStore      endpoint.Store
	testProcedureStore testprocedure.Store
	storage            storage.BlobStorage
	logger             logger.Logger
}

// NewPipeline creates a new agent pipeline.
func NewPipeline(
	config Config,
	jobStore job.Store,
	endpointStore endpoint.Store,
	testProcedureStore testprocedure.Store,
	blobStorage storage.BlobStorage,
	log logger.Logger,
) *Pipeline {
	return &Pipeline{
		config:             config,
		jobStore:           jobStore,
		endpointStore:      endpointStore,
		testProcedureStore: testProcedureStore,
		storage:            blobStorage,
		logger:             log,
	}
}

// Run executes the full exploration pipeline for a given job.
func (p *Pipeline) Run(ctx context.Context, jobID uuid.UUID) {
	p.logger.Info(ctx, "starting agent pipeline", map[string]interface{}{
		"job_id": jobID.String(),
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.TimeLimit)
	defer cancel()

	// 1. Fetch job and parse config
	j, err := p.jobStore.GetByID(ctx, jobID)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to fetch job: %v", err))
		return
	}

	endpointIDStr, ok := j.Config["endpoint_id"].(string)
	if !ok {
		p.failJob(ctx, jobID, "missing endpoint_id in job config")
		return
	}
	endpointID, err := uuid.Parse(endpointIDStr)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("invalid endpoint_id: %v", err))
		return
	}

	projectIDStr, ok := j.Config["project_id"].(string)
	if !ok {
		p.failJob(ctx, jobID, "missing project_id in job config")
		return
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("invalid project_id: %v", err))
		return
	}

	procedureName, _ := j.Config["procedure_name"].(string)
	if procedureName == "" {
		procedureName = "UI Exploration"
	}

	// 2. Fetch endpoint
	ep, err := p.endpointStore.GetByID(ctx, endpointID)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to fetch endpoint: %v", err))
		return
	}

	// 3. Mark job as running
	if err := p.jobStore.Start(ctx, jobID); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to start job: %v", err))
		return
	}

	// 4. Connect MCP bridge
	bridge := NewMCPBridge(p.config.PlaywrightMCPURL, p.logger)
	if err := bridge.Connect(ctx); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to connect to Playwright MCP: %v", err))
		return
	}
	defer bridge.Close()

	// 5. Convert endpoint credentials
	creds := make([]Credential, len(ep.Credentials))
	for i, c := range ep.Credentials {
		creds[i] = Credential{Key: c.Key, Value: c.Value}
	}

	// 6. Run planner agent
	planner := NewPlanner(p.config, p.logger)
	plan, err := planner.Plan(ctx, ep.URL, creds)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("planner failed: %v", err))
		return
	}

	// 7. Run explorer agent
	explorer := NewExplorer(p.config, bridge, p.storage, p.logger)
	result, err := explorer.Explore(ctx, jobID.String(), plan)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("explorer failed: %v", err))
		return
	}

	// 8. Run documenter agent
	documenter := NewDocumenter(p.config, p.logger)
	tp, err := documenter.Document(ctx, procedureName, result)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("documenter failed: %v", err))
		return
	}

	// 9. Save procedure
	tp.ProjectID = projectID
	tp.CreatedBy = j.CreatedBy
	if err := p.testProcedureStore.Create(ctx, tp); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to save procedure: %v", err))
		return
	}

	// 10. Mark job success
	if err := p.jobStore.Complete(ctx, jobID, job.StatusSuccess, job.JSONMap{
		"procedure_id":   tp.ID.String(),
		"procedure_name": tp.Name,
		"steps_count":    len(tp.Steps),
	}); err != nil {
		p.logger.Error(ctx, "failed to mark job as success", map[string]interface{}{
			"error":  err.Error(),
			"job_id": jobID.String(),
		})
	}

	p.logger.Info(ctx, "agent pipeline completed successfully", map[string]interface{}{
		"job_id":       jobID.String(),
		"procedure_id": tp.ID.String(),
	})
}

// failJob marks a job as failed with the given reason.
func (p *Pipeline) failJob(ctx context.Context, jobID uuid.UUID, reason string) {
	p.logger.Error(ctx, "agent pipeline failed", map[string]interface{}{
		"job_id": jobID.String(),
		"reason": reason,
	})

	// Try to mark the job as failed
	if err := p.jobStore.Complete(ctx, jobID, job.StatusFailed, job.JSONMap{
		"error": reason,
	}); err != nil {
		// If the job hasn't been started yet, we need a different approach
		// Update the job status directly
		if err2 := p.jobStore.Update(ctx, jobID, job.SetStatus(job.StatusFailed), job.SetResult(job.JSONMap{
			"error": reason,
		})); err2 != nil {
			p.logger.Error(ctx, "failed to mark job as failed", map[string]interface{}{
				"error":  err2.Error(),
				"job_id": jobID.String(),
			})
		}
	}
}
