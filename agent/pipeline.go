package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/endpoint"
	"github.com/hairizuan-noorazman/ui-automation/job"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/storage"
	"github.com/hairizuan-noorazman/ui-automation/testprocedure"
)

// Pipeline orchestrates UI exploration by spawning a Python agent subprocess.
type Pipeline struct {
	config             Config
	jobStore           job.Store
	endpointStore      endpoint.Store
	testProcedureStore testprocedure.Store
	storage            storage.BlobStorage
	logger             logger.Logger
	cancelFuncs        sync.Map // map[uuid.UUID]context.CancelFunc
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

	// Create context with timeout and store cancel func
	ctx, cancel := context.WithTimeout(ctx, p.config.TimeLimit)
	defer cancel()
	p.cancelFuncs.Store(jobID, cancel)
	defer p.cancelFuncs.Delete(jobID)

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

	// 4. Create temp directory for this job
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("agent-job-%s", jobID.String()))
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tmpDir)

	// 5. Build agent config
	creds := make([]Credential, len(ep.Credentials))
	for i, c := range ep.Credentials {
		creds[i] = Credential{Key: c.Key, Value: c.Value}
	}

	agentCfg := AgentConfig{
		TargetURL:        ep.URL,
		Credentials:      creds,
		ProcedureName:    procedureName,
		JobID:            jobID.String(),
		OutputDir:        tmpDir,
		PlaywrightMCPURL: p.config.PlaywrightMCPURL + "/sse",
	}

	configJSON, err := json.Marshal(agentCfg)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to marshal agent config: %v", err))
		return
	}

	// 6. Spawn Python agent subprocess
	p.logger.Info(ctx, "spawning agent subprocess", map[string]interface{}{
		"job_id":      jobID.String(),
		"script_path": p.config.AgentScriptPath,
		"target_url":  ep.URL,
	})

	cmd := exec.CommandContext(ctx, "python3", p.config.AgentScriptPath)
	cmd.Stdin = bytes.NewReader(configJSON)

	// Set environment variables for Bedrock auth
	cmd.Env = append(os.Environ(),
		"CLAUDE_CODE_USE_BEDROCK=1",
		fmt.Sprintf("AWS_REGION=%s", p.config.BedrockRegion),
	)
	if p.config.BedrockAccessKey != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", p.config.BedrockAccessKey))
	}
	if p.config.BedrockSecretKey != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", p.config.BedrockSecretKey))
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("agent subprocess failed: %v; stderr: %s", err, stderr.String()))
		return
	}

	// 7. Read result from output file
	resultPath := filepath.Join(tmpDir, "result.json")
	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to read agent result: %v", err))
		return
	}

	var agentResult AgentResult
	if err := json.Unmarshal(resultData, &agentResult); err != nil {
		p.failJob(ctx, jobID, fmt.Sprintf("failed to parse agent result: %v", err))
		return
	}

	// 8. Upload screenshots to storage and build test procedure steps
	steps := make(testprocedure.Steps, 0, len(agentResult.Steps))
	for _, step := range agentResult.Steps {
		storedPaths := make([]string, 0, len(step.ImagePaths))
		for _, imgPath := range step.ImagePaths {
			localPath := filepath.Join(tmpDir, imgPath)
			if _, err := os.Stat(localPath); err != nil {
				p.logger.Warn(ctx, "screenshot file not found, skipping", map[string]interface{}{
					"path": localPath,
				})
				continue
			}

			storagePath := fmt.Sprintf("test-procedures/%s/%s", projectID.String(), filepath.Base(imgPath))
			f, err := os.Open(localPath)
			if err != nil {
				p.logger.Warn(ctx, "failed to open screenshot, skipping", map[string]interface{}{
					"path":  localPath,
					"error": err.Error(),
				})
				continue
			}
			if err := p.storage.Upload(ctx, storagePath, f); err != nil {
				f.Close()
				p.logger.Warn(ctx, "failed to upload screenshot, skipping", map[string]interface{}{
					"path":  storagePath,
					"error": err.Error(),
				})
				continue
			}
			f.Close()

			url, err := p.storage.GetURL(ctx, storagePath)
			if err != nil {
				storedPaths = append(storedPaths, storagePath)
			} else {
				storedPaths = append(storedPaths, url)
			}
		}

		steps = append(steps, testprocedure.TestStep{
			Name:         step.Name,
			Instructions: step.Instructions,
			ImagePaths:   storedPaths,
		})
	}

	// If no steps were generated, create a placeholder
	if len(steps) == 0 {
		steps = append(steps, testprocedure.TestStep{
			Name:         "Initial observation",
			Instructions: agentResult.Summary,
			ImagePaths:   []string{},
		})
	}

	// 9. Save procedure
	tp := &testprocedure.TestProcedure{
		ProjectID:   projectID,
		Name:        agentResult.ProcedureName,
		Description: "Auto-generated from UI exploration: " + agentResult.Description,
		Steps:       steps,
		CreatedBy:   j.CreatedBy,
	}

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

// Stop cancels a running job's agent subprocess.
func (p *Pipeline) Stop(jobID uuid.UUID) {
	if cancelFn, ok := p.cancelFuncs.Load(jobID); ok {
		cancelFn.(context.CancelFunc)()
	}
}

// failJob marks a job as failed with the given reason.
func (p *Pipeline) failJob(ctx context.Context, jobID uuid.UUID, reason string) {
	p.logger.Error(ctx, "agent pipeline failed", map[string]interface{}{
		"job_id": jobID.String(),
		"reason": reason,
	})

	// Truncate long error messages for storage
	if len(reason) > 1000 {
		reason = reason[:1000] + "... (truncated)"
	}

	if err := p.jobStore.Complete(ctx, jobID, job.StatusFailed, job.JSONMap{
		"error": reason,
	}); err != nil {
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
