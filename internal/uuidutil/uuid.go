package uuidutil

import (
	"github.com/google/uuid"
)

// Parse safely parses a string into a UUID with error handling
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a string into a UUID and panics on error
// Use this only in tests or when you're certain the input is valid
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// New generates a new random UUID v4
func New() uuid.UUID {
	return uuid.New()
}

// IsValid checks if a string is a valid UUID format
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// Nil returns the nil UUID constant (00000000-0000-0000-0000-000000000000)
func Nil() uuid.UUID {
	return uuid.Nil
}
