package project

import (
	"errors"
	"time"
)

var (
	// ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")

	// ErrInvalidProjectName is returned when a project name is empty or invalid.
	ErrInvalidProjectName = errors.New("project name is required")

	// ErrInvalidOwner is returned when owner_id is not set.
	ErrInvalidOwner = errors.New("owner_id is required")
)

// Project represents a test procedure project in the system.
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	OwnerID     uint      `json:"owner_id" gorm:"not null;index:idx_owner_id"`
	IsActive    bool      `json:"is_active" gorm:"default:true;index:idx_is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks if the project has valid required fields.
func (p *Project) Validate() error {
	if p.Name == "" {
		return ErrInvalidProjectName
	}
	if p.OwnerID == 0 {
		return ErrInvalidOwner
	}
	return nil
}
