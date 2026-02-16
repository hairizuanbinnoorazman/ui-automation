package session

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrSessionNotFound is returned when a session is not found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when a session has expired.
	ErrSessionExpired = errors.New("session expired")
)

// Session represents a user session.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Store is an in-memory session store.
type Store struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
}

// NewStore creates a new in-memory session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[uuid.UUID]*Session),
	}
}

// Set stores a session in the store.
func (s *Store) Set(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

// Get retrieves a session from the store.
func (s *Store) Get(sessionID uuid.UUID) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// Delete removes a session from the store.
func (s *Store) Delete(sessionID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// Cleanup removes expired sessions from the store.
func (s *Store) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
			removed++
		}
	}

	return removed
}
