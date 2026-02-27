package job

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLStore_Create(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully create job", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
			Config:    JSONMap{"endpoint_id": uuid.New().String()},
		}
		err := store.Create(ctx, j)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, j.ID)
		assert.Equal(t, StatusCreated, j.Status)
		assert.Equal(t, JobTypeUIExploration, j.Type)
	})

	t.Run("create job without config", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		err := store.Create(ctx, j)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, j.ID)
	})

	t.Run("invalid job type returns error", func(t *testing.T) {
		j := &Job{
			Type:      JobType("invalid"),
			CreatedBy: uuid.New(),
		}
		err := store.Create(ctx, j)
		assert.ErrorIs(t, err, ErrInvalidJobType)
	})

	t.Run("missing created_by returns error", func(t *testing.T) {
		j := &Job{
			Type: JobTypeUIExploration,
		}
		err := store.Create(ctx, j)
		assert.ErrorIs(t, err, ErrInvalidCreatedBy)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing job", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
			Config:    JSONMap{"key": "value"},
		}
		require.NoError(t, store.Create(ctx, j))

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, j.ID, retrieved.ID)
		assert.Equal(t, j.Type, retrieved.Type)
		assert.Equal(t, StatusCreated, retrieved.Status)
		assert.Equal(t, j.CreatedBy, retrieved.CreatedBy)
	})

	t.Run("non-existent job returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrJobNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update config", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
			Config:    JSONMap{"key": "original"},
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Update(ctx, j.ID, SetConfig(JSONMap{"key": "updated"}))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated", retrieved.Config["key"])
	})

	t.Run("update result", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Update(ctx, j.ID, SetResult(JSONMap{"output": "data"}))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, "data", retrieved.Result["output"])
	})

	t.Run("update status", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Update(ctx, j.ID, SetStatus(StatusRunning))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, retrieved.Status)
	})

	t.Run("update with invalid status returns error", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Update(ctx, j.ID, SetStatus(Status("invalid")))
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, uuid.New(), SetConfig(JSONMap{}))
		assert.ErrorIs(t, err, ErrJobNotFound)
	})
}

func TestMySQLStore_ListByCreator(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list jobs for creator", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 3; i++ {
			j := &Job{
				Type:      JobTypeUIExploration,
				CreatedBy: createdBy,
			}
			require.NoError(t, store.Create(ctx, j))
		}

		jobs, err := store.ListByCreator(ctx, createdBy, 10, 0)
		require.NoError(t, err)
		assert.Len(t, jobs, 3)
	})

	t.Run("list with pagination", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 5; i++ {
			j := &Job{
				Type:      JobTypeUIExploration,
				CreatedBy: createdBy,
			}
			require.NoError(t, store.Create(ctx, j))
		}

		page1, err := store.ListByCreator(ctx, createdBy, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		page2, err := store.ListByCreator(ctx, createdBy, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("list returns empty for no jobs", func(t *testing.T) {
		jobs, err := store.ListByCreator(ctx, uuid.New(), 10, 0)
		require.NoError(t, err)
		assert.Empty(t, jobs)
	})

	t.Run("list does not return other users jobs", func(t *testing.T) {
		user1 := uuid.New()
		user2 := uuid.New()

		j1 := &Job{Type: JobTypeUIExploration, CreatedBy: user1}
		j2 := &Job{Type: JobTypeUIExploration, CreatedBy: user2}
		require.NoError(t, store.Create(ctx, j1))
		require.NoError(t, store.Create(ctx, j2))

		jobs, err := store.ListByCreator(ctx, user1, 10, 0)
		require.NoError(t, err)
		for _, j := range jobs {
			assert.Equal(t, user1, j.CreatedBy)
		}
	})
}

func TestMySQLStore_CountByCreator(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("count jobs for creator", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 3; i++ {
			j := &Job{
				Type:      JobTypeUIExploration,
				CreatedBy: createdBy,
			}
			require.NoError(t, store.Create(ctx, j))
		}

		count, err := store.CountByCreator(ctx, createdBy)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("count returns zero for no jobs", func(t *testing.T) {
		count, err := store.CountByCreator(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestMySQLStore_ListByType(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list jobs by type", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 3; i++ {
			j := &Job{
				Type:      JobTypeUIExploration,
				CreatedBy: createdBy,
			}
			require.NoError(t, store.Create(ctx, j))
		}

		jobs, err := store.ListByType(ctx, JobTypeUIExploration, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 3)
		for _, j := range jobs {
			assert.Equal(t, JobTypeUIExploration, j.Type)
		}
	})

	t.Run("list by type with pagination", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 5; i++ {
			j := &Job{
				Type:      JobTypeUIExploration,
				CreatedBy: createdBy,
			}
			require.NoError(t, store.Create(ctx, j))
		}

		page1, err := store.ListByType(ctx, JobTypeUIExploration, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)
	})
}

func TestMySQLStore_Start(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("start a created job", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Start(ctx, j.ID)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, retrieved.Status)
		assert.NotNil(t, retrieved.StartTime)
	})

	t.Run("start already running job returns error", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		require.NoError(t, store.Start(ctx, j.ID))

		err := store.Start(ctx, j.ID)
		assert.ErrorIs(t, err, ErrJobAlreadyStarted)
	})

	t.Run("start non-existent job returns error", func(t *testing.T) {
		err := store.Start(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrJobNotFound)
	})
}

