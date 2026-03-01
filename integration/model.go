package integration

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/issuetracker"
	"gorm.io/gorm"
)

var (
	ErrIntegrationNotFound = errors.New("integration not found")
	ErrIssueLinkNotFound   = errors.New("issue link not found")
	ErrInvalidName         = errors.New("name is required")
	ErrInvalidProvider     = errors.New("invalid provider type")
	ErrInvalidUserID       = errors.New("user_id is required")
	ErrInvalidTestRunID    = errors.New("test_run_id is required")
	ErrInvalidIntegrationID = errors.New("integration_id is required")
	ErrInvalidExternalID   = errors.New("external_id is required")
)

type Integration struct {
	ID                   uuid.UUID                 `json:"id" gorm:"type:char(36);primaryKey"`
	UserID               uuid.UUID                 `json:"user_id" gorm:"type:char(36);not null;index:idx_integrations_user_id"`
	Name                 string                    `json:"name" gorm:"type:varchar(255);not null"`
	Provider             issuetracker.ProviderType `json:"provider" gorm:"type:varchar(20);not null"`
	EncryptedCredentials []byte                    `json:"-" gorm:"type:blob;not null"`
	IsActive             bool                      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

func (i *Integration) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

func (i *Integration) Validate() error {
	if i.Name == "" {
		return ErrInvalidName
	}
	if !i.Provider.IsValid() {
		return ErrInvalidProvider
	}
	if i.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	return nil
}

type IssueLink struct {
	ID            uuid.UUID                 `json:"id" gorm:"type:char(36);primaryKey"`
	TestRunID     uuid.UUID                 `json:"test_run_id" gorm:"type:char(36);not null;index:idx_issue_links_test_run_id"`
	IntegrationID uuid.UUID                 `json:"integration_id" gorm:"type:char(36);not null;index:idx_issue_links_integration_id"`
	ExternalID    string                    `json:"external_id" gorm:"type:varchar(255);not null"`
	Title         string                    `json:"title" gorm:"type:varchar(500)"`
	Status        string                    `json:"status" gorm:"type:varchar(50)"`
	URL           string                    `json:"url" gorm:"type:varchar(1000)"`
	Provider      issuetracker.ProviderType `json:"provider" gorm:"type:varchar(20);not null"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

func (il *IssueLink) BeforeCreate(tx *gorm.DB) error {
	if il.ID == uuid.Nil {
		il.ID = uuid.New()
	}
	return nil
}

func (il *IssueLink) Validate() error {
	if il.TestRunID == uuid.Nil {
		return ErrInvalidTestRunID
	}
	if il.IntegrationID == uuid.Nil {
		return ErrInvalidIntegrationID
	}
	if il.ExternalID == "" {
		return ErrInvalidExternalID
	}
	if !il.Provider.IsValid() {
		return ErrInvalidProvider
	}
	return nil
}
