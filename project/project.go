package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"type:char(36);not null;index:idx_owner_id"`
	IsActive    bool      `json:"is_active" gorm:"default:true;index:idx_is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Validate checks if the project has valid required fields.
func (p *Project) Validate() error {
	if p.Name == "" {
		return ErrInvalidProjectName
	}
	if p.OwnerID == uuid.Nil {
		return ErrInvalidOwner
	}
	return nil
}
