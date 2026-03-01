package issuetracker

import (
	"context"
	"errors"
	"time"
)

var (
	ErrIssueNotFound    = errors.New("issue not found")
	ErrInvalidProvider  = errors.New("invalid provider type")
	ErrConnectionFailed = errors.New("connection validation failed")
)

type ProviderType string

const (
	ProviderJira   ProviderType = "jira"
	ProviderGitHub ProviderType = "github"
)

func (p ProviderType) IsValid() bool {
	return p == ProviderJira || p == ProviderGitHub
}

type Issue struct {
	ExternalID  string       `json:"external_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	URL         string       `json:"url"`
	Provider    ProviderType `json:"provider"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type CreateIssueInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ProjectKey  string   `json:"project_key"`
	IssueType   string   `json:"issue_type"`
	Repository  string   `json:"repository"`
	Labels      []string `json:"labels"`
}

type ListIssuesInput struct {
	ProjectKey string `json:"project_key"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
	Query      string `json:"query"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type ResolveInput struct {
	Resolution string `json:"resolution"`
	Comment    string `json:"comment"`
}

type Client interface {
	CreateIssue(ctx context.Context, input CreateIssueInput) (*Issue, error)
	GetIssue(ctx context.Context, externalID string) (*Issue, error)
	ListIssues(ctx context.Context, input ListIssuesInput) ([]*Issue, int, error)
	ResolveIssue(ctx context.Context, externalID string, input ResolveInput) (*Issue, error)
	ValidateConnection(ctx context.Context) error
}
