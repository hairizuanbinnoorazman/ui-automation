package job

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed job store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// Create creates a new job in the database.
func (s *MySQLStore) Create(ctx context.Context, j *Job) error {
	if err := j.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(j).Error; err != nil {
		s.logger.Error(ctx, "failed to create job", map[string]interface{}{
			"error": err.Error(),
			"type":  string(j.Type),
		})
		return err
	}

	s.logger.Info(ctx, "job created", map[string]interface{}{
		"job_id": j.ID.String(),
		"type":   string(j.Type),
	})

	return nil
}

// GetByID retrieves a job by its ID.
func (s *MySQLStore) GetByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	var j Job
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&j).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		s.logger.Error(ctx, "failed to get job by ID", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id.String(),
		})
		return nil, err
	}

	return &j, nil
}

// Update updates a job with the given setters.
func (s *MySQLStore) Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error {
	j, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	for _, setter := range setters {
		if err := setter(j); err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Save(j).Error; err != nil {
		s.logger.Error(ctx, "failed to update job", map[string]interface{}{
			"error":  err.Error(),
			"job_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "job updated", map[string]interface{}{
		"job_id": id.String(),
	})

	return nil
}

// ListByCreator retrieves a paginated list of jobs created by a specific user.
func (s *MySQLStore) ListByCreator(ctx context.Context, createdBy uuid.UUID, limit, offset int) ([]*Job, error) {
	var jobs []*Job
	err := s.db.WithContext(ctx).
		Where("created_by = ?", createdBy).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list jobs by creator", map[string]interface{}{
			"error":      err.Error(),
			"created_by": createdBy.String(),
			"limit":      limit,
			"offset":     offset,
		})
		return nil, err
	}

	return jobs, nil
}

// CountByCreator returns the total count of jobs created by a specific user.
func (s *MySQLStore) CountByCreator(ctx context.Context, createdBy uuid.UUID) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&Job{}).
		Where("created_by = ?", createdBy).
		Count(&count).Error

	if err != nil {
		s.logger.Error(ctx, "failed to count jobs by creator", map[string]interface{}{
			"error":      err.Error(),
			"created_by": createdBy.String(),
		})
		return 0, err
	}

	return int(count), nil
}

// ListByType retrieves a paginated list of jobs filtered by type.
func (s *MySQLStore) ListByType(ctx context.Context, jobType JobType, limit, offset int) ([]*Job, error) {
	var jobs []*Job
	err := s.db.WithContext(ctx).
		Where("type = ?", jobType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&jobs).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list jobs by type", map[string]interface{}{
			"error":  err.Error(),
			"type":   string(jobType),
			"limit":  limit,
			"offset": offset,
		})
		return nil, err
	}

	return jobs, nil
}

// Start marks a job as running.
func (s *MySQLStore) Start(ctx context.Context, id uuid.UUID) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var j Job
		if err := tx.WithContext(ctx).Where("id = ?", id).First(&j).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJobNotFound
			}
			return err
		}

		if err := j.Start(); err != nil {
			return err
		}

		return tx.WithContext(ctx).Save(&j).Error
	})

	if err != nil {
		if !errors.Is(err, ErrJobNotFound) && !errors.Is(err, ErrJobAlreadyStarted) {
			s.logger.Error(ctx, "failed to start job", map[string]interface{}{
				"error":  err.Error(),
				"job_id": id.String(),
			})
		}
		return err
	}

	s.logger.Info(ctx, "job started", map[string]interface{}{
		"job_id": id.String(),
	})

	return nil
}

// Complete marks a job as finished with the given status and result.
func (s *MySQLStore) Complete(ctx context.Context, id uuid.UUID, status Status, result JSONMap) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var j Job
		if err := tx.WithContext(ctx).Where("id = ?", id).First(&j).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrJobNotFound
			}
			return err
		}

		if err := j.Complete(status, result); err != nil {
			return err
		}

		return tx.WithContext(ctx).Save(&j).Error
	})

	if err != nil {
		if !errors.Is(err, ErrJobNotFound) && !errors.Is(err, ErrJobNotRunning) {
			s.logger.Error(ctx, "failed to complete job", map[string]interface{}{
				"error":  err.Error(),
				"job_id": id.String(),
				"status": string(status),
			})
		}
		return err
	}

	s.logger.Info(ctx, "job completed", map[string]interface{}{
		"job_id": id.String(),
		"status": string(status),
	})

	return nil
}
