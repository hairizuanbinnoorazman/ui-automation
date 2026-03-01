package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/ui-automation/issuetracker"
)

const defaultBaseURL = "https://api.github.com"

// Client implements the issuetracker.Client interface for GitHub.
type Client struct {
	httpClient   *http.Client
	token        string
	baseURL      string
	defaultOwner string
	defaultRepo  string
}

// NewClient creates a new GitHub issue tracker client.
func NewClient(credentials map[string]string) (*Client, error) {
	token, ok := credentials["token"]
	if !ok || token == "" {
		return nil, fmt.Errorf("github: token is required")
	}

	baseURL := defaultBaseURL
	if u, ok := credentials["base_url"]; ok && u != "" {
		baseURL = strings.TrimRight(u, "/")
	}

	return &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		token:        token,
		baseURL:      baseURL,
		defaultOwner: credentials["default_owner"],
		defaultRepo:  credentials["default_repo"],
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("github: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("github: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// parseExternalID parses "owner/repo#number" into owner, repo, and number.
func parseExternalID(externalID string) (owner, repo string, number int, err error) {
	parts := strings.SplitN(externalID, "#", 2)
	if len(parts) != 2 {
		return "", "", 0, fmt.Errorf("github: invalid external ID format, expected owner/repo#number")
	}

	repoParts := strings.SplitN(parts[0], "/", 2)
	if len(repoParts) != 2 {
		return "", "", 0, fmt.Errorf("github: invalid external ID format, expected owner/repo#number")
	}

	number, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", "", 0, fmt.Errorf("github: invalid issue number in external ID: %w", err)
	}

	return repoParts[0], repoParts[1], number, nil
}

// parseOwnerRepo parses "owner/repo" into owner and repo.
func parseOwnerRepo(repository string) (owner, repo string, err error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("github: invalid repository format, expected owner/repo")
	}
	return parts[0], parts[1], nil
}

type githubIssue struct {
	Number    int          `json:"number"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	State     string       `json:"state"`
	HTMLURL   string       `json:"html_url"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Labels    []labelEntry `json:"labels"`
}

type labelEntry struct {
	Name string `json:"name"`
}

func (c *Client) toIssue(gi *githubIssue, owner, repo string) *issuetracker.Issue {
	return &issuetracker.Issue{
		ExternalID:  fmt.Sprintf("%s/%s#%d", owner, repo, gi.Number),
		Title:       gi.Title,
		Description: gi.Body,
		Status:      gi.State,
		URL:         gi.HTMLURL,
		Provider:    issuetracker.ProviderGitHub,
		CreatedAt:   gi.CreatedAt,
		UpdatedAt:   gi.UpdatedAt,
	}
}

// CreateIssue creates a new GitHub issue.
func (c *Client) CreateIssue(ctx context.Context, input issuetracker.CreateIssueInput) (*issuetracker.Issue, error) {
	repository := input.Repository
	if repository == "" {
		if c.defaultOwner != "" && c.defaultRepo != "" {
			repository = c.defaultOwner + "/" + c.defaultRepo
		} else {
			return nil, fmt.Errorf("github: repository is required")
		}
	}

	owner, repo, err := parseOwnerRepo(repository)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"title": input.Title,
		"body":  input.Description,
	}
	if len(input.Labels) > 0 {
		reqBody["labels"] = input.Labels
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)
	resp, err := c.doRequest(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: create issue failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gi githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&gi); err != nil {
		return nil, fmt.Errorf("github: failed to decode response: %w", err)
	}

	return c.toIssue(&gi, owner, repo), nil
}

// GetIssue gets a GitHub issue by external ID.
func (c *Client) GetIssue(ctx context.Context, externalID string) (*issuetracker.Issue, error) {
	owner, repo, number, err := parseExternalID(externalID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, owner, repo, number)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, issuetracker.ErrIssueNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: get issue failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gi githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&gi); err != nil {
		return nil, fmt.Errorf("github: failed to decode response: %w", err)
	}

	return c.toIssue(&gi, owner, repo), nil
}

// ListIssues lists GitHub issues.
func (c *Client) ListIssues(ctx context.Context, input issuetracker.ListIssuesInput) ([]*issuetracker.Issue, int, error) {
	repository := input.Repository
	if repository == "" {
		if c.defaultOwner != "" && c.defaultRepo != "" {
			repository = c.defaultOwner + "/" + c.defaultRepo
		} else {
			return nil, 0, fmt.Errorf("github: repository is required")
		}
	}

	owner, repo, err := parseOwnerRepo(repository)
	if err != nil {
		return nil, 0, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues?per_page=%d&page=%d",
		c.baseURL, owner, repo,
		input.Limit, (input.Offset/max(input.Limit, 1))+1)

	if input.Status != "" {
		url += "&state=" + input.Status
	}

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("github: list issues failed with status %d: %s", resp.StatusCode, string(body))
	}

	var issues []githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, 0, fmt.Errorf("github: failed to decode response: %w", err)
	}

	result := make([]*issuetracker.Issue, 0, len(issues))
	for i := range issues {
		result = append(result, c.toIssue(&issues[i], owner, repo))
	}

	// GitHub API doesn't return total count in list endpoint; approximate with result length.
	return result, len(result), nil
}

// ResolveIssue closes a GitHub issue.
func (c *Client) ResolveIssue(ctx context.Context, externalID string, input issuetracker.ResolveInput) (*issuetracker.Issue, error) {
	owner, repo, number, err := parseExternalID(externalID)
	if err != nil {
		return nil, err
	}

	// Close the issue
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, owner, repo, number)
	reqBody := map[string]interface{}{
		"state": "closed",
	}
	resp, err := c.doRequest(ctx, http.MethodPatch, url, reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, issuetracker.ErrIssueNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: resolve issue failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gi githubIssue
	if err := json.NewDecoder(resp.Body).Decode(&gi); err != nil {
		return nil, fmt.Errorf("github: failed to decode response: %w", err)
	}

	// Add comment if provided
	if input.Comment != "" {
		commentURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.baseURL, owner, repo, number)
		commentBody := map[string]string{"body": input.Comment}
		commentResp, err := c.doRequest(ctx, http.MethodPost, commentURL, commentBody)
		if err != nil {
			// Issue was closed but comment failed; still return the issue.
			return c.toIssue(&gi, owner, repo), nil
		}
		commentResp.Body.Close()
	}

	return c.toIssue(&gi, owner, repo), nil
}

// ValidateConnection validates the GitHub connection by fetching the authenticated user.
func (c *Client) ValidateConnection(ctx context.Context) error {
	url := fmt.Sprintf("%s/user", c.baseURL)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", issuetracker.ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: unexpected status %d", issuetracker.ErrConnectionFailed, resp.StatusCode)
	}

	return nil
}
