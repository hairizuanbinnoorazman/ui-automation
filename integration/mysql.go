package integration

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"gorm.io/gorm"
)

// MySQLStore implements the Store interface using GORM and MySQL.
type MySQLStore struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMySQLStore creates a new MySQL-backed integration store.
func NewMySQLStore(db *gorm.DB, log logger.Logger) *MySQLStore {
	return &MySQLStore{
		db:     db,
		logger: log,
	}
}

// CreateIntegration creates a new integration in the database.
func (s *MySQLStore) CreateIntegration(ctx context.Context, integration *Integration) error {
	if err := integration.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(integration).Error; err != nil {
		s.logger.Error(ctx, "failed to create integration", map[string]interface{}{
			"error":   err.Error(),
			"user_id": integration.UserID.String(),
		})
		return err
	}

	s.logger.Info(ctx, "integration created", map[string]interface{}{
		"integration_id": integration.ID.String(),
		"user_id":        integration.UserID.String(),
	})

	return nil
}

// GetIntegrationByID retrieves an integration by its ID.
func (s *MySQLStore) GetIntegrationByID(ctx context.Context, id uuid.UUID) (*Integration, error) {
	var integ Integration
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&integ).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIntegrationNotFound
		}
		s.logger.Error(ctx, "failed to get integration by ID", map[string]interface{}{
			"error":          err.Error(),
			"integration_id": id.String(),
		})
		return nil, err
	}

	return &integ, nil
}

// ListIntegrationsByUser retrieves all integrations for a user.
func (s *MySQLStore) ListIntegrationsByUser(ctx context.Context, userID uuid.UUID) ([]*Integration, error) {
	var integrations []*Integration
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&integrations).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list integrations by user", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		return nil, err
	}

	return integrations, nil
}

// UpdateIntegration updates an integration with the given setters.
func (s *MySQLStore) UpdateIntegration(ctx context.Context, id uuid.UUID, setters ...IntegrationSetter) error {
	integ, err := s.GetIntegrationByID(ctx, id)
	if err != nil {
		return err
	}

	for _, setter := range setters {
		if err := setter(integ); err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Save(integ).Error; err != nil {
		s.logger.Error(ctx, "failed to update integration", map[string]interface{}{
			"error":          err.Error(),
			"integration_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "integration updated", map[string]interface{}{
		"integration_id": id.String(),
	})

	return nil
}

// DeleteIntegration deletes an integration by its ID.
func (s *MySQLStore) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&Integration{}, "id = ?", id)
	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete integration", map[string]interface{}{
			"error":          result.Error.Error(),
			"integration_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrIntegrationNotFound
	}

	s.logger.Info(ctx, "integration deleted", map[string]interface{}{
		"integration_id": id.String(),
	})

	return nil
}

// CreateIssueLink creates a new issue link in the database.
func (s *MySQLStore) CreateIssueLink(ctx context.Context, link *IssueLink) error {
	if err := link.Validate(); err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Create(link).Error; err != nil {
		s.logger.Error(ctx, "failed to create issue link", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": link.TestRunID.String(),
		})
		return err
	}

	s.logger.Info(ctx, "issue link created", map[string]interface{}{
		"issue_link_id": link.ID.String(),
		"test_run_id":   link.TestRunID.String(),
		"external_id":   link.ExternalID,
	})

	return nil
}

// GetIssueLinkByID retrieves an issue link by its ID.
func (s *MySQLStore) GetIssueLinkByID(ctx context.Context, id uuid.UUID) (*IssueLink, error) {
	var link IssueLink
	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&link).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIssueLinkNotFound
		}
		s.logger.Error(ctx, "failed to get issue link by ID", map[string]interface{}{
			"error":         err.Error(),
			"issue_link_id": id.String(),
		})
		return nil, err
	}

	return &link, nil
}

// ListIssueLinksByTestRun retrieves all issue links for a test run.
func (s *MySQLStore) ListIssueLinksByTestRun(ctx context.Context, testRunID uuid.UUID) ([]*IssueLink, error) {
	var links []*IssueLink
	err := s.db.WithContext(ctx).
		Where("test_run_id = ?", testRunID).
		Order("created_at DESC").
		Find(&links).Error

	if err != nil {
		s.logger.Error(ctx, "failed to list issue links by test run", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": testRunID.String(),
		})
		return nil, err
	}

	return links, nil
}

// UpdateIssueLink updates an issue link with the given setters.
func (s *MySQLStore) UpdateIssueLink(ctx context.Context, id uuid.UUID, setters ...IssueLinkSetter) error {
	link, err := s.GetIssueLinkByID(ctx, id)
	if err != nil {
		return err
	}

	for _, setter := range setters {
		if err := setter(link); err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Save(link).Error; err != nil {
		s.logger.Error(ctx, "failed to update issue link", map[string]interface{}{
			"error":         err.Error(),
			"issue_link_id": id.String(),
		})
		return err
	}

	s.logger.Info(ctx, "issue link updated", map[string]interface{}{
		"issue_link_id": id.String(),
	})

	return nil
}

// DeleteIssueLink deletes an issue link by its ID.
func (s *MySQLStore) DeleteIssueLink(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&IssueLink{}, "id = ?", id)
	if result.Error != nil {
		s.logger.Error(ctx, "failed to delete issue link", map[string]interface{}{
			"error":         result.Error.Error(),
			"issue_link_id": id.String(),
		})
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrIssueLinkNotFound
	}

	s.logger.Info(ctx, "issue link deleted", map[string]interface{}{
		"issue_link_id": id.String(),
	})

	return nil
}
