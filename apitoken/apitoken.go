package apitoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenNotFound    = errors.New("api token not found")
	ErrInvalidTokenName = errors.New("token name is required")
	ErrInvalidScope     = errors.New("invalid scope: must be read_only or read_write")
	ErrInvalidExpiry    = errors.New("invalid expiry duration")
	ErrMaxTokensReached = errors.New("maximum number of active tokens reached")
)

const (
	ScopeReadOnly  = "read_only"
	ScopeReadWrite = "read_write"

	MaxTokensPerUser = 5

	DefaultExpiry = 30 * 24 * time.Hour  // 1 month
	MinExpiry     = 24 * time.Hour       // 1 day
	MaxExpiry     = 365 * 24 * time.Hour // 1 year
)

// APIToken represents an API token for programmatic access.
type APIToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null;index:idx_api_tokens_user_id"`
	Name      string    `json:"name" gorm:"not null"`
	TokenHash string    `json:"-" gorm:"type:char(64);not null;uniqueIndex:idx_api_tokens_token_hash"`
	Scope     string    `json:"scope" gorm:"type:varchar(20);not null;default:read_only"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the database table name.
func (APIToken) TableName() string {
	return "api_tokens"
}

// BeforeCreate hook to generate UUID before creating a new token.
func (t *APIToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// Validate checks if the token has valid required fields.
func (t *APIToken) Validate() error {
	if t.Name == "" {
		return ErrInvalidTokenName
	}
	if t.Scope != ScopeReadOnly && t.Scope != ScopeReadWrite {
		return ErrInvalidScope
	}
	if t.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if t.TokenHash == "" {
		return errors.New("token_hash is required")
	}
	return nil
}

// IsExpired returns true if the token has expired.
func (t *APIToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// GenerateToken creates a new random token with the uat_ prefix.
// Returns the raw token string and its SHA-256 hash.
func GenerateToken() (rawToken string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	rawToken = "uat_" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	hash = HashToken(rawToken)
	return rawToken, hash, nil
}

// HashToken returns the SHA-256 hex digest of a raw token.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

// ValidateExpiry validates and normalizes an expiry duration.
// If duration is 0, returns the default (1 month).
// Clamps to min 1 day and max 1 year.
func ValidateExpiry(d time.Duration) (time.Duration, error) {
	if d == 0 {
		return DefaultExpiry, nil
	}
	if d < MinExpiry {
		return MinExpiry, nil
	}
	if d > MaxExpiry {
		return MaxExpiry, nil
	}
	return d, nil
}
