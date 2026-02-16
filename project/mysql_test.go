package project

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

	t.Run("successfully create project", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Test Project", "Test Description", ownerID)
		err := store.Create(ctx, project)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, project.ID)
		assert.NotZero(t, project.CreatedAt)
	})

	t.Run("create project without description", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Minimal Project", "", ownerID)
		err := store.Create(ctx, project)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, project.ID)
	})

	t.Run("invalid project returns error", func(t *testing.T) {
		ownerID := uuid.New()
		project := &Project{
			Description: "Missing name",
			OwnerID:     ownerID,
		}
		err := store.Create(ctx, project)
		assert.ErrorIs(t, err, ErrInvalidProjectName)
	})

	t.Run("missing owner_id returns error", func(t *testing.T) {
		project := &Project{
			Name: "Test Project",
		}
		err := store.Create(ctx, project)
		assert.ErrorIs(t, err, ErrInvalidOwner)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing project", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Get Test Project", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project))

		retrieved, err := store.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, project.ID, retrieved.ID)
		assert.Equal(t, project.Name, retrieved.Name)
		assert.Equal(t, project.Description, retrieved.Description)
		assert.Equal(t, project.OwnerID, retrieved.OwnerID)
		assert.True(t, retrieved.IsActive)
	})

	t.Run("non-existent project returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("soft-deleted project not found", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Deleted Project", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project))
		require.NoError(t, store.Delete(ctx, project.ID))

		_, err := store.GetByID(ctx, project.ID)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update single field", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Original Name", "Original Description", ownerID)
		require.NoError(t, store.Create(ctx, project))

		err := store.Update(ctx, project.ID, SetName("Updated Name"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "Original Description", retrieved.Description)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Original Name", "Original Description", ownerID)
		require.NoError(t, store.Create(ctx, project))

		err := store.Update(ctx, project.ID,
			SetName("New Name"),
			SetDescription("New Description"),
		)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", retrieved.Name)
		assert.Equal(t, "New Description", retrieved.Description)
	})

	t.Run("update non-existent project returns error", func(t *testing.T) {
		err := store.Update(ctx, uuid.New(), SetName("New Name"))
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("update with invalid name returns error", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Valid Name", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project))

		err := store.Update(ctx, project.ID, SetName(""))
		assert.ErrorIs(t, err, ErrInvalidProjectName)
	})
}

func TestMySQLStore_Delete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("delete existing project", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("To Delete", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project))

		err := store.Delete(ctx, project.ID)
		require.NoError(t, err)

		// Verify project cannot be retrieved
		_, err = store.GetByID(ctx, project.ID)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("delete non-existent project returns error", func(t *testing.T) {
		err := store.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("delete already deleted project returns error", func(t *testing.T) {
		ownerID := uuid.New()
		project := createTestProject("Already Deleted", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project))
		require.NoError(t, store.Delete(ctx, project.ID))

		err := store.Delete(ctx, project.ID)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}

func TestMySQLStore_ListByOwner(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list projects for owner with multiple projects", func(t *testing.T) {
		ownerID := uuid.New()
		// Create 3 projects for owner 1
		for i := 0; i < 3; i++ {
			project := createTestProject("Project "+string(rune('A'+i)), "Description", ownerID)
			require.NoError(t, store.Create(ctx, project))
		}

		projects, err := store.ListByOwner(ctx, ownerID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(projects), 3)
	})

	t.Run("list projects returns only owner's projects", func(t *testing.T) {
		owner1 := uuid.New()
		owner2 := uuid.New()

		project1 := createTestProject("Owner 1 Project", "Description", owner1)
		require.NoError(t, store.Create(ctx, project1))

		project2 := createTestProject("Owner 2 Project", "Description", owner2)
		require.NoError(t, store.Create(ctx, project2))

		projects, err := store.ListByOwner(ctx, owner1, 10, 0)
		require.NoError(t, err)

		// All projects should belong to owner1
		for _, p := range projects {
			assert.Equal(t, owner1, p.OwnerID)
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		ownerID := uuid.New()
		// Create 5 projects
		for i := 0; i < 5; i++ {
			project := createTestProject("Paginated Project "+string(rune('A'+i)), "Description", ownerID)
			require.NoError(t, store.Create(ctx, project))
		}

		// Get first page (2 items)
		page1, err := store.ListByOwner(ctx, ownerID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		// Get second page (2 items)
		page2, err := store.ListByOwner(ctx, ownerID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		// Pages should contain different projects
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("list excludes soft-deleted projects", func(t *testing.T) {
		ownerID := uuid.New()
		project1 := createTestProject("Active Project", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project1))

		project2 := createTestProject("Deleted Project", "Description", ownerID)
		require.NoError(t, store.Create(ctx, project2))
		require.NoError(t, store.Delete(ctx, project2.ID))

		projects, err := store.ListByOwner(ctx, ownerID, 10, 0)
		require.NoError(t, err)

		// Should only contain active project
		for _, p := range projects {
			assert.NotEqual(t, project2.ID, p.ID)
		}
	})

	t.Run("list returns empty for owner with no projects", func(t *testing.T) {
		projects, err := store.ListByOwner(ctx, uuid.New(), 10, 0)
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}
