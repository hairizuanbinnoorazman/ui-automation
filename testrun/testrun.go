package testrun

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrTestRunNotFound is returned when a test run is not found.
	ErrTestRunNotFound = errors.New("test run not found")

	// ErrInvalidTestProcedureID is returned when test_procedure_id is not set.
	ErrInvalidTestProcedureID = errors.New("test_procedure_id is required")

	// ErrInvalidExecutedBy is returned when executed_by is not set.
	ErrInvalidExecutedBy = errors.New("executed_by is required")

	// ErrInvalidStatus is returned when status is invalid.
	ErrInvalidStatus = errors.New("invalid status")

	// ErrTestRunNotRunning is returned when trying to complete a test run that's not running.
	ErrTestRunNotRunning = errors.New("test run is not running")

	// ErrTestRunAlreadyStarted is returned when trying to start an already started test run.
	ErrTestRunAlreadyStarted = errors.New("test run already started")
)

// Status represents the status of a test run.
type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

// IsValid checks if the status is valid.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusPassed, StatusFailed, StatusSkipped:
		return true
	default:
		return false
	}
}

// IsFinal checks if the status is a final status (can't be changed).
func (s Status) IsFinal() bool {
	return s == StatusPassed || s == StatusFailed || s == StatusSkipped
}

// TestRun represents a test run in the system.
type TestRun struct {
	ID              uuid.UUID  `json:"id" gorm:"type:char(36);primaryKey"`
	TestProcedureID uuid.UUID  `json:"test_procedure_id" gorm:"type:char(36);not null;index:idx_test_procedure_id"`
	ExecutedBy      uuid.UUID  `json:"executed_by" gorm:"type:char(36);not null;index:idx_executed_by"`
	AssignedTo      *uuid.UUID `json:"assigned_to" gorm:"type:char(36);index:idx_assigned_to"`
	Status          Status     `json:"status" gorm:"type:varchar(20);not null;default:'pending';index:idx_status"`
	Notes           string     `json:"notes" gorm:"type:text"`
	StartedAt       *time.Time `json:"started_at,omitempty" gorm:"index:idx_started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// BeforeCreate hook to generate UUID before creating a new test run
func (tr *TestRun) BeforeCreate(tx *gorm.DB) error {
	if tr.ID == uuid.Nil {
		tr.ID = uuid.New()
	}
	return nil
}

// Validate checks if the test run has valid required fields.
func (tr *TestRun) Validate() error {
	if tr.TestProcedureID == uuid.Nil {
		return ErrInvalidTestProcedureID
	}
	if tr.ExecutedBy == uuid.Nil {
		return ErrInvalidExecutedBy
	}
	if !tr.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

// Start sets the started_at timestamp and changes status to running.
// Returns an error if the test run has already been started.
func (tr *TestRun) Start() error {
	if tr.StartedAt != nil {
		return ErrTestRunAlreadyStarted
	}
	now := time.Now()
	tr.StartedAt = &now
	tr.Status = StatusRunning
	return nil
}

// Complete sets the completed_at timestamp and final status.
// Returns an error if the test run is not currently running.
func (tr *TestRun) Complete(status Status, notes string) error {
	if tr.Status != StatusRunning {
		return ErrTestRunNotRunning
	}
	if !status.IsFinal() {
		return ErrInvalidStatus
	}
	now := time.Now()
	tr.CompletedAt = &now
	tr.Status = status
	if notes != "" {
		tr.Notes = notes
	}
	return nil
}
