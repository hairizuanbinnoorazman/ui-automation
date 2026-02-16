package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_SetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "password at minimum length",
			password: "12345678",
			wantErr:  nil,
		},
		{
			name:     "short password",
			password: "short",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{}
			err := user.SetPassword(tt.password)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, user.PasswordHash)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEqual(t, tt.password, user.PasswordHash)
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "password123"
	require.NoError(t, user.SetPassword(password))

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "password123",
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.password)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr error
	}{
		{
			name: "valid user",
			user: User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			wantErr: nil,
		},
		{
			name: "missing email",
			user: User{
				Username: "testuser",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "missing username",
			user: User{
				Email: "test@example.com",
			},
			wantErr: ErrInvalidUsername,
		},
		{
			name:    "missing both",
			user:    User{},
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
