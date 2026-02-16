package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooShort is returned when a password is less than 8 characters.
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// ErrInvalidEmail is returned when an email is empty or invalid.
	ErrInvalidEmail = errors.New("email is required")

	// ErrInvalidUsername is returned when a username is empty or invalid.
	ErrInvalidUsername = errors.New("username is required")
)

// User represents a user in the system.
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Username     string    `json:"username" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetPassword hashes and sets the user's password.
// Returns an error if the password is too short.
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies if the provided password matches the user's password hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Validate checks if the user has valid required fields.
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.Username == "" {
		return ErrInvalidUsername
	}
	return nil
}
