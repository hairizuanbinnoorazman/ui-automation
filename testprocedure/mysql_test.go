package testprocedure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLStore_Create(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully create test procedure", func(t *testing.T) {
		steps := Steps{
			{"action": "click", "selector": "#button"},
		}
		tp := createTestProcedure("Test Procedure", "Description", 1, 1, steps)
		err := store.Create(ctx, tp)
		require.NoError(t, err)
		assert.NotZero(t, tp.ID)
		assert.Equal(t, uint(1), tp.Version)
		assert.True(t, tp.IsLatest)
		assert.Nil(t, tp.ParentID)
	})

	t.Run("create test procedure without steps", func(t *testing.T) {
		tp := createTestProcedure("Minimal Procedure", "Description", 1, 1, nil)
		err := store.Create(ctx, tp)
		require.NoError(t, err)
		assert.NotZero(t, tp.ID)
	})

	t.Run("invalid test procedure returns error", func(t *testing.T) {
		tp := &TestProcedure{
			Description: "Missing name",
			ProjectID:   1,
			CreatedBy:   1,
		}
		err := store.Create(ctx, tp)
		assert.ErrorIs(t, err, ErrInvalidTestProcedureName)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing test procedure", func(t *testing.T) {
		steps := Steps{
			{"action": "navigate", "url": "https://example.com"},
		}
		tp := createTestProcedure("Get Test", "Description", 1, 1, steps)
		require.NoError(t, store.Create(ctx, tp))

		retrieved, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, tp.ID, retrieved.ID)
		assert.Equal(t, tp.Name, retrieved.Name)
		assert.Equal(t, tp.ProjectID, retrieved.ProjectID)
		assert.Equal(t, len(steps), len(retrieved.Steps))
	})

	t.Run("non-existent test procedure returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update name", func(t *testing.T) {
		tp := createTestProcedure("Original Name", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, tp))

		err := store.Update(ctx, tp.ID, SetName("Updated Name"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
	})

	t.Run("update steps", func(t *testing.T) {
		tp := createTestProcedure("Test", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, tp))

		newSteps := Steps{
			{"action": "click", "selector": "#new-button"},
		}
		err := store.Update(ctx, tp.ID, SetSteps(newSteps))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, len(newSteps), len(retrieved.Steps))
		assert.Equal(t, "click", retrieved.Steps[0]["action"])
	})

	t.Run("update multiple fields", func(t *testing.T) {
		tp := createTestProcedure("Original", "Original Desc", 1, 1, nil)
		require.NoError(t, store.Create(ctx, tp))

		err := store.Update(ctx, tp.ID,
			SetName("New Name"),
			SetDescription("New Description"),
		)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", retrieved.Name)
		assert.Equal(t, "New Description", retrieved.Description)
	})

	t.Run("update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, 99999, SetName("New Name"))
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_Delete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("delete existing test procedure", func(t *testing.T) {
		tp := createTestProcedure("To Delete", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, tp))

		err := store.Delete(ctx, tp.ID)
		require.NoError(t, err)

		_, err = store.GetByID(ctx, tp.ID)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		err := store.Delete(ctx, 99999)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_ListByProject(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list procedures for project", func(t *testing.T) {
		projectID := uint(10)
		// Create 3 procedures for project 10
		for i := 0; i < 3; i++ {
			tp := createTestProcedure("Procedure "+string(rune('A'+i)), "Description", projectID, 1, nil)
			require.NoError(t, store.Create(ctx, tp))
		}

		procedures, err := store.ListByProject(ctx, projectID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(procedures), 3)
	})

	t.Run("list returns only latest versions", func(t *testing.T) {
		projectID := uint(20)
		tp := createTestProcedure("Versioned Procedure", "Description", projectID, 1, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Create a new version
		_, err := store.CreateVersion(ctx, tp.ID)
		require.NoError(t, err)

		procedures, err := store.ListByProject(ctx, projectID, 10, 0)
		require.NoError(t, err)

		// Should only return one procedure (the latest version)
		count := 0
		for _, p := range procedures {
			if p.Name == "Versioned Procedure" {
				count++
				assert.True(t, p.IsLatest)
				assert.Equal(t, uint(2), p.Version)
			}
		}
		assert.Equal(t, 1, count)
	})

	t.Run("list with pagination", func(t *testing.T) {
		projectID := uint(30)
		for i := 0; i < 5; i++ {
			tp := createTestProcedure("Paginated "+string(rune('A'+i)), "Description", projectID, 1, nil)
			require.NoError(t, store.Create(ctx, tp))
		}

		page1, err := store.ListByProject(ctx, projectID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		page2, err := store.ListByProject(ctx, projectID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})
}

func TestMySQLStore_CreateVersion(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("create version from original", func(t *testing.T) {
		original := createTestProcedure("Original Procedure", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		// Create version
		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)
		assert.NotEqual(t, original.ID, version2.ID)
		assert.Equal(t, uint(2), version2.Version)
		assert.True(t, version2.IsLatest)
		assert.NotNil(t, version2.ParentID)
		assert.Equal(t, original.ID, *version2.ParentID)

		// Original should no longer be latest
		originalRetrieved, err := store.GetByID(ctx, original.ID)
		require.NoError(t, err)
		assert.False(t, originalRetrieved.IsLatest)
		assert.Equal(t, uint(1), originalRetrieved.Version)
	})

	t.Run("create version from version", func(t *testing.T) {
		original := createTestProcedure("Test", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)

		// Create version from version2
		version3, err := store.CreateVersion(ctx, version2.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(3), version3.Version)
		assert.True(t, version3.IsLatest)
		assert.NotNil(t, version3.ParentID)
		assert.Equal(t, original.ID, *version3.ParentID)

		// Both original and version2 should not be latest
		originalRetrieved, _ := store.GetByID(ctx, original.ID)
		assert.False(t, originalRetrieved.IsLatest)

		version2Retrieved, _ := store.GetByID(ctx, version2.ID)
		assert.False(t, version2Retrieved.IsLatest)
	})

	t.Run("version preserves content", func(t *testing.T) {
		steps := Steps{
			{"action": "click", "selector": "#button"},
		}
		original := createTestProcedure("Original", "Original Description", 1, 1, steps)
		require.NoError(t, store.Create(ctx, original))

		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)
		assert.Equal(t, original.Name, version2.Name)
		assert.Equal(t, original.Description, version2.Description)
		assert.Equal(t, len(original.Steps), len(version2.Steps))
		assert.Equal(t, original.ProjectID, version2.ProjectID)
		assert.Equal(t, original.CreatedBy, version2.CreatedBy)
	})

	t.Run("create version from non-existent returns error", func(t *testing.T) {
		_, err := store.CreateVersion(ctx, 99999)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_GetVersionHistory(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("get history of original", func(t *testing.T) {
		original := createTestProcedure("Versioned", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		history, err := store.GetVersionHistory(ctx, original.ID)
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, original.ID, history[0].ID)
		assert.Equal(t, uint(1), history[0].Version)
	})

	t.Run("get history with multiple versions", func(t *testing.T) {
		original := createTestProcedure("Multi-Version", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)

		version3, err := store.CreateVersion(ctx, version2.ID)
		require.NoError(t, err)

		// Get history from any version
		history, err := store.GetVersionHistory(ctx, version3.ID)
		require.NoError(t, err)
		assert.Len(t, history, 3)

		// Should be ordered by version DESC
		assert.Equal(t, uint(3), history[0].Version)
		assert.Equal(t, uint(2), history[1].Version)
		assert.Equal(t, uint(1), history[2].Version)

		// All should have same parent_id (pointing to original)
		assert.NotNil(t, history[0].ParentID)
		assert.Equal(t, original.ID, *history[0].ParentID)
		assert.NotNil(t, history[1].ParentID)
		assert.Equal(t, original.ID, *history[1].ParentID)
	})

	t.Run("get history from middle version", func(t *testing.T) {
		original := createTestProcedure("Test History", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)

		version3, err := store.CreateVersion(ctx, version2.ID)
		require.NoError(t, err)

		// Get history from version2
		history, err := store.GetVersionHistory(ctx, version2.ID)
		require.NoError(t, err)
		assert.Len(t, history, 3)

		// Should still get all versions
		versions := make(map[uint]bool)
		for _, v := range history {
			versions[v.Version] = true
		}
		assert.True(t, versions[1])
		assert.True(t, versions[2])
		assert.True(t, versions[3])

		_ = version3 // Use variable
	})

	t.Run("get history from non-existent returns error", func(t *testing.T) {
		_, err := store.GetVersionHistory(ctx, 99999)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_VersioningComplexScenario(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update vs version behavior", func(t *testing.T) {
		// Create original
		original := createTestProcedure("Original", "Description", 1, 1, nil)
		require.NoError(t, store.Create(ctx, original))

		// Update modifies in-place (no new version)
		err := store.Update(ctx, original.ID, SetName("Updated Name"))
		require.NoError(t, err)

		updated, err := store.GetByID(ctx, original.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, uint(1), updated.Version) // Version unchanged
		assert.True(t, updated.IsLatest)

		// Create version creates new immutable copy
		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", version2.Name) // Preserves updated name
		assert.Equal(t, uint(2), version2.Version)

		// Original is no longer latest
		originalAgain, err := store.GetByID(ctx, original.ID)
		require.NoError(t, err)
		assert.False(t, originalAgain.IsLatest)
	})
}
