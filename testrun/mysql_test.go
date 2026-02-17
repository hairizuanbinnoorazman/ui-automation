package testrun

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLStore_Create(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully create test run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "Initial notes")
		err := store.Create(ctx, tr)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, tr.ID)
		assert.Equal(t, StatusPending, tr.Status)
	})

	t.Run("create test run with default status", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := &TestRun{
			TestProcedureID: testProcedureID,
			ExecutedBy:      executedBy,
		}
		err := store.Create(ctx, tr)
		require.NoError(t, err)
		assert.Equal(t, StatusPending, tr.Status)
	})

	t.Run("invalid test run returns error", func(t *testing.T) {
		executedBy := uuid.New()
		tr := &TestRun{
			ExecutedBy: executedBy,
			Status:     StatusPending,
		}
		err := store.Create(ctx, tr)
		assert.ErrorIs(t, err, ErrInvalidTestProcedureID)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing test run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "Test notes")
		require.NoError(t, store.Create(ctx, tr))

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, tr.ID, retrieved.ID)
		assert.Equal(t, tr.TestProcedureID, retrieved.TestProcedureID)
		assert.Equal(t, tr.Status, retrieved.Status)
	})

	t.Run("non-existent test run returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestRunNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("update notes", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "Original notes")
		require.NoError(t, store.Create(ctx, tr))

		err := store.Update(ctx, tr.ID, SetNotes("Updated notes"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated notes", retrieved.Notes)
	})

	t.Run("update status", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))

		err := store.Update(ctx, tr.ID, SetStatus(StatusRunning))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, retrieved.Status)
	})

	t.Run("update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, uuid.New(), SetNotes("New notes"))
		assert.ErrorIs(t, err, ErrTestRunNotFound)
	})
}

func TestMySQLStore_ListByTestProcedure(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("list test runs for procedure", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		// Create 3 runs
		for i := 0; i < 3; i++ {
			tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
			require.NoError(t, store.Create(ctx, tr))
		}

		runs, err := store.ListByTestProcedure(ctx, testProcedureID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(runs), 3)
	})

	t.Run("list with pagination", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		for i := 0; i < 5; i++ {
			tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
			require.NoError(t, store.Create(ctx, tr))
		}

		page1, err := store.ListByTestProcedure(ctx, testProcedureID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		page2, err := store.ListByTestProcedure(ctx, testProcedureID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})
}

func TestMySQLStore_Start(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully start test run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))

		err := store.Start(ctx, tr.ID)
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusRunning, retrieved.Status)
		assert.NotNil(t, retrieved.StartedAt)
	})

	t.Run("cannot start already started run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))
		require.NoError(t, store.Start(ctx, tr.ID))

		err := store.Start(ctx, tr.ID)
		assert.ErrorIs(t, err, ErrTestRunAlreadyStarted)
	})

	t.Run("start non-existent returns error", func(t *testing.T) {
		err := store.Start(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrTestRunNotFound)
	})
}

func TestMySQLStore_Complete(t *testing.T) {
	_, store, _ := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully complete with passed", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))
		require.NoError(t, store.Start(ctx, tr.ID))

		err := store.Complete(ctx, tr.ID, StatusPassed, "All tests passed")
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusPassed, retrieved.Status)
		assert.NotNil(t, retrieved.CompletedAt)
		assert.Equal(t, "All tests passed", retrieved.Notes)
	})

	t.Run("successfully complete with failed", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))
		require.NoError(t, store.Start(ctx, tr.ID))

		err := store.Complete(ctx, tr.ID, StatusFailed, "Failed at step 3")
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, StatusFailed, retrieved.Status)
		assert.Equal(t, "Failed at step 3", retrieved.Notes)
	})

	t.Run("cannot complete non-running run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusPending, "")
		require.NoError(t, store.Create(ctx, tr))

		err := store.Complete(ctx, tr.ID, StatusPassed, "")
		assert.ErrorIs(t, err, ErrTestRunNotRunning)
	})

	t.Run("complete non-existent returns error", func(t *testing.T) {
		err := store.Complete(ctx, uuid.New(), StatusPassed, "")
		assert.ErrorIs(t, err, ErrTestRunNotFound)
	})
}

