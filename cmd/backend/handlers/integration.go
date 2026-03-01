package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hairizuanbinnoorazman/ui-automation/integration"
	"github.com/hairizuanbinnoorazman/ui-automation/issuetracker"
	"github.com/hairizuanbinnoorazman/ui-automation/logger"
	"github.com/hairizuanbinnoorazman/ui-automation/project"
	"github.com/hairizuanbinnoorazman/ui-automation/testprocedure"
	"github.com/hairizuanbinnoorazman/ui-automation/testrun"
)

// IntegrationHandler handles integration and issue link requests.
type IntegrationHandler struct {
	integrationStore   integration.Store
	clientFactory      issuetracker.ClientFactory
	encryptionKey      []byte
	testRunStore       testrun.Store
	testProcedureStore testprocedure.Store
	projectStore       project.Store
	logger             logger.Logger
}

// NewIntegrationHandler creates a new integration handler.
func NewIntegrationHandler(
	integrationStore integration.Store,
	clientFactory issuetracker.ClientFactory,
	encryptionKey []byte,
	testRunStore testrun.Store,
	testProcedureStore testprocedure.Store,
	projectStore project.Store,
	log logger.Logger,
) *IntegrationHandler {
	return &IntegrationHandler{
		integrationStore:   integrationStore,
		clientFactory:      clientFactory,
		encryptionKey:      encryptionKey,
		testRunStore:       testRunStore,
		testProcedureStore: testProcedureStore,
		projectStore:       projectStore,
		logger:             log,
	}
}

// checkIntegrationOwnership verifies that the authenticated user owns the integration.
func (h *IntegrationHandler) checkIntegrationOwnership(w http.ResponseWriter, r *http.Request, integrationID uuid.UUID) (*integration.Integration, bool) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return nil, false
	}

	integ, err := h.integrationStore.GetIntegrationByID(r.Context(), integrationID)
	if err != nil {
		if errors.Is(err, integration.ErrIntegrationNotFound) {
			respondError(w, http.StatusNotFound, "integration not found")
			return nil, false
		}
		respondError(w, http.StatusInternalServerError, "failed to verify integration")
		return nil, false
	}

	if integ.UserID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return nil, false
	}

	return integ, true
}

// checkRunOwnership verifies that the authenticated user owns the project
// associated with the given test run via test run -> procedure -> project -> owner.
func (h *IntegrationHandler) checkRunOwnership(w http.ResponseWriter, r *http.Request, runID uuid.UUID) bool {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return false
	}

	tr, err := h.testRunStore.GetByID(r.Context(), runID)
	if err != nil {
		if errors.Is(err, testrun.ErrTestRunNotFound) {
			respondError(w, http.StatusNotFound, "test run not found")
			return false
		}
		respondError(w, http.StatusInternalServerError, "failed to verify test run")
		return false
	}

	tp, err := h.testProcedureStore.GetByID(r.Context(), tr.TestProcedureID)
	if err != nil {
		if errors.Is(err, testprocedure.ErrTestProcedureNotFound) {
			respondError(w, http.StatusNotFound, "test procedure not found")
			return false
		}
		respondError(w, http.StatusInternalServerError, "failed to verify test procedure")
		return false
	}

	proj, err := h.projectStore.GetByID(r.Context(), tp.ProjectID)
	if err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
			return false
		}
		respondError(w, http.StatusInternalServerError, "failed to verify project")
		return false
	}

	if proj.OwnerID != userID {
		respondError(w, http.StatusForbidden, "access denied")
		return false
	}

	return true
}

// credentialEntry represents a single credential key-value pair from the frontend.
type credentialEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateIntegrationRequest represents the request body for creating an integration.
type CreateIntegrationRequest struct {
	Name        string                    `json:"name"`
	Provider    issuetracker.ProviderType `json:"provider"`
	Credentials []credentialEntry         `json:"credentials"`
}

// toMap converts a credential entry list to a map.
func credentialsToMap(entries []credentialEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}

