package apitoken

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	rawToken, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Token should have uat_ prefix
	if !strings.HasPrefix(rawToken, "uat_") {
		t.Errorf("GenerateToken() raw token should have 'uat_' prefix, got %s", rawToken)
	}

	// Token should be reasonably long (uat_ + base64url of 32 bytes = 4 + 43 = 47 chars)
	if len(rawToken) < 40 {
		t.Errorf("GenerateToken() raw token too short: %d chars", len(rawToken))
	}

	// Hash should differ from raw token
	if hash == rawToken {
		t.Error("GenerateToken() hash should differ from raw token")
	}

	// Hash should be 64 hex chars (SHA-256)
	if len(hash) != 64 {
		t.Errorf("GenerateToken() hash length = %d, want 64", len(hash))
	}
}

func TestHashToken(t *testing.T) {
	t.Parallel()

	raw := "uat_test_token_value"
	hash1 := HashToken(raw)
	hash2 := HashToken(raw)

	// Hash should be deterministic
	if hash1 != hash2 {
		t.Error("HashToken() should be deterministic")
	}

	// Hash should be 64 hex chars
	if len(hash1) != 64 {
		t.Errorf("HashToken() length = %d, want 64", len(hash1))
	}

	// Different inputs should produce different hashes
	hash3 := HashToken("uat_different_token")
	if hash1 == hash3 {
		t.Error("HashToken() different inputs should produce different hashes")
	}
}

func TestValidateExpiry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{
			name:     "zero returns default (1 month)",
			input:    0,
			expected: DefaultExpiry,
		},
		{
			name:     "below min clamps to min (1 day)",
			input:    1 * time.Hour,
			expected: MinExpiry,
		},
		{
			name:     "above max clamps to max (1 year)",
			input:    400 * 24 * time.Hour,
			expected: MaxExpiry,
		},
		{
			name:     "valid duration passes through",
			input:    7 * 24 * time.Hour,
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "exactly min passes through",
			input:    MinExpiry,
			expected: MinExpiry,
		},
		{
			name:     "exactly max passes through",
			input:    MaxExpiry,
			expected: MaxExpiry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ValidateExpiry(tt.input)
			if err != nil {
				t.Fatalf("ValidateExpiry() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("ValidateExpiry(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
