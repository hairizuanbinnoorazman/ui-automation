package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/hairizuanbinnoorazman/ui-automation/apitoken"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
)

// APITokenHandler handles API token-related requests.
type APITokenHandler struct {
	tokenStore apitoken.Store
	logger     logger.Logger
}

// NewAPITokenHandler creates a new API token handler.
func NewAPITokenHandler(tokenStore apitoken.Store, log logger.Logger) *APITokenHandler {
	return &APITokenHandler{
		tokenStore: tokenStore,
		logger:     log,
	}
}

// CreateTokenRequest represents a token creation request.
type CreateTokenRequest struct {
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	ExpiresInHours int    `json:"expires_in_hours"`
}

// CreateTokenResponse includes the raw token (shown once).
type CreateTokenResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// TokenListItem represents a token in list responses (no secret).
type TokenListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	ExpiresAt string `json:"expires_at"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// TokenListResponse is the response for listing tokens.
type TokenListResponse struct {
	Tokens []TokenListItem `json:"tokens"`
	Total  int             `json:"total"`
}

// Create handles creating a new API token.
func (h *APITokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req CreateTokenRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "token name is required")
		return
	}

	// Default scope
	if req.Scope == "" {
		req.Scope = apitoken.ScopeReadOnly
	}
	if req.Scope != apitoken.ScopeReadOnly && req.Scope != apitoken.ScopeReadWrite {
		respondError(w, http.StatusBadRequest, "invalid scope: must be read_only or read_write")
		return
	}

	// Validate expiry
	var expiryDuration time.Duration
	if req.ExpiresInHours > 0 {
		expiryDuration = time.Duration(req.ExpiresInHours) * time.Hour
	}
	expiryDuration, err := apitoken.ValidateExpiry(expiryDuration)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Generate token
	rawToken, hash, err := apitoken.GenerateToken()
	if err != nil {
		h.logger.Error(r.Context(), "failed to generate token", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	token := &apitoken.APIToken{
		UserID:    userID,
		Name:      req.Name,
		TokenHash: hash,
		Scope:     req.Scope,
		ExpiresAt: time.Now().Add(expiryDuration),
		IsActive:  true,
	}

	if err := h.tokenStore.Create(r.Context(), token); err != nil {
		if errors.Is(err, apitoken.ErrMaxTokensReached) {
			respondError(w, http.StatusConflict, "maximum number of active tokens reached (limit: 5)")
			return
		}
		if errors.Is(err, apitoken.ErrInvalidTokenName) ||
			errors.Is(err, apitoken.ErrInvalidScope) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error(r.Context(), "failed to create token", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	respondJSON(w, http.StatusCreated, CreateTokenResponse{
		ID:        token.ID.String(),
		Name:      token.Name,
		Scope:     token.Scope,
		Token:     rawToken,
		ExpiresAt: token.ExpiresAt.Format(time.RFC3339),
		CreatedAt: token.CreatedAt.Format(time.RFC3339),
	})
}

// List handles listing tokens for the authenticated user.
func (h *APITokenHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	tokens, err := h.tokenStore.ListByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list tokens", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}

	items := make([]TokenListItem, len(tokens))
	for i, t := range tokens {
		items[i] = TokenListItem{
			ID:        t.ID.String(),
			Name:      t.Name,
			Scope:     t.Scope,
			ExpiresAt: t.ExpiresAt.Format(time.RFC3339),
			IsActive:  t.IsActive,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	respondJSON(w, http.StatusOK, TokenListResponse{
		Tokens: items,
		Total:  len(items),
	})
}

// Revoke handles revoking (soft-deleting) a token.
func (h *APITokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	tokenID, ok := parseUUIDOrRespond(w, r, "token_id", "token")
	if !ok {
		return
	}

	// Verify ownership
	token, err := h.tokenStore.GetByID(r.Context(), tokenID)
	if err != nil {
		if errors.Is(err, apitoken.ErrTokenNotFound) {
			respondError(w, http.StatusNotFound, "token not found")
			return
		}
		h.logger.Error(r.Context(), "failed to get token for authorization", map[string]interface{}{
			"error":    err.Error(),
			"token_id": tokenID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to verify token ownership")
		return
	}

	if token.UserID != userID {
		h.logger.Warn(r.Context(), "unauthorized token revoke attempt", map[string]interface{}{
			"user_id":  userID.String(),
			"token_id": tokenID.String(),
			"owner_id": token.UserID.String(),
		})
		respondError(w, http.StatusForbidden, "you don't have access to this token")
		return
	}

	if err := h.tokenStore.Revoke(r.Context(), tokenID); err != nil {
		if errors.Is(err, apitoken.ErrTokenNotFound) {
			respondError(w, http.StatusNotFound, "token not found")
			return
		}
		h.logger.Error(r.Context(), "failed to revoke token", map[string]interface{}{
			"error":    err.Error(),
			"token_id": tokenID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to revoke token")
		return
	}

	respondSuccess(w, "token revoked successfully")
}
