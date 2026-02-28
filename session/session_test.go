package session

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Hour),
			want:      true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, session.IsExpired())
		})
	}
}

func TestStore_SetAndGet(t *testing.T) {
	store := NewStore()

	sessionID := uuid.New()
	userID := uuid.New()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	store.Set(session)

	retrieved, err := store.Get(sessionID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, session.UserID, retrieved.UserID)
	assert.Equal(t, session.Email, retrieved.Email)
}

func TestStore_GetNonExistent(t *testing.T) {
	store := NewStore()

	_, err := store.Get(uuid.New())
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_GetExpired(t *testing.T) {
	store := NewStore()

	sessionID := uuid.New()
	userID := uuid.New()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     "test@example.com",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	store.Set(session)

	_, err := store.Get(sessionID)
	assert.ErrorIs(t, err, ErrSessionExpired)
}

func TestStore_Delete(t *testing.T) {
	store := NewStore()

	sessionID := uuid.New()
	userID := uuid.New()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	store.Set(session)
	store.Delete(sessionID)

	_, err := store.Get(sessionID)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestStore_Cleanup(t *testing.T) {
	store := NewStore()

	// Add active session
	activeSessionID := uuid.New()
	activeUserID := uuid.New()
	activeSession := &Session{
		ID:        activeSessionID,
		UserID:    activeUserID,
		Email:     "active@example.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	store.Set(activeSession)

	// Add expired session
	expiredSessionID := uuid.New()
	expiredUserID := uuid.New()
	expiredSession := &Session{
		ID:        expiredSessionID,
		UserID:    expiredUserID,
		Email:     "expired@example.com",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	store.Set(expiredSession)

	// Cleanup should remove only expired session
	removed := store.Cleanup()
	assert.Equal(t, 1, removed)

	// Active session should still exist
	_, err := store.Get(activeSessionID)
	assert.NoError(t, err)

	// Expired session should be removed
	_, err = store.Get(expiredSessionID)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestManager_Create(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(24*time.Hour, log)

	userID := uuid.New()
	session, err := manager.Create(userID, "test@example.com")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, "test@example.com", session.Email)
	assert.False(t, session.IsExpired())
}

func TestManager_Get(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(24*time.Hour, log)

	userID := uuid.New()
	created, err := manager.Create(userID, "test@example.com")
	require.NoError(t, err)

	retrieved, err := manager.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.UserID, retrieved.UserID)
}

func TestManager_GetExpired(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(time.Millisecond, log)

	userID := uuid.New()
	created, err := manager.Create(userID, "test@example.com")
	require.NoError(t, err)

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	_, err = manager.Get(created.ID)
	assert.ErrorIs(t, err, ErrSessionExpired)
}

func TestManager_Delete(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(24*time.Hour, log)

	userID := uuid.New()
	created, err := manager.Create(userID, "test@example.com")
	require.NoError(t, err)

	manager.Delete(created.ID)

	_, err = manager.Get(created.ID)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestManager_Cleanup(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(50*time.Millisecond, log)

	// Create session that will expire soon
	userID1 := uuid.New()
	_, err := manager.Create(userID1, "test@example.com")
	require.NoError(t, err)

	// Create another active session with longer duration
	manager2 := NewManager(24*time.Hour, log)
	manager2.store = manager.store // Share store
	userID2 := uuid.New()
	activeSession, err := manager2.Create(userID2, "active@example.com")
	require.NoError(t, err)

	// Wait for first session to expire
	time.Sleep(100 * time.Millisecond)

	// Manual cleanup
	removed := manager.store.Cleanup()
	assert.Equal(t, 1, removed)

	// Active session should still be retrievable
	_, err = manager.Get(activeSession.ID)
	assert.NoError(t, err)
}

func TestManager_Concurrent(t *testing.T) {
	log := logger.NewTestLogger()
	manager := NewManager(24*time.Hour, log)

	var wg sync.WaitGroup
	sessionIDs := make(chan uuid.UUID, 100)

	// Create 100 sessions concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := uuid.New()
			session, err := manager.Create(userID, "test@example.com")
			if err == nil {
				sessionIDs <- session.ID
			}
		}(i)
	}

	wg.Wait()
	close(sessionIDs)

	// Verify all sessions can be retrieved
	count := 0
	for sessionID := range sessionIDs {
		_, err := manager.Get(sessionID)
		assert.NoError(t, err)
		count++
	}

	assert.Equal(t, 100, count)
}

func TestGenerateSessionID(t *testing.T) {
	id1 := uuid.New()
	assert.NotEqual(t, uuid.Nil, id1)

	id2 := uuid.New()
	assert.NotEqual(t, uuid.Nil, id2)

	// IDs should be unique
	assert.NotEqual(t, id1, id2)
}
