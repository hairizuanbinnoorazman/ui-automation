package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	t.Parallel()
	key := DeriveKey("test-passphrase")
	assert.Len(t, key, 32)
	// Same input should produce same key
	key2 := DeriveKey("test-passphrase")
	assert.Equal(t, key, key2)
	// Different input should produce different key
	key3 := DeriveKey("different-passphrase")
	assert.NotEqual(t, key, key3)
}

func TestEncryptDecryptCredentials(t *testing.T) {
	t.Parallel()
	key := DeriveKey("test-passphrase")
	creds := map[string]string{
		"token": "ghp_test_token",
		"url":   "https://example.com",
	}

	encrypted, err := EncryptCredentials(key, creds)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := DecryptCredentials(key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, creds, decrypted)
}

func TestDecryptWithWrongKey(t *testing.T) {
	t.Parallel()
	key1 := DeriveKey("passphrase-1")
	key2 := DeriveKey("passphrase-2")
	creds := map[string]string{"token": "secret"}

	encrypted, err := EncryptCredentials(key1, creds)
	require.NoError(t, err)

	_, err = DecryptCredentials(key2, encrypted)
	assert.Error(t, err)
}

func TestEncryptEmptyCredentials(t *testing.T) {
	t.Parallel()
	key := DeriveKey("test")
	creds := map[string]string{}

	encrypted, err := EncryptCredentials(key, creds)
	require.NoError(t, err)

	decrypted, err := DecryptCredentials(key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, creds, decrypted)
}

func TestDecryptTooShortCiphertext(t *testing.T) {
	t.Parallel()
	key := DeriveKey("test")
	_, err := DecryptCredentials(key, []byte("short"))
	assert.Error(t, err)
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	t.Parallel()
	key := DeriveKey("test")
	creds := map[string]string{"token": "value"}

	enc1, err := EncryptCredentials(key, creds)
	require.NoError(t, err)

	enc2, err := EncryptCredentials(key, creds)
	require.NoError(t, err)

	// Due to random nonce, same plaintext should produce different ciphertexts
	assert.NotEqual(t, enc1, enc2)
}
