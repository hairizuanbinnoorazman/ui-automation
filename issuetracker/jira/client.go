package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/ui-automation/issuetracker"
)

// Client implements the issuetracker.Client interface for Jira.
type Client struct {
	httpClient     *http.Client
	baseURL        string
	email          string
	apiToken       string
	defaultProject string
}

// NewClient creates a new Jira issue tracker client.
func NewClient(credentials map[string]string) (*Client, error) {
	baseURL, ok := credentials["url"]
	if !ok || baseURL == "" {
		return nil, fmt.Errorf("jira: url is required")
	}
	baseURL = strings.TrimRight(baseURL, "/")

	email, ok := credentials["email"]
	if !ok || email == "" {
		return nil, fmt.Errorf("jira: email is required")
	}

	apiToken, ok := credentials["api_token"]
	if !ok || apiToken == "" {
		return nil, fmt.Errorf("jira: api_token is required")
	}

	return &Client{
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		baseURL:        baseURL,
		email:          email,
		apiToken:       apiToken,
		defaultProject: credentials["default_project"],
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("jira: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("jira: failed to create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

type jiraIssueFields struct {
	Summary     string          `json:"summary"`
	Description interface{}     `json:"description"`
	Status      jiraStatus      `json:"status"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
	IssueType   jiraIssueType   `json:"issuetype"`
}

type jiraStatus struct {
	Name string `json:"name"`
}

type jiraIssueType struct {
	Name string `json:"name"`
}

type jiraIssue struct {
	ID     string          `json:"id"`
	Key    string          `json:"key"`
	Self   string          `json:"self"`
	Fields jiraIssueFields `json:"fields"`
}

func (c *Client) toIssue(ji *jiraIssue) *issuetracker.Issue {
	desc := ""
	if ji.Fields.Description != nil {
		if s, ok := ji.Fields.Description.(string); ok {
			desc = s
		}
	}

	created, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Created)
	updated, _ := time.Parse("2006-01-02T15:04:05.000-0700", ji.Fields.Updated)

	issueURL := fmt.Sprintf("%s/browse/%s", c.baseURL, ji.Key)

	return &issuetracker.Issue{
		ExternalID:  ji.Key,
		Title:       ji.Fields.Summary,
		Description: desc,
		Status:      ji.Fields.Status.Name,
		URL:         issueURL,
		Provider:    issuetracker.ProviderJira,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}

// CreateIssue creates a new Jira issue.
func (c *Client) CreateIssue(ctx context.Context, input issuetracker.CreateIssueInput) (*issuetracker.Issue, error) {
	projectKey := input.ProjectKey
	if projectKey == "" {
		projectKey = c.defaultProject
	}
	if projectKey == "" {
		return nil, fmt.Errorf("jira: project_key is required")
	}

	issueType := input.IssueType
	if issueType == "" {
		issueType = "Task"
	}

	reqBody := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": projectKey,
			},
			"summary":     input.Title,
			"description": input.Description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}

	apiURL := fmt.Sprintf("%s/rest/api/3/issue", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodPost, apiURL, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira: create issue failed with status %d: %s", resp.StatusCode, string(body))
	}

	var created struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("jira: failed to decode response: %w", err)
	}

	// Fetch the full issue to get all fields.
	return c.GetIssue(ctx, created.Key)
}

// GetIssue gets a Jira issue by key.
func (c *Client) GetIssue(ctx context.Context, externalID string) (*issuetracker.Issue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s", c.baseURL, externalID)
	resp, err := c.doRequest(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, issuetracker.ErrIssueNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira: get issue failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ji jiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&ji); err != nil {
		return nil, fmt.Errorf("jira: failed to decode response: %w", err)
	}

	return c.toIssue(&ji), nil
}

// ListIssues lists Jira issues using JQL search.
func (c *Client) ListIssues(ctx context.Context, input issuetracker.ListIssuesInput) ([]*issuetracker.Issue, int, error) {
	projectKey := input.ProjectKey
	if projectKey == "" {
		projectKey = c.defaultProject
	}

	// Build JQL query.
	var jqlParts []string
	if projectKey != "" {
		jqlParts = append(jqlParts, fmt.Sprintf("project = %s", projectKey))
	}
	if input.Status != "" {
		jqlParts = append(jqlParts, fmt.Sprintf("status = \"%s\"", input.Status))
	}
	if input.Query != "" {
		jqlParts = append(jqlParts, fmt.Sprintf("summary ~ \"%s\"", input.Query))
	}

	jql := strings.Join(jqlParts, " AND ")
	if jql == "" {
		jql = "order by created DESC"
	} else {
		jql += " order by created DESC"
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}

	apiURL := fmt.Sprintf("%s/rest/api/3/search?jql=%s&maxResults=%d&startAt=%d",
		c.baseURL, url.QueryEscape(jql), limit, input.Offset)

	resp, err := c.doRequest(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("jira: search issues failed with status %d: %s", resp.StatusCode, string(body))
	}

	var searchResult struct {
		Issues     []jiraIssue `json:"issues"`
		Total      int         `json:"total"`
		MaxResults int         `json:"maxResults"`
		StartAt    int         `json:"startAt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, 0, fmt.Errorf("jira: failed to decode response: %w", err)
	}

	result := make([]*issuetracker.Issue, 0, len(searchResult.Issues))
	for i := range searchResult.Issues {
		result = append(result, c.toIssue(&searchResult.Issues[i]))
	}

	return result, searchResult.Total, nil
}

// ResolveIssue transitions a Jira issue to Done/Resolved status.
func (c *Client) ResolveIssue(ctx context.Context, externalID string, input issuetracker.ResolveInput) (*issuetracker.Issue, error) {
	// Get available transitions.
	transURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.baseURL, externalID)
	resp, err := c.doRequest(ctx, http.MethodGet, transURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, issuetracker.ErrIssueNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira: get transitions failed with status %d: %s", resp.StatusCode, string(body))
	}

	var transResult struct {
		Transitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			To   struct {
				Name string `json:"name"`
			} `json:"to"`
		} `json:"transitions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&transResult); err != nil {
		return nil, fmt.Errorf("jira: failed to decode transitions: %w", err)
	}

	// Find a "Done" or "Resolved" transition.
	var transitionID string
	for _, t := range transResult.Transitions {
		name := strings.ToLower(t.To.Name)
		if name == "done" || name == "resolved" {
			transitionID = t.ID
			break
		}
	}

	if transitionID == "" {
		return nil, fmt.Errorf("jira: no Done or Resolved transition available for issue %s", externalID)
	}

	// Perform the transition.
	transBody := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionID,
		},
	}
	transResp, err := c.doRequest(ctx, http.MethodPost, transURL, transBody)
	if err != nil {
		return nil, err
	}
	defer transResp.Body.Close()

	if transResp.StatusCode != http.StatusNoContent && transResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(transResp.Body)
		return nil, fmt.Errorf("jira: transition failed with status %d: %s", transResp.StatusCode, string(body))
	}

	// Add comment if provided.
	if input.Comment != "" {
		commentURL := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", c.baseURL, externalID)
		commentBody := map[string]interface{}{
			"body": map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []map[string]interface{}{
					{
						"type": "paragraph",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": input.Comment,
							},
						},
					},
				},
			},
		}
		commentResp, err := c.doRequest(ctx, http.MethodPost, commentURL, commentBody)
		if err == nil {
			commentResp.Body.Close()
		}
	}

	// Fetch the updated issue.
	return c.GetIssue(ctx, externalID)
}

// ValidateConnection validates the Jira connection by fetching the authenticated user.
func (c *Client) ValidateConnection(ctx context.Context) error {
	apiURL := fmt.Sprintf("%s/rest/api/3/myself", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", issuetracker.ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: unexpected status %d", issuetracker.ErrConnectionFailed, resp.StatusCode)
	}

	return nil
}