func TestMySQLStore_Complete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("complete running job with success", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		require.NoError(t, store.Start(ctx, j.ID))

		result := JSONMap{"pages_found": float64(5)}
		err := store.Complete(ctx, j.ID, StatusSuccess, result)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusSuccess, retrieved.Status)
		assert.NotNil(t, retrieved.EndTime)
		assert.NotNil(t, retrieved.Duration)
		assert.Equal(t, float64(5), retrieved.Result["pages_found"])
	})

	t.Run("complete running job with failure", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		require.NoError(t, store.Start(ctx, j.ID))

		result := JSONMap{"error": "connection timeout"}
		err := store.Complete(ctx, j.ID, StatusFailed, result)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusFailed, retrieved.Status)
		assert.NotNil(t, retrieved.EndTime)
	})

	t.Run("complete running job with stopped", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		require.NoError(t, store.Start(ctx, j.ID))

		result := JSONMap{"reason": "stopped by user"}
		err := store.Complete(ctx, j.ID, StatusStopped, result)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, j.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusStopped, retrieved.Status)
	})

	t.Run("complete non-running job returns error", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		err := store.Complete(ctx, j.ID, StatusSuccess, nil)
		assert.ErrorIs(t, err, ErrJobNotRunning)
	})

	t.Run("complete non-existent job returns error", func(t *testing.T) {
		err := store.Complete(ctx, uuid.New(), StatusSuccess, nil)
		assert.ErrorIs(t, err, ErrJobNotFound)
	})

	t.Run("complete already completed job returns error", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		require.NoError(t, store.Start(ctx, j.ID))
		require.NoError(t, store.Complete(ctx, j.ID, StatusSuccess, nil))

		err := store.Complete(ctx, j.ID, StatusFailed, nil)
		assert.ErrorIs(t, err, ErrJobNotRunning)
	})
}

func TestMySQLStore_StatusTransitions(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("created to running to success", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))
		assert.Equal(t, StatusCreated, j.Status)

		require.NoError(t, store.Start(ctx, j.ID))
		started, _ := store.GetByID(ctx, j.ID)
		assert.Equal(t, StatusRunning, started.Status)

		require.NoError(t, store.Complete(ctx, j.ID, StatusSuccess, JSONMap{"result": "ok"}))
		completed, _ := store.GetByID(ctx, j.ID)
		assert.Equal(t, StatusSuccess, completed.Status)
	})

	t.Run("created to running to failed", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		require.NoError(t, store.Start(ctx, j.ID))
		require.NoError(t, store.Complete(ctx, j.ID, StatusFailed, JSONMap{"error": "timeout"}))

		completed, _ := store.GetByID(ctx, j.ID)
		assert.Equal(t, StatusFailed, completed.Status)
	})

	t.Run("created to running to stopped", func(t *testing.T) {
		j := &Job{
			Type:      JobTypeUIExploration,
			CreatedBy: uuid.New(),
		}
		require.NoError(t, store.Create(ctx, j))

		require.NoError(t, store.Start(ctx, j.ID))
		require.NoError(t, store.Complete(ctx, j.ID, StatusStopped, JSONMap{"reason": "user requested"}))

		completed, _ := store.GetByID(ctx, j.ID)
		assert.Equal(t, StatusStopped, completed.Status)
	})
}