func TestMySQLAssetStore_Create(t *testing.T) {
	_, store, assetStore := setupTestStore(t)
	ctx := context.Background()

	// Create a test run first
	testProcedureID := uuid.New()
	executedBy := uuid.New()
	tr := createTestRun(testProcedureID, executedBy, StatusRunning, "")
	require.NoError(t, store.Create(ctx, tr))

	t.Run("successfully create asset", func(t *testing.T) {
		asset := createTestAsset(tr.ID, AssetTypeImage, "path/to/image.png", "image.png", 1024)
		err := assetStore.Create(ctx, asset)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, asset.ID)
	})

	t.Run("create multiple assets for same run", func(t *testing.T) {
		asset1 := createTestAsset(tr.ID, AssetTypeImage, "path/to/screenshot1.png", "screenshot1.png", 2048)
		err := assetStore.Create(ctx, asset1)
		require.NoError(t, err)

		asset2 := createTestAsset(tr.ID, AssetTypeVideo, "path/to/video.mp4", "video.mp4", 1048576)
		err = assetStore.Create(ctx, asset2)
		require.NoError(t, err)

		assert.NotEqual(t, asset1.ID, asset2.ID)
	})

	t.Run("invalid asset returns error", func(t *testing.T) {
		asset := &TestRunAsset{
			TestRunID: tr.ID,
			AssetType: AssetType("invalid"),
			AssetPath: "path",
			FileName:  "file",
		}
		err := assetStore.Create(ctx, asset)
		assert.ErrorIs(t, err, ErrInvalidAssetType)
	})
}

func TestMySQLAssetStore_GetByID(t *testing.T) {
	_, store, assetStore := setupTestStore(t)
	ctx := context.Background()

	testProcedureID := uuid.New()
	executedBy := uuid.New()
	tr := createTestRun(testProcedureID, executedBy, StatusRunning, "")
	require.NoError(t, store.Create(ctx, tr))

	t.Run("retrieve existing asset", func(t *testing.T) {
		asset := createTestAsset(tr.ID, AssetTypeImage, "path/to/file.png", "file.png", 1024)
		require.NoError(t, assetStore.Create(ctx, asset))

		retrieved, err := assetStore.GetByID(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, asset.ID, retrieved.ID)
		assert.Equal(t, asset.FileName, retrieved.FileName)
		assert.Equal(t, asset.AssetType, retrieved.AssetType)
	})

	t.Run("non-existent asset returns error", func(t *testing.T) {
		_, err := assetStore.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})
}

func TestMySQLAssetStore_ListByTestRun(t *testing.T) {
	_, store, assetStore := setupTestStore(t)
	ctx := context.Background()

	t.Run("list assets for test run", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusRunning, "")
		require.NoError(t, store.Create(ctx, tr))

		// Create 3 assets
		for i := 0; i < 3; i++ {
			asset := createTestAsset(tr.ID, AssetTypeImage, "path", "file", 100)
			require.NoError(t, assetStore.Create(ctx, asset))
		}

		assets, err := assetStore.ListByTestRun(ctx, tr.ID)
		require.NoError(t, err)
		assert.Len(t, assets, 3)
	})

	t.Run("list returns empty for run with no assets", func(t *testing.T) {
		testProcedureID := uuid.New()
		executedBy := uuid.New()
		tr := createTestRun(testProcedureID, executedBy, StatusRunning, "")
		require.NoError(t, store.Create(ctx, tr))

		assets, err := assetStore.ListByTestRun(ctx, tr.ID)
		require.NoError(t, err)
		assert.Empty(t, assets)
	})
}

func TestMySQLAssetStore_Delete(t *testing.T) {
	_, store, assetStore := setupTestStore(t)
	ctx := context.Background()

	testProcedureID := uuid.New()
	executedBy := uuid.New()
	tr := createTestRun(testProcedureID, executedBy, StatusRunning, "")
	require.NoError(t, store.Create(ctx, tr))

	t.Run("delete existing asset", func(t *testing.T) {
		asset := createTestAsset(tr.ID, AssetTypeImage, "path", "file.png", 1024)
		require.NoError(t, assetStore.Create(ctx, asset))

		err := assetStore.Delete(ctx, asset.ID)
		require.NoError(t, err)

		_, err = assetStore.GetByID(ctx, asset.ID)
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		err := assetStore.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})
}
