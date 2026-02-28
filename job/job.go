package job

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrJobNotFound      = errors.New("job not found")
	ErrInvalidJobType   = errors.New("job type is required")
	ErrInvalidCreatedBy = errors.New("created_by is required")
	ErrInvalidStatus    = errors.New("invalid job status")
	ErrJobAlreadyStarted = errors.New("job already started")
	ErrJobNotRunning    = errors.New("job is not running")
)

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusFailed  Status = "failed"
	StatusSuccess Status = "success"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusCreated, StatusRunning, StatusStopped, StatusFailed, StatusSuccess:
		return true
	}
	return false
}

type JobType string

const (
	JobTypeUIExploration JobType = "ui_exploration"
)

func (jt JobType) IsValid() bool {
	switch jt {
	case JobTypeUIExploration:
		return true
	}
	return false
}

// JSONMap is a custom type for JSON columns.
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: not a byte slice")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	*j = m
	return nil
}

type Job struct {
	ID        uuid.UUID  `json:"id" gorm:"type:char(36);primaryKey"`
	Type      JobType    `json:"type" gorm:"column:type;type:varchar(50);not null"`
	Status    Status     `json:"status" gorm:"type:varchar(20);not null;default:'created'"`
	Config    JSONMap    `json:"config" gorm:"type:json"`
	Result    JSONMap    `json:"result" gorm:"type:json"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Duration  *int64     `json:"duration,omitempty"`
	CreatedBy uuid.UUID  `json:"created_by" gorm:"type:char(36);not null;index:idx_jobs_created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (j *Job) Validate() error {
	if !j.Type.IsValid() {
		return ErrInvalidJobType
	}
	if j.CreatedBy == uuid.Nil {
		return ErrInvalidCreatedBy
	}
	return nil
}

// Start marks the job as running.
func (j *Job) Start() error {
	if j.Status != StatusCreated {
		return ErrJobAlreadyStarted
	}
	now := time.Now()
	j.Status = StatusRunning
	j.StartTime = &now
	return nil
}

// Complete marks the job as finished with the given status and result.
func (j *Job) Complete(status Status, result JSONMap) error {
	if j.Status != StatusRunning {
		return ErrJobNotRunning
	}
	now := time.Now()
	j.Status = status
	j.EndTime = &now
	j.Result = result
	if j.StartTime != nil {
		duration := now.Sub(*j.StartTime).Milliseconds()
		j.Duration = &duration
	}
	return nil
}
