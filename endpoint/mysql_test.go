package endpoint

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

	t.Run("successfully create endpoint with credentials", func(t *testing.T) {
		createdBy := uuid.New()
		creds := Credentials{
			{Key: "api_key", Value: "secret123"},
		}
		ep := createTestEndpoint("Test Endpoint", "https://example.com", createdBy, creds)
		err := store.Create(ctx, ep)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, ep.ID)
		assert.Equal(t, "Test Endpoint", ep.Name)
		assert.Equal(t, "https://example.com", ep.URL)
		assert.Len(t, ep.Credentials, 1)
		assert.Equal(t, "api_key", ep.Credentials[0].Key)
	})

	t.Run("successfully create endpoint without credentials defaults", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("No Creds Endpoint", "https://example.com/api", createdBy, nil)
		err := store.Create(ctx, ep)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, ep.ID)
		// Should have default credentials
		assert.Len(t, ep.Credentials, 2)
		assert.Equal(t, "username", ep.Credentials[0].Key)
		assert.Equal(t, "email", ep.Credentials[1].Key)
	})

	t.Run("empty credentials gets defaults", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Empty Creds", "https://example.com", createdBy, Credentials{})
		err := store.Create(ctx, ep)
		require.NoError(t, err)
		assert.Len(t, ep.Credentials, 2)
	})

	t.Run("missing name returns error", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("", "https://example.com", createdBy, nil)
		err := store.Create(ctx, ep)
		assert.ErrorIs(t, err, ErrInvalidEndpointName)
	})

	t.Run("missing URL returns error", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Test", "", createdBy, nil)
		err := store.Create(ctx, ep)
		assert.ErrorIs(t, err, ErrInvalidEndpointURL)
	})

	t.Run("missing created_by returns error", func(t *testing.T) {
		ep := createTestEndpoint("Test", "https://example.com", uuid.Nil, nil)
		err := store.Create(ctx, ep)
		assert.ErrorIs(t, err, ErrInvalidCreatedBy)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing endpoint", func(t *testing.T) {
		createdBy := uuid.New()
		creds := Credentials{
			{Key: "token", Value: "abc123"},
		}
		ep := createTestEndpoint("Get Test", "https://example.com", createdBy, creds)
		require.NoError(t, store.Create(ctx, ep))

		retrieved, err := store.GetByID(ctx, ep.ID)
		require.NoError(t, err)
		assert.Equal(t, ep.ID, retrieved.ID)
		assert.Equal(t, ep.Name, retrieved.Name)
		assert.Equal(t, ep.URL, retrieved.URL)
		assert.Equal(t, ep.CreatedBy, retrieved.CreatedBy)
		assert.Len(t, retrieved.Credentials, 1)
		assert.Equal(t, "token", retrieved.Credentials[0].Key)
	})

	t.Run("non-existent endpoint returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrEndpointNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update name", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Original Name", "https://example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Update(ctx, ep.ID, SetName("Updated Name"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, ep.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
	})

	t.Run("update URL", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Test", "https://old.example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Update(ctx, ep.ID, SetURL("https://new.example.com"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, ep.ID)
		require.NoError(t, err)
		assert.Equal(t, "https://new.example.com", retrieved.URL)
	})

	t.Run("update credentials", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Test", "https://example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		newCreds := Credentials{
			{Key: "api_key", Value: "new_secret"},
			{Key: "api_secret", Value: "more_secret"},
		}
		err := store.Update(ctx, ep.ID, SetCredentials(newCreds))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, ep.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.Credentials, 2)
		assert.Equal(t, "api_key", retrieved.Credentials[0].Key)
		assert.Equal(t, "new_secret", retrieved.Credentials[0].Value)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Old", "https://old.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Update(ctx, ep.ID, SetName("New"), SetURL("https://new.com"))
		require.NoError(t, err)

		retrieved, err := store.GetByID(ctx, ep.ID)
		require.NoError(t, err)
		assert.Equal(t, "New", retrieved.Name)
		assert.Equal(t, "https://new.com", retrieved.URL)
	})

	t.Run("update with empty name returns error", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Test", "https://example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Update(ctx, ep.ID, SetName(""))
		assert.ErrorIs(t, err, ErrInvalidEndpointName)
	})

	t.Run("update with empty URL returns error", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("Test", "https://example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Update(ctx, ep.ID, SetURL(""))
		assert.ErrorIs(t, err, ErrInvalidEndpointURL)
	})

	t.Run("update non-existent returns error", func(t *testing.T) {
		err := store.Update(ctx, uuid.New(), SetName("New Name"))
		assert.ErrorIs(t, err, ErrEndpointNotFound)
	})
}

func TestMySQLStore_Delete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("delete existing endpoint", func(t *testing.T) {
		createdBy := uuid.New()
		ep := createTestEndpoint("To Delete", "https://example.com", createdBy, nil)
		require.NoError(t, store.Create(ctx, ep))

		err := store.Delete(ctx, ep.ID)
		require.NoError(t, err)

		_, err = store.GetByID(ctx, ep.ID)
		assert.ErrorIs(t, err, ErrEndpointNotFound)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		err := store.Delete(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrEndpointNotFound)
	})
}

func TestMySQLStore_ListByCreator(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("list endpoints for creator", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 3; i++ {
			ep := createTestEndpoint("Endpoint "+string(rune('A'+i)), "https://example.com/"+string(rune('a'+i)), createdBy, nil)
			require.NoError(t, store.Create(ctx, ep))
		}

		endpoints, err := store.ListByCreator(ctx, createdBy, 10, 0)
		require.NoError(t, err)
		assert.Len(t, endpoints, 3)
	})

	t.Run("list returns only creator's endpoints", func(t *testing.T) {
		creator1 := uuid.New()
		creator2 := uuid.New()

		ep1 := createTestEndpoint("Creator1 Endpoint", "https://c1.example.com", creator1, nil)
		require.NoError(t, store.Create(ctx, ep1))

		ep2 := createTestEndpoint("Creator2 Endpoint", "https://c2.example.com", creator2, nil)
		require.NoError(t, store.Create(ctx, ep2))

		endpoints, err := store.ListByCreator(ctx, creator1, 10, 0)
		require.NoError(t, err)
		assert.Len(t, endpoints, 1)
		assert.Equal(t, "Creator1 Endpoint", endpoints[0].Name)
	})

	t.Run("list with pagination", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 5; i++ {
			ep := createTestEndpoint("Paginated "+string(rune('A'+i)), "https://example.com/"+string(rune('a'+i)), createdBy, nil)
			require.NoError(t, store.Create(ctx, ep))
		}

		page1, err := store.ListByCreator(ctx, createdBy, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		page2, err := store.ListByCreator(ctx, createdBy, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})

	t.Run("list for creator with no endpoints", func(t *testing.T) {
		endpoints, err := store.ListByCreator(ctx, uuid.New(), 10, 0)
		require.NoError(t, err)
		assert.Len(t, endpoints, 0)
	})
}

func TestMySQLStore_CountByCreator(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("count endpoints for creator", func(t *testing.T) {
		createdBy := uuid.New()
		for i := 0; i < 3; i++ {
			ep := createTestEndpoint("Count "+string(rune('A'+i)), "https://example.com/"+string(rune('a'+i)), createdBy, nil)
			require.NoError(t, store.Create(ctx, ep))
		}

		count, err := store.CountByCreator(ctx, createdBy)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("count for creator with no endpoints", func(t *testing.T) {
		count, err := store.CountByCreator(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
