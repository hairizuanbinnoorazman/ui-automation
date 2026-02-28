package apitoken

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		userID := uuid.New()
		_, hash, _ := GenerateToken()

		token := &APIToken{
			UserID:    userID,
			Name:      "test-token",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}

		err := store.Create(ctx, token)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if token.ID == uuid.Nil {
			t.Error("Create() should generate an ID")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}

		err := store.Create(ctx, token)
		if err != ErrInvalidTokenName {
			t.Errorf("Create() error = %v, want %v", err, ErrInvalidTokenName)
		}
	})

	t.Run("invalid scope", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "test-token",
			TokenHash: hash,
			Scope:     "invalid",
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}

		err := store.Create(ctx, token)
		if err != ErrInvalidScope {
			t.Errorf("Create() error = %v, want %v", err, ErrInvalidScope)
		}
	})

	t.Run("max tokens reached", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		userID := uuid.New()

		// Create max tokens
		for i := 0; i < MaxTokensPerUser; i++ {
			_, hash, _ := GenerateToken()
			token := &APIToken{
				UserID:    userID,
				Name:      "token-" + string(rune('A'+i)),
				TokenHash: hash,
				Scope:     ScopeReadOnly,
				ExpiresAt: time.Now().Add(DefaultExpiry),
				IsActive:  true,
			}
			if err := store.Create(ctx, token); err != nil {
				t.Fatalf("Create() token %d error = %v", i, err)
			}
		}

		// 6th token should fail
		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    userID,
			Name:      "one-too-many",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}

		err := store.Create(ctx, token)
		if err != ErrMaxTokensReached {
			t.Errorf("Create() error = %v, want %v", err, ErrMaxTokensReached)
		}
	})
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "test-token",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}
		store.Create(ctx, token)

		found, err := store.GetByID(ctx, token.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if found.Name != "test-token" {
			t.Errorf("GetByID() name = %s, want test-token", found.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, err := store.GetByID(ctx, uuid.New())
		if err != ErrTokenNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, ErrTokenNotFound)
		}
	})
}

func TestGetByTokenHash(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "test-token",
			TokenHash: hash,
			Scope:     ScopeReadWrite,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}
		store.Create(ctx, token)

		found, err := store.GetByTokenHash(ctx, hash)
		if err != nil {
			t.Fatalf("GetByTokenHash() error = %v", err)
		}
		if found.ID != token.ID {
			t.Errorf("GetByTokenHash() ID = %s, want %s", found.ID, token.ID)
		}
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "expired-token",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			IsActive:  true,
		}
		store.Create(ctx, token)

		_, err := store.GetByTokenHash(ctx, hash)
		if err != ErrTokenNotFound {
			t.Errorf("GetByTokenHash() error = %v, want %v", err, ErrTokenNotFound)
		}
	})

	t.Run("revoked", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "revoked-token",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}
		store.Create(ctx, token)

		// Revoke the token
		store.Revoke(ctx, token.ID)

		_, err := store.GetByTokenHash(ctx, hash)
		if err != ErrTokenNotFound {
			t.Errorf("GetByTokenHash() error = %v, want %v", err, ErrTokenNotFound)
		}
	})
}

func TestListByUser(t *testing.T) {
	t.Parallel()

	_, store := setupTestStore(t)
	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create tokens for user
	for i := 0; i < 3; i++ {
		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    userID,
			Name:      "token-" + string(rune('A'+i)),
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}
		store.Create(ctx, token)
	}

	// Create a token and then revoke it (should not appear in list)
	_, hash, _ := GenerateToken()
	revokedToken := &APIToken{
		UserID:    userID,
		Name:      "revoked",
		TokenHash: hash,
		Scope:     ScopeReadOnly,
		ExpiresAt: time.Now().Add(DefaultExpiry),
		IsActive:  true,
	}
	store.Create(ctx, revokedToken)
	store.Revoke(ctx, revokedToken.ID)

	// Create token for other user (should not appear)
	_, hash2, _ := GenerateToken()
	otherToken := &APIToken{
		UserID:    otherUserID,
		Name:      "other-user-token",
		TokenHash: hash2,
		Scope:     ScopeReadOnly,
		ExpiresAt: time.Now().Add(DefaultExpiry),
		IsActive:  true,
	}
	store.Create(ctx, otherToken)

	tokens, err := store.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}

	if len(tokens) != 3 {
		t.Errorf("ListByUser() returned %d tokens, want 3", len(tokens))
	}

	for _, token := range tokens {
		if token.UserID != userID {
			t.Errorf("ListByUser() returned token with wrong user ID: %s", token.UserID)
		}
		if !token.IsActive {
			t.Error("ListByUser() returned inactive token")
		}
	}
}

func TestRevoke(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		_, hash, _ := GenerateToken()
		token := &APIToken{
			UserID:    uuid.New(),
			Name:      "to-revoke",
			TokenHash: hash,
			Scope:     ScopeReadOnly,
			ExpiresAt: time.Now().Add(DefaultExpiry),
			IsActive:  true,
		}
		store.Create(ctx, token)

		err := store.Revoke(ctx, token.ID)
		if err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}

		// Should no longer be found by hash
		_, err = store.GetByTokenHash(ctx, hash)
		if err != ErrTokenNotFound {
			t.Errorf("GetByTokenHash() after revoke: error = %v, want %v", err, ErrTokenNotFound)
		}

		// Should still exist by ID
		found, err := store.GetByID(ctx, token.ID)
		if err != nil {
			t.Fatalf("GetByID() after revoke: error = %v", err)
		}
		if found.IsActive {
			t.Error("Revoke() token should be inactive")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, store := setupTestStore(t)
		ctx := context.Background()

		err := store.Revoke(ctx, uuid.New())
		if err != ErrTokenNotFound {
			t.Errorf("Revoke() error = %v, want %v", err, ErrTokenNotFound)
		}
	})
}
