package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLStore_Create(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("successfully create user", func(t *testing.T) {
		user := createTestUser("test@example.com", "testuser", "password123")
		err := store.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		user1 := createTestUser("duplicate@example.com", "user1", "password123")
		require.NoError(t, store.Create(ctx, user1))

		user2 := createTestUser("duplicate@example.com", "user2", "password123")
		err := store.Create(ctx, user2)
		assert.ErrorIs(t, err, ErrDuplicateEmail)
	})

	t.Run("invalid user returns error", func(t *testing.T) {
		user := &User{
			Username: "testuser",
			// Missing email
		}
		err := store.Create(ctx, user)
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})
}

func TestMySQLStore_GetByID(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve existing user", func(t *testing.T) {
		user := createTestUser("get@example.com", "getuser", "password123")
		require.NoError(t, store.Create(ctx, user))

		retrieved, err := store.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.Equal(t, user.Username, retrieved.Username)
	})

	t.Run("non-existent user returns error", func(t *testing.T) {
		_, err := store.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("soft-deleted user not found", func(t *testing.T) {
		user := createTestUser("deleted@example.com", "deleteduser", "password123")
		require.NoError(t, store.Create(ctx, user))
		require.NoError(t, store.Delete(ctx, user.ID))

		_, err := store.GetByID(ctx, user.ID)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMySQLStore_GetByEmail(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("retrieve user by email", func(t *testing.T) {
		user := createTestUser("email@example.com", "emailuser", "password123")
		require.NoError(t, store.Create(ctx, user))

		retrieved, err := store.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("non-existent email returns error", func(t *testing.T) {
		_, err := store.GetByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMySQLStore_Update(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("update single field", func(t *testing.T) {
		user := createTestUser("update1@example.com", "updateuser1", "password123")
		require.NoError(t, store.Create(ctx, user))

		err := store.Update(ctx, user.ID, SetUsername("newusername"))
		require.NoError(t, err)

		updated, err := store.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "newusername", updated.Username)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		user := createTestUser("update2@example.com", "updateuser2", "password123")
		require.NoError(t, store.Create(ctx, user))

		err := store.Update(ctx, user.ID,
			SetUsername("newusername2"),
			SetEmail("newemail@example.com"),
		)
		require.NoError(t, err)

		updated, err := store.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "newusername2", updated.Username)
		assert.Equal(t, "newemail@example.com", updated.Email)
	})

	t.Run("update password", func(t *testing.T) {
		user := createTestUser("update3@example.com", "updateuser3", "password123")
		require.NoError(t, store.Create(ctx, user))
		oldHash := user.PasswordHash

		err := store.Update(ctx, user.ID, SetPassword("newpassword123"))
		require.NoError(t, err)

		updated, err := store.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotEqual(t, oldHash, updated.PasswordHash)
		assert.True(t, updated.CheckPassword("newpassword123"))
	})

	t.Run("invalid setter returns error", func(t *testing.T) {
		user := createTestUser("update4@example.com", "updateuser4", "password123")
		require.NoError(t, store.Create(ctx, user))

		err := store.Update(ctx, user.ID, SetEmail(""))
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})

	t.Run("non-existent user returns error", func(t *testing.T) {
		err := store.Update(ctx, 99999, SetUsername("test"))
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMySQLStore_Delete(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	t.Run("soft delete user", func(t *testing.T) {
		user := createTestUser("delete1@example.com", "deleteuser1", "password123")
		require.NoError(t, store.Create(ctx, user))

		err := store.Delete(ctx, user.ID)
		require.NoError(t, err)

		_, err = store.GetByID(ctx, user.ID)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("non-existent user returns error", func(t *testing.T) {
		err := store.Delete(ctx, 99999)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("deleting already deleted user returns error", func(t *testing.T) {
		user := createTestUser("delete2@example.com", "deleteuser2", "password123")
		require.NoError(t, store.Create(ctx, user))
		require.NoError(t, store.Delete(ctx, user.ID))

		err := store.Delete(ctx, user.ID)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMySQLStore_List(t *testing.T) {
	_, store := setupTestStore(t)
	ctx := context.Background()

	// Create 5 test users
	for i := 1; i <= 5; i++ {
		user := createTestUser(
			"list"+string(rune('0'+i))+"@example.com",
			"listuser"+string(rune('0'+i)),
			"password123",
		)
		require.NoError(t, store.Create(ctx, user))
	}

	t.Run("list with pagination - first page", func(t *testing.T) {
		users, err := store.List(ctx, 3, 0)
		require.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("list with pagination - second page", func(t *testing.T) {
		users, err := store.List(ctx, 3, 3)
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("only active users returned", func(t *testing.T) {
		// Delete one user
		users, err := store.List(ctx, 10, 0)
		require.NoError(t, err)
		initialCount := len(users)

		if initialCount > 0 {
			require.NoError(t, store.Delete(ctx, users[0].ID))

			users, err = store.List(ctx, 10, 0)
			require.NoError(t, err)
			assert.Equal(t, initialCount-1, len(users))
		}
	})
}
