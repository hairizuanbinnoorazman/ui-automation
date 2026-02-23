package scriptgen

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrScriptNotFound is returned when a generated script is not found.
	ErrScriptNotFound = errors.New("generated script not found")

	// ErrInvalidTestProcedureID is returned when test_procedure_id is not set.
	ErrInvalidTestProcedureID = errors.New("test_procedure_id is required")

	// ErrInvalidFramework is returned when framework is invalid.
	ErrInvalidFramework = errors.New("invalid framework")

	// ErrInvalidScriptPath is returned when script_path is empty.
	ErrInvalidScriptPath = errors.New("script_path is required")

	// ErrInvalidFileName is returned when file_name is empty.
	ErrInvalidFileName = errors.New("file_name is required")

	// ErrInvalidGeneratedBy is returned when generated_by is not set.
	ErrInvalidGeneratedBy = errors.New("generated_by is required")

	// ErrScriptAlreadyExists is returned when a script already exists for the procedure and framework.
	ErrScriptAlreadyExists = errors.New("script already exists for this procedure and framework")
)

// Framework represents the automation framework type.
type Framework string

const (
	FrameworkSelenium   Framework = "selenium"
	FrameworkPlaywright Framework = "playwright"
)

// IsValid checks if the framework is valid.
func (f Framework) IsValid() bool {
	switch f {
	case FrameworkSelenium, FrameworkPlaywright:
		return true
	default:
		return false
	}
}

// GenerationStatus represents the status of script generation.
type GenerationStatus string

const (
	StatusPending    GenerationStatus = "pending"
	StatusGenerating GenerationStatus = "generating"
	StatusCompleted  GenerationStatus = "completed"
	StatusFailed     GenerationStatus = "failed"
)

// IsValid checks if the generation status is valid.
func (s GenerationStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusGenerating, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

// GeneratedScript represents a generated automation script.
type GeneratedScript struct {
	ID                uuid.UUID        `json:"id" gorm:"type:char(36);primaryKey"`
	TestProcedureID   uuid.UUID        `json:"test_procedure_id" gorm:"type:char(36);not null"`
	Framework         Framework        `json:"framework" gorm:"type:varchar(20);not null"`
	ScriptPath        string           `json:"script_path" gorm:"type:varchar(512);not null"`
	FileName          string           `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize          int64            `json:"file_size" gorm:"not null"`
	GenerationStatus  GenerationStatus `json:"generation_status" gorm:"type:varchar(20);not null;default:'pending'"`
	ErrorMessage      *string          `json:"error_message,omitempty" gorm:"type:text"`
	GeneratedBy       uuid.UUID        `json:"generated_by" gorm:"type:char(36);not null"`
	GeneratedAt       time.Time        `json:"generated_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new generated script
func (gs *GeneratedScript) BeforeCreate(tx *gorm.DB) error {
	if gs.ID == uuid.Nil {
		gs.ID = uuid.New()
	}
	return nil
}

// Validate checks if the generated script has valid required fields.
func (gs *GeneratedScript) Validate() error {
	if gs.TestProcedureID == uuid.Nil {
		return ErrInvalidTestProcedureID
	}
	if !gs.Framework.IsValid() {
		return ErrInvalidFramework
	}
	if gs.GeneratedBy == uuid.Nil {
		return ErrInvalidGeneratedBy
	}
	if !gs.GenerationStatus.IsValid() {
		return errors.New("invalid generation status")
	}
	// ScriptPath and FileName are only required once generation has completed.
	if gs.GenerationStatus == StatusCompleted {
		if gs.ScriptPath == "" {
			return ErrInvalidScriptPath
		}
		if gs.FileName == "" {
			return ErrInvalidFileName
		}
	}
	return nil
}
