package testprocedure

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

	t.Run("successfully create test procedure with draft", func(t *testing.T) {
		steps := Steps{
			{Name: "Step 1", Instructions: "Click the button", ImagePaths: []string{}},
		}
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test Procedure", "Description", projectID, createdBy, steps)
		err := store.Create(ctx, tp)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tp.ID)
		assert.Equal(t, uint(1), tp.Version)
		assert.True(t, tp.IsLatest)
		assert.Nil(t, tp.ParentID)

		// Verify draft (v0) was also created
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(0), draft.Version)
		assert.False(t, draft.IsLatest)
		assert.NotNil(t, draft.ParentID)
		assert.Equal(t, tp.ID, *draft.ParentID)
	})

	t.Run("create test procedure without steps", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Minimal Procedure", "Description", projectID, createdBy, nil)
		err := store.Create(ctx, tp)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tp.ID)
	})

	t.Run("invalid test procedure returns error", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := &TestProcedure{
			Description: "Missing name",
			ProjectID:   projectID,
			CreatedBy:   createdBy,
		}
		err := store.Create(ctx, tp)
		assert.ErrorIs(t, err, ErrInvalidTestProcedureName)
	})

	t.Run("step without name returns error", func(t *testing.T) {
		steps := Steps{
			{Name: "", Instructions: "No name", ImagePaths: []string{}},
		}
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, steps)
		err := store.Create(ctx, tp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step")
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing test procedure", func(t *testing.T) {
		steps := Steps{
			{Name: "Navigate", Instructions: "Go to example.com", ImagePaths: []string{"path/to/image.png"}},
		}
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Get Test", "Description", projectID, createdBy, steps)
		require.NoError(t, store.Create(ctx, tp))

		retrieved, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, tp.ID, retrieved.ID)
		assert.Equal(t, tp.Name, retrieved.Name)
		assert.Equal(t, tp.ProjectID, retrieved.ProjectID)
		assert.Equal(t, len(steps), len(retrieved.Steps))
	})

	t.Run("non-existent test procedure returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update name", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Original Name", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Note: Update now updates the draft (v0), not the committed version
		err := store.Update(ctx, tp.ID, SetName("Updated Name"))
		require.NoError(t, err)

		// Verify draft was updated
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", draft.Name)

		// Verify committed version is unchanged
		committed, err := store.GetLatestCommitted(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", committed.Name)
	})

	t.Run("update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, uuid.New(), SetName("New Name"))
		assert.Error(t, err)
	})
}

