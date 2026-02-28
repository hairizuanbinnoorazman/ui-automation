package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/apitoken"
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

	// ScopeKey is the context key for the authenticated scope.
	ScopeKey ContextKey = "scope"

	// AuthMethodKey is the context key for the authentication method.
	AuthMethodKey ContextKey = "auth_method"
)

// AuthMiddleware validates session cookies or Bearer tokens and adds user info to context.
type AuthMiddleware struct {
	sessionManager *session.Manager
	tokenStore     apitoken.Store
	cookieName     string
	logger         logger.Logger
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(sessionManager *session.Manager, tokenStore apitoken.Store, cookieName string, log logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
		tokenStore:     tokenStore,
		cookieName:     cookieName,
		logger:         log,
	}
}

// Handler wraps an HTTP handler with authentication.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for Bearer token first
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			rawToken := strings.TrimPrefix(authHeader, "Bearer ")
			m.handleBearerAuth(w, r, next, rawToken)
			return
		}

		// Fall back to session cookie
		m.handleSessionAuth(w, r, next)
	})
}

// handleBearerAuth authenticates via API token.
func (m *AuthMiddleware) handleBearerAuth(w http.ResponseWriter, r *http.Request, next http.Handler, rawToken string) {
	hash := apitoken.HashToken(rawToken)
	token, err := m.tokenStore.GetByTokenHash(r.Context(), hash)
	if err != nil {
		m.logger.Warn(r.Context(), "invalid bearer token", map[string]interface{}{
			"path": r.URL.Path,
		})
		respondError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	ctx := context.WithValue(r.Context(), UserIDKey, token.UserID)
	ctx = context.WithValue(ctx, ScopeKey, token.Scope)
	ctx = context.WithValue(ctx, AuthMethodKey, "bearer")

	next.ServeHTTP(w, r.WithContext(ctx))
}

// handleSessionAuth authenticates via session cookie.
func (m *AuthMiddleware) handleSessionAuth(w http.ResponseWriter, r *http.Request, next http.Handler) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		m.logger.Warn(r.Context(), "missing session cookie", map[string]interface{}{
			"path": r.URL.Path,
		})
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		m.logger.Warn(r.Context(), "invalid session ID format", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	sess, err := m.sessionManager.Get(sessionID)
	if err != nil {
		m.logger.Warn(r.Context(), "invalid or expired session", map[string]interface{}{
			"error":      err.Error(),
			"session_id": sessionID.String(),
		})
		respondError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	ctx := context.WithValue(r.Context(), UserIDKey, sess.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, sess.Email)
	ctx = context.WithValue(ctx, ScopeKey, apitoken.ScopeReadWrite)
	ctx = context.WithValue(ctx, AuthMethodKey, "session")

	next.ServeHTTP(w, r.WithContext(ctx))
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

// GetScope extracts the scope from the request context.
func GetScope(ctx context.Context) string {
	scope, ok := ctx.Value(ScopeKey).(string)
	if !ok {
		return apitoken.ScopeReadWrite
	}
	return scope
}

// GetAuthMethod extracts the authentication method from the request context.
func GetAuthMethod(ctx context.Context) string {
	method, ok := ctx.Value(AuthMethodKey).(string)
	if !ok {
		return "session"
	}
	return method
}

// RequireWriteScope checks if the current request has write scope.
// Returns true if the scope is read_write, false otherwise (and writes a 403 response).
func RequireWriteScope(w http.ResponseWriter, r *http.Request) bool {
	scope := GetScope(r.Context())
	if scope != apitoken.ScopeReadWrite {
		respondError(w, http.StatusForbidden, "write access required")
		return false
	}
	return true
}

// WriteScopeMiddleware enforces write scope for state-mutating HTTP methods.
// GET and HEAD requests pass through regardless of scope. POST, PUT, DELETE,
// and PATCH require read_write scope.
func WriteScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
			if !RequireWriteScope(w, r) {
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
