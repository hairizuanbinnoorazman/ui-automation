package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/session"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"

	// UserEmailKey is the context key for user email.
	UserEmailKey ContextKey = "user_email"
)

// AuthMiddleware validates session cookies and adds user info to context.
type AuthMiddleware struct {
	sessionManager *session.Manager
	cookieName     string
	logger         logger.Logger
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(sessionManager *session.Manager, cookieName string, log logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		cookieName:     cookieName,
		logger:         log,
	}
}

// Handler wraps an HTTP handler with authentication.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract session cookie
		cookie, err := r.Cookie(m.cookieName)
		if err != nil {
			m.logger.Warn(r.Context(), "missing session cookie", map[string]interface{}{
				"path": r.URL.Path,
			})
			respondError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		// Parse session ID as UUID
		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			m.logger.Warn(r.Context(), "invalid session ID format", map[string]interface{}{
				"error": err.Error(),
			})
			respondError(w, http.StatusUnauthorized, "invalid session")
			return
		}

		// Validate session
		sess, err := m.sessionManager.Get(sessionID)
		if err != nil {
			m.logger.Warn(r.Context(), "invalid or expired session", map[string]interface{}{
				"error":      err.Error(),
				"session_id": sessionID.String(),
			})
			respondError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, sess.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, sess.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserEmail extracts the user email from the request context.
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
