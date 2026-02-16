package session

import (
	"context"
	"time"

	"github.com/hairizuan-noorazman/ui-automation/logger"
)

// Manager manages user sessions with automatic cleanup.
type Manager struct {
	store    *Store
	duration time.Duration
	logger   logger.Logger
	stopCh   chan struct{}
}

// NewManager creates a new session manager with the given duration.
func NewManager(duration time.Duration, log logger.Logger) *Manager {
	return &Manager{
		store:    NewStore(),
		duration: duration,
		logger:   log,
		stopCh:   make(chan struct{}),
	}
}

// Create creates a new session for the given user.
func (m *Manager) Create(userID uint, email string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(m.duration),
	}

	m.store.Set(session)

	m.logger.Info(context.Background(), "session created", map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"email":      email,
	})

	return session, nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(sessionID string) (*Session, error) {
	return m.store.Get(sessionID)
}

// Delete deletes a session by ID.
func (m *Manager) Delete(sessionID string) {
	m.store.Delete(sessionID)
	m.logger.Info(context.Background(), "session deleted", map[string]interface{}{
		"session_id": sessionID,
	})
}

// StartCleanup starts a background goroutine that periodically cleans up expired sessions.
func (m *Manager) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				removed := m.store.Cleanup()
				if removed > 0 {
					m.logger.Info(context.Background(), "cleaned up expired sessions", map[string]interface{}{
						"removed_count": removed,
					})
				}
			case <-m.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// StopCleanup stops the cleanup goroutine.
func (m *Manager) StopCleanup() {
	close(m.stopCh)
}
