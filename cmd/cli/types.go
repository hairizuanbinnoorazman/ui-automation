package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/testrun"
)

// PaginatedResponse matches handlers.PaginatedResponse.
type PaginatedResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ErrorResponse matches handlers.ErrorResponse.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse matches handlers.SuccessResponse.
type SuccessResponse struct {
	Message string `json:"message"`
}

// CreateProjectRequest matches handlers.CreateProjectRequest.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateProjectRequest matches handlers.UpdateProjectRequest.
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateTestProcedureRequest matches handlers.CreateTestProcedureRequest.
type CreateTestProcedureRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Steps       []struct {
		Name         string   `json:"name"`
		Instructions string   `json:"instructions"`
		ImagePaths   []string `json:"image_paths"`
	} `json:"steps"`
}

// UpdateTestProcedureRequest matches handlers.UpdateTestProcedureRequest.
type UpdateTestProcedureRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Steps       *[]struct {
		Name         string   `json:"name"`
		Instructions string   `json:"instructions"`
		ImagePaths   []string `json:"image_paths"`
	} `json:"steps,omitempty"`
}

// UpdateTestRunRequest matches handlers.UpdateTestRunRequest.
type UpdateTestRunRequest struct {
	Notes      *string `json:"notes,omitempty"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

// CompleteTestRunRequest matches handlers.CompleteTestRunRequest.
type CompleteTestRunRequest struct {
	Status testrun.Status `json:"status"`
	Notes  string         `json:"notes"`
}

// TestRunWithVersion matches handlers.testRunWithVersion.
type TestRunWithVersion struct {
	testrun.TestRun
	ProcedureVersion uint `json:"procedure_version"`
}

// CreateTokenRequest matches handlers.CreateTokenRequest.
type CreateTokenRequest struct {
	Name           string `json:"name"`
	Scope          string `json:"scope"`
	ExpiresInHours int    `json:"expires_in_hours"`
}

// CreateTokenResponse matches handlers.CreateTokenResponse.
type CreateTokenResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// TokenListItem matches handlers.TokenListItem.
type TokenListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	ExpiresAt string `json:"expires_at"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// TokenListResponse matches handlers.TokenListResponse.
type TokenListResponse struct {
	Tokens []TokenListItem `json:"tokens"`
	Total  int             `json:"total"`
}

// ProjectResponse is used for deserializing project responses.
type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TestProcedureResponse is used for deserializing test procedure responses.
type TestProcedureResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []StepJSON `json:"steps"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	Version     uint       `json:"version"`
	IsLatest    bool       `json:"is_latest"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// StepJSON is used for deserializing step data from API responses.
type StepJSON struct {
	Name         string   `json:"name"`
	Instructions string   `json:"instructions"`
	ImagePaths   []string `json:"image_paths"`
}

// TestRunResponse is used for deserializing test run responses.
type TestRunResponse struct {
	ID               uuid.UUID      `json:"id"`
	TestProcedureID  uuid.UUID      `json:"test_procedure_id"`
	ExecutedBy       uuid.UUID      `json:"executed_by"`
	AssignedTo       *uuid.UUID     `json:"assigned_to"`
	Status           testrun.Status `json:"status"`
	Notes            string         `json:"notes"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	ProcedureVersion uint           `json:"procedure_version"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}
