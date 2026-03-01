package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hairizuanbinnoorazman/ui-automation/issuetracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := NewClient(map[string]string{
		"token":         "test-token",
		"base_url":      server.URL,
		"default_owner": "owner",
		"default_repo":  "repo",
	})
	require.NoError(t, err)
	return client, server
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials map[string]string
		wantErr     bool
	}{
		{
			name:        "valid credentials",
			credentials: map[string]string{"token": "ghp_test"},
			wantErr:     false,
		},
		{
			name:        "missing token",
			credentials: map[string]string{},
			wantErr:     true,
		},
		{
			name:        "empty token",
			credentials: map[string]string{"token": ""},
			wantErr:     true,
		},
		{
			name:        "with base_url",
			credentials: map[string]string{"token": "ghp_test", "base_url": "https://custom.github.com"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, err := NewClient(tt.credentials)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestCreateIssue(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/repos/owner/repo/issues")
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"number":     42,
			"title":      "Test Issue",
			"body":       "Test Description",
			"state":      "open",
			"html_url":   "https://github.com/owner/repo/issues/42",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	issue, err := client.CreateIssue(context.Background(), issuetracker.CreateIssueInput{
		Title:       "Test Issue",
		Description: "Test Description",
	})
	require.NoError(t, err)
	assert.Equal(t, "owner/repo#42", issue.ExternalID)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "Test Description", issue.Description)
	assert.Equal(t, "open", issue.Status)
	assert.Equal(t, issuetracker.ProviderGitHub, issue.Provider)
}

func TestCreateIssueWithRepository(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/repos/other/project/issues")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"number":     1,
			"title":      "Custom Repo Issue",
			"body":       "",
			"state":      "open",
			"html_url":   "https://github.com/other/project/issues/1",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	issue, err := client.CreateIssue(context.Background(), issuetracker.CreateIssueInput{
		Title:      "Custom Repo Issue",
		Repository: "other/project",
	})
	require.NoError(t, err)
	assert.Equal(t, "other/project#1", issue.ExternalID)
}

func TestCreateIssueServerError(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	_, err := client.CreateIssue(context.Background(), issuetracker.CreateIssueInput{
		Title: "Fail",
	})
	assert.Error(t, err)
}

func TestGetIssue(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/repos/owner/repo/issues/42", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"number":     42,
			"title":      "Existing Issue",
			"body":       "Some description",
			"state":      "open",
			"html_url":   "https://github.com/owner/repo/issues/42",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-02T00:00:00Z",
		})
	}))
	defer server.Close()

	issue, err := client.GetIssue(context.Background(), "owner/repo#42")
	require.NoError(t, err)
	assert.Equal(t, "owner/repo#42", issue.ExternalID)
	assert.Equal(t, "Existing Issue", issue.Title)
	assert.Equal(t, "open", issue.Status)
}

func TestGetIssueNotFound(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := client.GetIssue(context.Background(), "owner/repo#999")
	assert.ErrorIs(t, err, issuetracker.ErrIssueNotFound)
}

func TestGetIssueInvalidExternalID(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	}))
	defer server.Close()

	_, err := client.GetIssue(context.Background(), "invalid-format")
	assert.Error(t, err)
}

func TestListIssues(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/repos/owner/repo/issues")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"number":     1,
				"title":      "First Issue",
				"body":       "",
				"state":      "open",
				"html_url":   "https://github.com/owner/repo/issues/1",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z",
			},
			{
				"number":     2,
				"title":      "Second Issue",
				"body":       "",
				"state":      "closed",
				"html_url":   "https://github.com/owner/repo/issues/2",
				"created_at": "2024-01-02T00:00:00Z",
				"updated_at": "2024-01-02T00:00:00Z",
			},
		})
	}))
	defer server.Close()

	issues, total, err := client.ListIssues(context.Background(), issuetracker.ListIssuesInput{
		Limit:  20,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, issues, 2)
	assert.Equal(t, "owner/repo#1", issues[0].ExternalID)
	assert.Equal(t, "owner/repo#2", issues[1].ExternalID)
}

func TestResolveIssue(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			assert.Equal(t, "/repos/owner/repo/issues/42", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"number":     42,
				"title":      "Resolved Issue",
				"body":       "",
				"state":      "closed",
				"html_url":   "https://github.com/owner/repo/issues/42",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-03T00:00:00Z",
			})
			return
		}
		// Comment POST
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	issue, err := client.ResolveIssue(context.Background(), "owner/repo#42", issuetracker.ResolveInput{
		Comment: "Fixed in latest release",
	})
	require.NoError(t, err)
	assert.Equal(t, "closed", issue.Status)
	assert.Equal(t, "owner/repo#42", issue.ExternalID)
}

func TestValidateConnection(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"login": "testuser"})
	}))
	defer server.Close()

	err := client.ValidateConnection(context.Background())
	assert.NoError(t, err)
}

func TestValidateConnectionFailed(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	err := client.ValidateConnection(context.Background())
	assert.Error(t, err)
	assert.ErrorIs(t, err, issuetracker.ErrConnectionFailed)
}

func TestParseExternalID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		externalID string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid",
			externalID: "owner/repo#42",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantNumber: 42,
		},
		{
			name:       "missing hash",
			externalID: "owner/repo",
			wantErr:    true,
		},
		{
			name:       "missing slash",
			externalID: "repo#42",
			wantErr:    true,
		},
		{
			name:       "non-numeric number",
			externalID: "owner/repo#abc",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, number, err := parseExternalID(tt.externalID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
				assert.Equal(t, tt.wantNumber, number)
			}
		})
	}
}
