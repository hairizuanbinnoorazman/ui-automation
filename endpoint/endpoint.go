package endpoint

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrEndpointNotFound    = errors.New("endpoint not found")
	ErrInvalidEndpointName = errors.New("endpoint name is required")
	ErrInvalidEndpointURL  = errors.New("endpoint URL is required")
	ErrInvalidCreatedBy    = errors.New("created_by is required")
)

type Credential struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Credentials []Credential

// Value implements driver.Valuer for database storage.
func (c Credentials) Value() (driver.Value, error) {
	if c == nil {
		return json.Marshal([]Credential{})
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for database retrieval.
func (c *Credentials) Scan(value interface{}) error {
	if value == nil {
		*c = []Credential{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Credentials: not a byte slice")
	}
	var creds []Credential
	if err := json.Unmarshal(bytes, &creds); err != nil {
		return err
	}
	*c = creds
	return nil
}

type Endpoint struct {
	ID          uuid.UUID   `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string      `json:"name" gorm:"not null"`
	URL         string      `json:"url" gorm:"not null"`
	Credentials Credentials `json:"credentials" gorm:"type:json"`
	CreatedBy   uuid.UUID   `json:"created_by" gorm:"type:char(36);not null;index:idx_endpoints_created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new endpoint.
func (e *Endpoint) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// Validate checks if the endpoint has valid required fields.
func (e *Endpoint) Validate() error {
	if e.Name == "" {
		return ErrInvalidEndpointName
	}
	if e.URL == "" {
		return ErrInvalidEndpointURL
	}
	if e.CreatedBy == uuid.Nil {
		return ErrInvalidCreatedBy
	}
	return nil
}

// DefaultCredentials returns the default credential template.
func DefaultCredentials() Credentials {
	return Credentials{
		{Key: "username", Value: ""},
		{Key: "email", Value: ""},
	}
}
