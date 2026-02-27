package job

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	Update(ctx context.Context, id uuid.UUID, setters ...UpdateSetter) error
	ListByCreator(ctx context.Context, createdBy uuid.UUID, limit, offset int) ([]*Job, error)
	CountByCreator(ctx context.Context, createdBy uuid.UUID) (int, error)
	ListByType(ctx context.Context, jobType JobType, limit, offset int) ([]*Job, error)
	Start(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID, status Status, result JSONMap) error
}

type UpdateSetter func(*Job) error
