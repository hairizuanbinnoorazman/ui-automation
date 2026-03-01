package integration

import (
	"context"

	"github.com/google/uuid"
)

// Store defines the interface for integration and issue link persistence operations.
type Store interface {
	// CreateIntegration creates a new integration in the store.
	CreateIntegration(ctx context.Context, integration *Integration) error

	// GetIntegrationByID retrieves an integration by its ID.
	GetIntegrationByID(ctx context.Context, id uuid.UUID) (*Integration, error)

	// ListIntegrationsByUser retrieves all integrations for a user.
	ListIntegrationsByUser(ctx context.Context, userID uuid.UUID) ([]*Integration, error)

	// UpdateIntegration updates an integration with the given setters.
	UpdateIntegration(ctx context.Context, id uuid.UUID, setters ...IntegrationSetter) error

	// DeleteIntegration deletes an integration by its ID.
	DeleteIntegration(ctx context.Context, id uuid.UUID) error

	// CreateIssueLink creates a new issue link in the store.
	CreateIssueLink(ctx context.Context, link *IssueLink) error

	// GetIssueLinkByID retrieves an issue link by its ID.
	GetIssueLinkByID(ctx context.Context, id uuid.UUID) (*IssueLink, error)

	// ListIssueLinksByTestRun retrieves all issue links for a test run.
	ListIssueLinksByTestRun(ctx context.Context, testRunID uuid.UUID) ([]*IssueLink, error)

	// UpdateIssueLink updates an issue link with the given setters.
	UpdateIssueLink(ctx context.Context, id uuid.UUID, setters ...IssueLinkSetter) error

	// DeleteIssueLink deletes an issue link by its ID.
	DeleteIssueLink(ctx context.Context, id uuid.UUID) error
}

// IntegrationSetter is a function that updates an integration field.
type IntegrationSetter func(*Integration) error

// IssueLinkSetter is a function that updates an issue link field.
type IssueLinkSetter func(*IssueLink) error
