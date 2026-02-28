package agent

import (
	"context"

	"github.com/hairizuanbinnoorazman/ui-automation/job"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
)

// WorkerPool manages a pool of goroutines that process jobs from the database.
// Workers are notified via a channel when new jobs are created, and each worker
// atomically claims jobs using SELECT FOR UPDATE to prevent double-processing.
type WorkerPool struct {
	Work       chan struct{}
	maxWorkers int
	jobStore   job.Store
	pipeline   *Pipeline
	logger     logger.Logger
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(maxWorkers int, jobStore job.Store, pipeline *Pipeline, log logger.Logger) *WorkerPool {
	return &WorkerPool{
		Work:       make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
		jobStore:   jobStore,
		pipeline:   pipeline,
		logger:     log,
	}
}

// Start spawns worker goroutines that listen for job notifications.
func (p *WorkerPool) Start(ctx context.Context) {
	p.logger.Info(ctx, "starting worker pool", map[string]interface{}{
		"max_workers": p.maxWorkers,
	})
	for i := 0; i < p.maxWorkers; i++ {
		go p.worker(ctx, i)
	}
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	p.logger.Info(ctx, "worker started", map[string]interface{}{
		"worker_id": id,
	})
	for {
		select {
		case <-p.Work:
			// Drain all available created jobs before going back to wait
			for {
				j, err := p.jobStore.ClaimNextCreated(ctx)
				if err != nil {
					p.logger.Error(ctx, "worker failed to claim job", map[string]interface{}{
						"worker_id": id,
						"error":     err.Error(),
					})
					break
				}
				if j == nil {
					break
				}
				p.logger.Info(ctx, "worker processing job", map[string]interface{}{
					"worker_id": id,
					"job_id":    j.ID.String(),
				})
				p.pipeline.RunAfterClaim(ctx, j.ID)
			}
		case <-ctx.Done():
			p.logger.Info(ctx, "worker stopping", map[string]interface{}{
				"worker_id": id,
			})
			return
		}
	}
}