func TestMySQLStore_Delete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("delete existing test procedure", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("To Delete", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		err := store.Delete(ctx, tp.ID)
		require.NoError(t, err)

		_, err = store.GetByID(ctx, tp.ID)
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		err := store.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_ListByProject(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list procedures for project", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		// Create 3 procedures for project
		for i := 0; i < 3; i++ {
			tp := createTestProcedure("Procedure "+string(rune('A'+i)), "Description", projectID, createdBy, nil)
			require.NoError(t, store.Create(ctx, tp))
		}

		procedures, err := store.ListByProject(ctx, projectID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(procedures), 3)

		// Verify no drafts are included (version 0)
		for _, p := range procedures {
			assert.NotEqual(t, uint(0), p.Version)
		}
	})

	t.Run("list returns only latest versions", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Versioned Procedure", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Commit draft to create version 2
		_, err := store.CommitDraft(ctx, tp.ID)
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
		projectID := uuid.New()
		createdBy := uuid.New()
		for i := 0; i < 5; i++ {
			tp := createTestProcedure("Paginated "+string(rune('A'+i)), "Description", projectID, createdBy, nil)
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
		projectID := uuid.New()
		createdBy := uuid.New()
		original := createTestProcedure("Original Procedure", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, original))

		// Create version (legacy method)
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
		projectID := uuid.New()
		createdBy := uuid.New()
		original := createTestProcedure("Test", "Description", projectID, createdBy, nil)
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
			{Name: "Click", Instructions: "Click the button", ImagePaths: []string{"image.png"}},
		}
		projectID := uuid.New()
		createdBy := uuid.New()
		original := createTestProcedure("Original", "Original Description", projectID, createdBy, steps)
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
		_, err := store.CreateVersion(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

func TestMySQLStore_GetVersionHistory(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("get history excludes draft", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		original := createTestProcedure("Versioned", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, original))

		history, err := store.GetVersionHistory(ctx, original.ID)
		require.NoError(t, err)

		// Should include v1 and v0 (draft)
		assert.GreaterOrEqual(t, len(history), 2)

		// Verify versions
		hasV1 := false
		hasV0 := false
		for _, v := range history {
			if v.Version == 1 {
				hasV1 = true
			}
			if v.Version == 0 {
				hasV0 = true
			}
		}
		assert.True(t, hasV1)
		assert.True(t, hasV0)
	})

	t.Run("get history with multiple versions", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		original := createTestProcedure("Multi-Version", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, original))

		version2, err := store.CreateVersion(ctx, original.ID)
		require.NoError(t, err)

		version3, err := store.CreateVersion(ctx, version2.ID)
		require.NoError(t, err)

		// Get history from any version
		history, err := store.GetVersionHistory(ctx, version3.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(history), 4) // v0, v1, v2, v3

		// Find max version to verify ordering
		maxVersion := uint(0)
		for _, v := range history {
			if v.Version > maxVersion {
				maxVersion = v.Version
			}
		}
		assert.Equal(t, uint(3), maxVersion)
	})

	t.Run("get history from non-existent returns error", func(t *testing.T) {
		_, err := store.GetVersionHistory(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestProcedureNotFound)
	})
}

// Draft workflow tests
func TestMySQLStore_GetDraft(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("get draft after creation", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(0), draft.Version)
		assert.False(t, draft.IsLatest)
		assert.NotNil(t, draft.ParentID)
		assert.Equal(t, tp.ID, *draft.ParentID)
	})

	t.Run("get draft for non-existent returns error", func(t *testing.T) {
		_, err := store.GetDraft(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestMySQLStore_GetLatestCommitted(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("get latest committed after creation", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		committed, err := store.GetLatestCommitted(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(1), committed.Version)
		assert.True(t, committed.IsLatest)
	})

	t.Run("get latest committed after multiple versions", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Create versions
		v2, err := store.CommitDraft(ctx, tp.ID)
		require.NoError(t, err)

		_, err = store.CommitDraft(ctx, v2.ID)
		require.NoError(t, err)

		// Get latest should be v3
		latest, err := store.GetLatestCommitted(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(3), latest.Version)
		assert.True(t, latest.IsLatest)
	})
}

func TestMySQLStore_UpdateDraft(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update draft name", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Original", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Update draft
		err := store.UpdateDraft(ctx, tp.ID, SetName("Draft Name"))
		require.NoError(t, err)

		// Draft should be updated
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Draft Name", draft.Name)

		// Committed version should be unchanged
		committed, err := store.GetLatestCommitted(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Original", committed.Name)
	})

	t.Run("update draft steps", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		newSteps := Steps{
			{Name: "New Step", Instructions: "Do something", ImagePaths: []string{}},
		}
		err := store.UpdateDraft(ctx, tp.ID, SetSteps(newSteps))
		require.NoError(t, err)

		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Len(t, draft.Steps, 1)
		assert.Equal(t, "New Step", draft.Steps[0].Name)
	})

	t.Run("update draft for non-existent returns error", func(t *testing.T) {
		err := store.UpdateDraft(ctx, uuid.New(), SetName("Test"))
		assert.Error(t, err)
	})
}

