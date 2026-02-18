package testprocedure

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrTestProcedureNotFound is returned when a test procedure is not found.
	ErrTestProcedureNotFound = errors.New("test procedure not found")

	// ErrInvalidTestProcedureName is returned when a test procedure name is empty or invalid.
	ErrInvalidTestProcedureName = errors.New("test procedure name is required")

	// ErrInvalidProjectID is returned when project_id is not set.
	ErrInvalidProjectID = errors.New("project_id is required")

	// ErrInvalidCreatedBy is returned when created_by is not set.
	ErrInvalidCreatedBy = errors.New("created_by is required")

	// ErrInvalidSteps is returned when steps JSON is invalid.
	ErrInvalidSteps = errors.New("invalid steps JSON")

	// ErrDraftNotFound is returned when a draft version is not found.
	ErrDraftNotFound = errors.New("draft version not found")

	// ErrNoCommittedVersion is returned when no committed version exists.
	ErrNoCommittedVersion = errors.New("no committed version exists")

	// ErrInvalidStepName is returned when a step name is empty.
	ErrInvalidStepName = errors.New("step name is required")
)

// TestStep represents a single step in a test procedure.
type TestStep struct {
	Name         string   `json:"name"`
	Instructions string   `json:"instructions"`
	ImagePaths   []string `json:"image_paths"`
}

// Steps represents the JSON steps for a test procedure.
// It's a custom type to handle JSON marshaling/unmarshaling.
type Steps []TestStep

// Value implements the driver.Valuer interface for database storage.
func (s Steps) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]TestStep{})
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval.
func (s *Steps) Scan(value interface{}) error {
	if value == nil {
		*s = []TestStep{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Steps: not a byte slice")
	}

	var steps []TestStep
	if err := json.Unmarshal(bytes, &steps); err != nil {
		return err
	}
	*s = steps
	return nil
}

// TestProcedure represents a test procedure in the system.
type TestProcedure struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primaryKey"`
	ProjectID   uuid.UUID  `json:"project_id" gorm:"type:char(36);not null;index:idx_project_id"`
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Steps       Steps      `json:"steps" gorm:"type:json"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:char(36);not null;index:idx_created_by"`
	Version     uint       `json:"version" gorm:"not null;default:0;index:idx_version"`
	IsLatest    bool       `json:"is_latest" gorm:"not null;default:false;index:idx_is_latest"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" gorm:"type:char(36);index:idx_parent_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new test procedure
func (tp *TestProcedure) BeforeCreate(tx *gorm.DB) error {
	if tp.ID == uuid.Nil {
		tp.ID = uuid.New()
	}
	return nil
}

// Validate checks if the test procedure has valid required fields.
func (tp *TestProcedure) Validate() error {
	if tp.Name == "" {
		return ErrInvalidTestProcedureName
	}
	if tp.ProjectID == uuid.Nil {
		return ErrInvalidProjectID
	}
	if tp.CreatedBy == uuid.Nil {
		return ErrInvalidCreatedBy
	}
	// Validate steps: ensure all step names are non-empty
	for i, step := range tp.Steps {
		if step.Name == "" {
			return errors.New("step " + strconv.Itoa(i+1) + ": " + ErrInvalidStepName.Error())
		}
	}
	return nil
}
