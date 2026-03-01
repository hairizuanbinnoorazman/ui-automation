package jira

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
		"url":             server.URL,
		"email":           "test@example.com",
		"api_token":       "test-api-token",
		"default_project": "TEST",
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
			name: "valid credentials",
			credentials: map[string]string{
				"url":       "https://example.atlassian.net",
				"email":     "user@example.com",
				"api_token": "token",
			},
			wantErr: false,
		},
		{
			name: "missing url",
			credentials: map[string]string{
				"email":     "user@example.com",
				"api_token": "token",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			credentials: map[string]string{
				"url":       "https://example.atlassian.net",
				"api_token": "token",
			},
			wantErr: true,
		},
		{
			name: "missing api_token",
			credentials: map[string]string{
				"url":   "https://example.atlassian.net",
				"email": "user@example.com",
			},
			wantErr: true,
		},
		{
			name:        "empty credentials",
			credentials: map[string]string{},
			wantErr:     true,
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

	callCount := 0
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Method == "POST" && r.URL.Path == "/rest/api/3/issue" {
			// Verify basic auth
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test@example.com", user)
			assert.Equal(t, "test-api-token", pass)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "10001",
				"key":  "TEST-1",
				"self": "https://example.atlassian.net/rest/api/3/issue/10001",
			})
			return
		}
		if r.Method == "GET" && r.URL.Path == "/rest/api/3/issue/TEST-1" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "10001",
				"key":  "TEST-1",
				"self": "https://example.atlassian.net/rest/api/3/issue/10001",
				"fields": map[string]interface{}{
					"summary":     "Test Issue",
					"description": "Test Description",
					"status":      map[string]string{"name": "To Do"},
					"issuetype":   map[string]string{"name": "Task"},
					"created":     "2024-01-01T00:00:00.000+0000",
					"updated":     "2024-01-01T00:00:00.000+0000",
				},
			})
			return
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	issue, err := client.CreateIssue(context.Background(), issuetracker.CreateIssueInput{
		Title:       "Test Issue",
		Description: "Test Description",
	})
	require.NoError(t, err)
	assert.Equal(t, "TEST-1", issue.ExternalID)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "To Do", issue.Status)
	assert.Equal(t, issuetracker.ProviderJira, issue.Provider)
}

func TestCreateIssueMissingProject(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	}))
	defer server.Close()

	client, err := NewClient(map[string]string{
		"url":       server.URL,
		"email":     "test@example.com",
		"api_token": "test-api-token",
	})
	require.NoError(t, err)

	_, err = client.CreateIssue(context.Background(), issuetracker.CreateIssueInput{
		Title: "No Project",
	})
	assert.Error(t, err)
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
		assert.Equal(t, "/rest/api/3/issue/TEST-42", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "10042",
			"key":  "TEST-42",
			"self": "https://example.atlassian.net/rest/api/3/issue/10042",
			"fields": map[string]interface{}{
				"summary":     "Existing Issue",
				"description": "Some description",
				"status":      map[string]string{"name": "In Progress"},
				"issuetype":   map[string]string{"name": "Bug"},
				"created":     "2024-01-01T00:00:00.000+0000",
				"updated":     "2024-01-02T00:00:00.000+0000",
			},
		})
	}))
	defer server.Close()

	issue, err := client.GetIssue(context.Background(), "TEST-42")
	require.NoError(t, err)
	assert.Equal(t, "TEST-42", issue.ExternalID)
	assert.Equal(t, "Existing Issue", issue.Title)
	assert.Equal(t, "In Progress", issue.Status)
	assert.Equal(t, issuetracker.ProviderJira, issue.Provider)
	assert.Contains(t, issue.URL, "/browse/TEST-42")
}

func TestGetIssueNotFound(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := client.GetIssue(context.Background(), "TEST-999")
	assert.ErrorIs(t, err, issuetracker.ErrIssueNotFound)
}

func TestListIssues(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/search")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issues": []map[string]interface{}{
				{
					"id":  "10001",
					"key": "TEST-1",
					"fields": map[string]interface{}{
						"summary":   "First Issue",
						"status":    map[string]string{"name": "To Do"},
						"issuetype": map[string]string{"name": "Task"},
						"created":   "2024-01-01T00:00:00.000+0000",
						"updated":   "2024-01-01T00:00:00.000+0000",
					},
				},
				{
					"id":  "10002",
					"key": "TEST-2",
					"fields": map[string]interface{}{
						"summary":   "Second Issue",
						"status":    map[string]string{"name": "Done"},
						"issuetype": map[string]string{"name": "Bug"},
						"created":   "2024-01-02T00:00:00.000+0000",
						"updated":   "2024-01-02T00:00:00.000+0000",
					},
				},
			},
			"total":      2,
			"maxResults": 20,
			"startAt":    0,
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
	assert.Equal(t, "TEST-1", issues[0].ExternalID)
	assert.Equal(t, "TEST-2", issues[1].ExternalID)
}

func TestResolveIssue(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/rest/api/3/issue/TEST-1/transitions":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"transitions": []map[string]interface{}{
					{
						"id":   "31",
						"name": "Done",
						"to":   map[string]string{"name": "Done"},
					},
				},
			})
		case r.Method == "POST" && r.URL.Path == "/rest/api/3/issue/TEST-1/transitions":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "POST" && r.URL.Path == "/rest/api/3/issue/TEST-1/comment":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{"id": "1"})
		case r.Method == "GET" && r.URL.Path == "/rest/api/3/issue/TEST-1":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":  "10001",
				"key": "TEST-1",
				"fields": map[string]interface{}{
					"summary":   "Resolved Issue",
					"status":    map[string]string{"name": "Done"},
					"issuetype": map[string]string{"name": "Task"},
					"created":   "2024-01-01T00:00:00.000+0000",
					"updated":   "2024-01-03T00:00:00.000+0000",
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	issue, err := client.ResolveIssue(context.Background(), "TEST-1", issuetracker.ResolveInput{
		Comment: "Fixed",
	})
	require.NoError(t, err)
	assert.Equal(t, "Done", issue.Status)
	assert.Equal(t, "TEST-1", issue.ExternalID)
}

func TestResolveIssueNotFound(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := client.ResolveIssue(context.Background(), "TEST-999", issuetracker.ResolveInput{})
	assert.ErrorIs(t, err, issuetracker.ErrIssueNotFound)
}

func TestValidateConnection(t *testing.T) {
	t.Parallel()
	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/rest/api/3/myself", r.URL.Path)

		// Verify basic auth
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", user)
		assert.Equal(t, "test-api-token", pass)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accountId":   "abc123",
			"displayName": "Test User",
		})
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