func TestMySQLStore_ResetDraft(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("reset draft to committed", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Original", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Modify draft
		err := store.UpdateDraft(ctx, tp.ID, SetName("Modified Draft"))
		require.NoError(t, err)

		// Reset draft
		err = store.ResetDraft(ctx, tp.ID)
		require.NoError(t, err)

		// Draft should match committed
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Original", draft.Name)
	})

	t.Run("reset draft with no committed version returns error", func(t *testing.T) {
		err := store.ResetDraft(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestMySQLStore_CommitDraft(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("commit draft creates new version", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Original", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// Modify draft
		err := store.UpdateDraft(ctx, tp.ID, SetName("Modified"))
		require.NoError(t, err)

		// Commit draft
		v2, err := store.CommitDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(2), v2.Version)
		assert.True(t, v2.IsLatest)
		assert.Equal(t, "Modified", v2.Name)

		// Original should not be latest
		v1, err := store.GetByID(ctx, tp.ID)
		require.NoError(t, err)
		assert.False(t, v1.IsLatest)

		// Draft should still exist and remain unchanged
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(0), draft.Version)
		assert.Equal(t, "Modified", draft.Name)
	})

	t.Run("commit draft multiple times", func(t *testing.T) {
		projectID := uuid.New()
		createdBy := uuid.New()
		tp := createTestProcedure("Test", "Description", projectID, createdBy, nil)
		require.NoError(t, store.Create(ctx, tp))

		// First commit
		v2, err := store.CommitDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(2), v2.Version)

		// Modify draft again
		err = store.UpdateDraft(ctx, v2.ID, SetName("Second Modification"))
		require.NoError(t, err)

		// Second commit
		v3, err := store.CommitDraft(ctx, v2.ID)
		require.NoError(t, err)
		assert.Equal(t, uint(3), v3.Version)
		assert.Equal(t, "Second Modification", v3.Name)
	})

	t.Run("commit draft for non-existent returns error", func(t *testing.T) {
		_, err := store.CommitDraft(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestMySQLStore_CompleteWorkflow(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("complete draft workflow", func(t *testing.T) {
		// Create procedure (creates v1 and v0 draft)
		projectID := uuid.New()
		createdBy := uuid.New()
		steps1 := Steps{
			{Name: "Step 1", Instructions: "Original step", ImagePaths: []string{}},
		}
		tp := createTestProcedure("Test Procedure", "Description", projectID, createdBy, steps1)
		require.NoError(t, store.Create(ctx, tp))

		// Edit draft
		steps2 := Steps{
			{Name: "Step 1", Instructions: "Modified step", ImagePaths: []string{"image.png"}},
			{Name: "Step 2", Instructions: "New step", ImagePaths: []string{}},
		}
		err := store.UpdateDraft(ctx, tp.ID,
			SetName("Modified Procedure"),
			SetSteps(steps2),
		)
		require.NoError(t, err)

		// Verify draft has changes
		draft, err := store.GetDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Modified Procedure", draft.Name)
		assert.Len(t, draft.Steps, 2)

		// Verify committed is unchanged
		committed, err := store.GetLatestCommitted(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test Procedure", committed.Name)
		assert.Len(t, committed.Steps, 1)

		// Commit draft
		v2, err := store.CommitDraft(ctx, tp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Modified Procedure", v2.Name)
		assert.Len(t, v2.Steps, 2)
		assert.Equal(t, uint(2), v2.Version)

		// List should only show latest version
		procedures, err := store.ListByProject(ctx, projectID, 10, 0)
		require.NoError(t, err)
		procedureCount := 0
		for _, p := range procedures {
			if p.Name == "Modified Procedure" {
				procedureCount++
				assert.Equal(t, uint(2), p.Version)
			}
		}
		assert.Equal(t, 1, procedureCount)

		// Reset draft back to v2
		err = store.UpdateDraft(ctx, v2.ID, SetName("Draft Change"))
		require.NoError(t, err)

		err = store.ResetDraft(ctx, v2.ID)
		require.NoError(t, err)

		draft, err = store.GetDraft(ctx, v2.ID)
		require.NoError(t, err)
		assert.Equal(t, "Modified Procedure", draft.Name)
	})
}
