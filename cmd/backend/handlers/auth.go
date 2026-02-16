package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/hairizuan-noorazman/ui-automation/logger"
	"github.com/hairizuan-noorazman/ui-automation/session"
	"github.com/hairizuan-noorazman/ui-automation/user"
)

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	userStore      user.Store
	sessionManager *session.Manager
	secureCookie   *securecookie.SecureCookie
	cookieName     string
	cookieSecure   bool
	logger         logger.Logger
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(
	userStore user.Store,
	sessionManager *session.Manager,
	cookieSecret string,
	cookieName string,
	cookieSecure bool,
	log logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		userStore:      userStore,
		sessionManager: sessionManager,
		secureCookie:   securecookie.New([]byte(cookieSecret), nil),
		cookieName:     cookieName,
		cookieSecure:   cookieSecure,
		logger:         log,
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create user
	newUser := &user.User{
		Email:    req.Email,
		Username: req.Username,
		IsActive: true,
	}

	if err := newUser.SetPassword(req.Password); err != nil {
		h.logger.Error(r.Context(), "failed to set password", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userStore.Create(r.Context(), newUser); err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		h.logger.Error(r.Context(), "failed to create user", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create session
	sess, err := h.sessionManager.Create(newUser.ID, newUser.Email)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create session", map[string]interface{}{
			"error":   err.Error(),
			"user_id": newUser.ID,
		})
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, sess.ID)

	h.logger.Info(r.Context(), "user registered", map[string]interface{}{
		"user_id": newUser.ID,
		"email":   newUser.Email,
	})

	respondJSON(w, http.StatusCreated, newUser)
}

// Login handles user login requests.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user by email
	existingUser, err := h.userStore.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error(r.Context(), "failed to get user", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		respondError(w, http.StatusInternalServerError, "authentication failed")
		return
	}

	// Check password
	if !existingUser.CheckPassword(req.Password) {
		h.logger.Warn(r.Context(), "invalid password attempt", map[string]interface{}{
			"email": req.Email,
		})
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Create session
	sess, err := h.sessionManager.Create(existingUser.ID, existingUser.Email)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create session", map[string]interface{}{
			"error":   err.Error(),
			"user_id": existingUser.ID,
		})
		respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, sess.ID)

	h.logger.Info(r.Context(), "user logged in", map[string]interface{}{
		"user_id": existingUser.ID,
		"email":   existingUser.Email,
	})

	respondJSON(w, http.StatusOK, existingUser)
}

// Logout handles user logout requests.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract session cookie
	cookie, err := r.Cookie(h.cookieName)
	if err == nil {
		// Delete session
		h.sessionManager.Delete(cookie.Value)
	}

	// Clear cookie
	h.clearSessionCookie(w)

	respondSuccess(w, "logged out successfully")
}

// setSessionCookie sets a session cookie in the response.
func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

// clearSessionCookie clears the session cookie.
func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}