// UpdateIntegrationRequest represents the request body for updating an integration.
type UpdateIntegrationRequest struct {
	Name        *string           `json:"name,omitempty"`
	IsActive    *bool             `json:"is_active,omitempty"`
	Credentials []credentialEntry `json:"credentials,omitempty"`
}

// CreateAndLinkIssueRequest represents the request body for creating and linking an issue.
type CreateAndLinkIssueRequest struct {
	IntegrationID string `json:"integration_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ProjectKey    string `json:"project_key"`
	IssueType     string `json:"issue_type"`
	Repository    string `json:"repository"`
	Labels      []string `json:"labels"`
}

// LinkExistingIssueRequest represents the request body for linking an existing issue.
type LinkExistingIssueRequest struct {
	IntegrationID string `json:"integration_id"`
	ExternalID    string `json:"external_id"`
}

// ResolveLinkedIssueRequest represents the request body for resolving a linked issue.
type ResolveLinkedIssueRequest struct {
	Resolution string `json:"resolution"`
	Comment    string `json:"comment"`
}

// SearchExternalIssuesRequest represents query parameters for searching issues.
type SearchExternalIssuesRequest struct {
	ProjectKey string `json:"project_key"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
	Query      string `json:"query"`
}

// IntegrationResponse represents an integration in API responses (without encrypted credentials).
type IntegrationResponse struct {
	ID        uuid.UUID                 `json:"id"`
	UserID    uuid.UUID                 `json:"user_id"`
	Name      string                    `json:"name"`
	Provider  issuetracker.ProviderType `json:"provider"`
	IsActive  bool                      `json:"is_active"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
}

func toIntegrationResponse(integ *integration.Integration) IntegrationResponse {
	return IntegrationResponse{
		ID:        integ.ID,
		UserID:    integ.UserID,
		Name:      integ.Name,
		Provider:  integ.Provider,
		IsActive:  integ.IsActive,
		CreatedAt: integ.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: integ.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListIntegrations handles GET /integrations.
func (h *IntegrationHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	integrations, err := h.integrationStore.ListIntegrationsByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list integrations", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list integrations")
		return
	}

	result := make([]IntegrationResponse, len(integrations))
	for i, integ := range integrations {
		result[i] = toIntegrationResponse(integ)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items": result,
		"total": len(result),
	})
}

// CreateIntegration handles POST /integrations.
func (h *IntegrationHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req CreateIntegrationRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	if !req.Provider.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid provider type")
		return
	}

	if len(req.Credentials) == 0 {
		respondError(w, http.StatusBadRequest, "credentials are required")
		return
	}

	encrypted, err := integration.EncryptCredentials(h.encryptionKey, credentialsToMap(req.Credentials))
	if err != nil {
		h.logger.Error(r.Context(), "failed to encrypt credentials", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to encrypt credentials")
		return
	}

	integ := &integration.Integration{
		UserID:               userID,
		Name:                 req.Name,
		Provider:             req.Provider,
		EncryptedCredentials: encrypted,
		IsActive:             true,
	}

	if err := h.integrationStore.CreateIntegration(r.Context(), integ); err != nil {
		h.logger.Error(r.Context(), "failed to create integration", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create integration")
		return
	}

	respondJSON(w, http.StatusCreated, toIntegrationResponse(integ))
}

// GetIntegration handles GET /integrations/{integration_id}.
func (h *IntegrationHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID, ok := parseUUIDOrRespond(w, r, "integration_id", "integration")
	if !ok {
		return
	}

	integ, ok := h.checkIntegrationOwnership(w, r, integrationID)
	if !ok {
		return
	}

	respondJSON(w, http.StatusOK, toIntegrationResponse(integ))
}

// UpdateIntegration handles PUT /integrations/{integration_id}.
func (h *IntegrationHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID, ok := parseUUIDOrRespond(w, r, "integration_id", "integration")
	if !ok {
		return
	}

	if _, ok := h.checkIntegrationOwnership(w, r, integrationID); !ok {
		return
	}

	var req UpdateIntegrationRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var setters []integration.IntegrationSetter

	if req.Name != nil {
		setters = append(setters, integration.SetName(*req.Name))
	}

	if req.IsActive != nil {
		setters = append(setters, integration.SetIsActive(*req.IsActive))
	}

	if len(req.Credentials) > 0 {
		encrypted, err := integration.EncryptCredentials(h.encryptionKey, credentialsToMap(req.Credentials))
		if err != nil {
			h.logger.Error(r.Context(), "failed to encrypt credentials", map[string]interface{}{
				"error": err.Error(),
			})
			respondError(w, http.StatusInternalServerError, "failed to encrypt credentials")
			return
		}
		setters = append(setters, integration.SetEncryptedCredentials(encrypted))
	}

	if len(setters) == 0 {
		respondError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := h.integrationStore.UpdateIntegration(r.Context(), integrationID, setters...); err != nil {
		h.logger.Error(r.Context(), "failed to update integration", map[string]interface{}{
			"error":          err.Error(),
			"integration_id": integrationID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to update integration")
		return
	}

	updated, err := h.integrationStore.GetIntegrationByID(r.Context(), integrationID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get updated integration")
		return
	}

	respondJSON(w, http.StatusOK, toIntegrationResponse(updated))
}

// DeleteIntegration handles DELETE /integrations/{integration_id}.
func (h *IntegrationHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	integrationID, ok := parseUUIDOrRespond(w, r, "integration_id", "integration")
	if !ok {
		return
	}

	if _, ok := h.checkIntegrationOwnership(w, r, integrationID); !ok {
		return
	}

	if err := h.integrationStore.DeleteIntegration(r.Context(), integrationID); err != nil {
		h.logger.Error(r.Context(), "failed to delete integration", map[string]interface{}{
			"error":          err.Error(),
			"integration_id": integrationID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to delete integration")
		return
	}

	respondSuccess(w, "integration deleted successfully")
}

// TestConnection handles POST /integrations/{integration_id}/test.
func (h *IntegrationHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	integrationID, ok := parseUUIDOrRespond(w, r, "integration_id", "integration")
	if !ok {
		return
	}

	integ, ok := h.checkIntegrationOwnership(w, r, integrationID)
	if !ok {
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		h.logger.Error(r.Context(), "failed to decrypt credentials", map[string]interface{}{
			"error":          err.Error(),
			"integration_id": integrationID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create issue tracker client", map[string]interface{}{
			"error":    err.Error(),
			"provider": string(integ.Provider),
		})
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	if err := client.ValidateConnection(r.Context()); err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "connection successful",
	})
}

// ListIssueLinks handles GET /runs/{run_id}/issues.
func (h *IntegrationHandler) ListIssueLinks(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	links, err := h.integrationStore.ListIssueLinksByTestRun(r.Context(), runID)
	if err != nil {
		h.logger.Error(r.Context(), "failed to list issue links", map[string]interface{}{
			"error":       err.Error(),
			"test_run_id": runID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to list issue links")
		return
	}

	respondJSON(w, http.StatusOK, links)
}

// CreateAndLinkIssue handles POST /runs/{run_id}/issues.
func (h *IntegrationHandler) CreateAndLinkIssue(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	var req CreateAndLinkIssueRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	integrationID, err := uuid.Parse(req.IntegrationID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid integration_id")
		return
	}

	integ, ok := h.checkIntegrationOwnership(w, r, integrationID)
	if !ok {
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		h.logger.Error(r.Context(), "failed to decrypt credentials", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create issue tracker client", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	issue, err := client.CreateIssue(r.Context(), issuetracker.CreateIssueInput{
		Title:       req.Title,
		Description: req.Description,
		ProjectKey:  req.ProjectKey,
		IssueType:   req.IssueType,
		Repository:  req.Repository,
		Labels:      req.Labels,
	})
	if err != nil {
		h.logger.Error(r.Context(), "failed to create issue", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create issue in external tracker")
		return
	}

	link := &integration.IssueLink{
		TestRunID:     runID,
		IntegrationID: integrationID,
		ExternalID:    issue.ExternalID,
		Title:         issue.Title,
		Status:        issue.Status,
		URL:           issue.URL,
		Provider:      integ.Provider,
	}

	if err := h.integrationStore.CreateIssueLink(r.Context(), link); err != nil {
		h.logger.Error(r.Context(), "failed to create issue link", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to link issue")
		return
	}

	respondJSON(w, http.StatusCreated, link)
}

// LinkExistingIssue handles POST /runs/{run_id}/issues/link.
func (h *IntegrationHandler) LinkExistingIssue(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	var req LinkExistingIssueRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ExternalID == "" {
		respondError(w, http.StatusBadRequest, "external_id is required")
		return
	}

	integrationID, err := uuid.Parse(req.IntegrationID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid integration_id")
		return
	}

	integ, ok := h.checkIntegrationOwnership(w, r, integrationID)
	if !ok {
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		h.logger.Error(r.Context(), "failed to decrypt credentials", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create issue tracker client", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	issue, err := client.GetIssue(r.Context(), req.ExternalID)
	if err != nil {
		if errors.Is(err, issuetracker.ErrIssueNotFound) {
			respondError(w, http.StatusNotFound, "issue not found in external tracker")
			return
		}
		h.logger.Error(r.Context(), "failed to get issue from external tracker", map[string]interface{}{
			"error":       err.Error(),
			"external_id": req.ExternalID,
		})
		respondError(w, http.StatusInternalServerError, "failed to get issue from external tracker")
		return
	}

	link := &integration.IssueLink{
		TestRunID:     runID,
		IntegrationID: integrationID,
		ExternalID:    issue.ExternalID,
		Title:         issue.Title,
		Status:        issue.Status,
		URL:           issue.URL,
		Provider:      integ.Provider,
	}

	if err := h.integrationStore.CreateIssueLink(r.Context(), link); err != nil {
		h.logger.Error(r.Context(), "failed to create issue link", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to link issue")
		return
	}

	respondJSON(w, http.StatusCreated, link)
}

// UnlinkIssue handles DELETE /runs/{run_id}/issues/{link_id}.
func (h *IntegrationHandler) UnlinkIssue(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	linkID, ok := parseUUIDOrRespond(w, r, "link_id", "issue link")
	if !ok {
		return
	}

	if err := h.integrationStore.DeleteIssueLink(r.Context(), linkID); err != nil {
		if errors.Is(err, integration.ErrIssueLinkNotFound) {
			respondError(w, http.StatusNotFound, "issue link not found")
			return
		}
		h.logger.Error(r.Context(), "failed to delete issue link", map[string]interface{}{
			"error":         err.Error(),
			"issue_link_id": linkID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to unlink issue")
		return
	}

	respondSuccess(w, "issue unlinked successfully")
}

// ResolveLinkedIssue handles POST /runs/{run_id}/issues/{link_id}/resolve.
func (h *IntegrationHandler) ResolveLinkedIssue(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	linkID, ok := parseUUIDOrRespond(w, r, "link_id", "issue link")
	if !ok {
		return
	}

	link, err := h.integrationStore.GetIssueLinkByID(r.Context(), linkID)
	if err != nil {
		if errors.Is(err, integration.ErrIssueLinkNotFound) {
			respondError(w, http.StatusNotFound, "issue link not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get issue link")
		return
	}

	var req ResolveLinkedIssueRequest
	if err := parseJSON(r, &req, h.logger); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	integ, err := h.integrationStore.GetIntegrationByID(r.Context(), link.IntegrationID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get integration")
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	issue, err := client.ResolveIssue(r.Context(), link.ExternalID, issuetracker.ResolveInput{
		Resolution: req.Resolution,
		Comment:    req.Comment,
	})
	if err != nil {
		if errors.Is(err, issuetracker.ErrIssueNotFound) {
			respondError(w, http.StatusNotFound, "issue not found in external tracker")
			return
		}
		h.logger.Error(r.Context(), "failed to resolve issue", map[string]interface{}{
			"error":       err.Error(),
			"external_id": link.ExternalID,
		})
		respondError(w, http.StatusInternalServerError, "failed to resolve issue")
		return
	}

	// Update the link with latest status.
	if err := h.integrationStore.UpdateIssueLink(r.Context(), linkID,
		integration.SetStatus(issue.Status),
		integration.SetTitle(issue.Title),
		integration.SetURL(issue.URL),
	); err != nil {
		h.logger.Warn(r.Context(), "failed to update issue link after resolve", map[string]interface{}{
			"error":         err.Error(),
			"issue_link_id": linkID.String(),
		})
	}

	updatedLink, err := h.integrationStore.GetIssueLinkByID(r.Context(), linkID)
	if err != nil {
		respondJSON(w, http.StatusOK, link)
		return
	}

	respondJSON(w, http.StatusOK, updatedLink)
}

// SyncIssueStatus handles POST /runs/{run_id}/issues/{link_id}/sync.
func (h *IntegrationHandler) SyncIssueStatus(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseUUIDOrRespond(w, r, "run_id", "test run")
	if !ok {
		return
	}

	if !h.checkRunOwnership(w, r, runID) {
		return
	}

	linkID, ok := parseUUIDOrRespond(w, r, "link_id", "issue link")
	if !ok {
		return
	}

	link, err := h.integrationStore.GetIssueLinkByID(r.Context(), linkID)
	if err != nil {
		if errors.Is(err, integration.ErrIssueLinkNotFound) {
			respondError(w, http.StatusNotFound, "issue link not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get issue link")
		return
	}

	integ, err := h.integrationStore.GetIntegrationByID(r.Context(), link.IntegrationID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get integration")
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	issue, err := client.GetIssue(r.Context(), link.ExternalID)
	if err != nil {
		if errors.Is(err, issuetracker.ErrIssueNotFound) {
			respondError(w, http.StatusNotFound, "issue not found in external tracker")
			return
		}
		h.logger.Error(r.Context(), "failed to get issue from external tracker", map[string]interface{}{
			"error":       err.Error(),
			"external_id": link.ExternalID,
		})
		respondError(w, http.StatusInternalServerError, "failed to sync issue status")
		return
	}

	if err := h.integrationStore.UpdateIssueLink(r.Context(), linkID,
		integration.SetStatus(issue.Status),
		integration.SetTitle(issue.Title),
		integration.SetURL(issue.URL),
	); err != nil {
		h.logger.Error(r.Context(), "failed to update issue link", map[string]interface{}{
			"error":         err.Error(),
			"issue_link_id": linkID.String(),
		})
		respondError(w, http.StatusInternalServerError, "failed to update issue link")
		return
	}

	updatedLink, err := h.integrationStore.GetIssueLinkByID(r.Context(), linkID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get updated issue link")
		return
	}

	respondJSON(w, http.StatusOK, updatedLink)
}

// SearchExternalIssues handles GET /integrations/{integration_id}/issues.
func (h *IntegrationHandler) SearchExternalIssues(w http.ResponseWriter, r *http.Request) {
	integrationID, ok := parseUUIDOrRespond(w, r, "integration_id", "integration")
	if !ok {
		return
	}

	integ, ok := h.checkIntegrationOwnership(w, r, integrationID)
	if !ok {
		return
	}

	creds, err := integration.DecryptCredentials(h.encryptionKey, integ.EncryptedCredentials)
	if err != nil {
		h.logger.Error(r.Context(), "failed to decrypt credentials", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to decrypt credentials")
		return
	}

	client, err := h.clientFactory.NewClient(integ.Provider, creds)
	if err != nil {
		h.logger.Error(r.Context(), "failed to create issue tracker client", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	query := r.URL.Query()
	issues, total, err := client.ListIssues(r.Context(), issuetracker.ListIssuesInput{
		ProjectKey: query.Get("project_key"),
		Repository: query.Get("repository"),
		Status:     query.Get("status"),
		Query:      query.Get("query"),
		Limit:      20,
		Offset:     0,
	})
	if err != nil {
		h.logger.Error(r.Context(), "failed to search issues", map[string]interface{}{
			"error": err.Error(),
		})
		respondError(w, http.StatusInternalServerError, "failed to search issues")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items": issues,
		"total": total,
	})
}
