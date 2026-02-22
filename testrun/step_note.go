package testrun

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrStepNoteNotFound is returned when a step note is not found.
	ErrStepNoteNotFound = errors.New("step note not found")
)

// StepNote represents notes for a specific test procedure step within a test run.
type StepNote struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	TestRunID uuid.UUID `json:"test_run_id" gorm:"type:char(36);not null"`
	StepIndex int       `json:"step_index" gorm:"not null"`
	Notes     string    `json:"notes" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new step note.
func (sn *StepNote) BeforeCreate(tx *gorm.DB) error {
	if sn.ID == uuid.Nil {
		sn.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for GORM.
func (sn *StepNote) TableName() string {
	return "test_run_step_notes"
}
