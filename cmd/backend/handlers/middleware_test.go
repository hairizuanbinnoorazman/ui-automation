package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hairizuanbinnoorazman/ui-automation/apitoken"
)

func TestRequireWriteScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		scope      string
		wantOK     bool
		wantStatus int
	}{
		{
			name:   "read_write scope passes",
			scope:  apitoken.ScopeReadWrite,
			wantOK: true,
		},
		{
			name:       "read_only scope returns 403",
			scope:      apitoken.ScopeReadOnly,
			wantOK:     false,
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "no scope in context defaults to read_write",
			scope:  "",
			wantOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tc.scope != "" {
				ctx := context.WithValue(req.Context(), ScopeKey, tc.scope)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			got := RequireWriteScope(w, req)
			if got != tc.wantOK {
				t.Errorf("RequireWriteScope() = %v, want %v", got, tc.wantOK)
			}
			if !tc.wantOK && w.Code != tc.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestWriteScopeMiddleware(t *testing.T) {
	t.Parallel()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		method     string
		scope      string
		wantStatus int
	}{
		{
			name:       "GET with read_only passes",
			method:     http.MethodGet,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET with read_write passes",
			method:     http.MethodGet,
			scope:      apitoken.ScopeReadWrite,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST with read_write passes",
			method:     http.MethodPost,
			scope:      apitoken.ScopeReadWrite,
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST with read_only blocked",
			method:     http.MethodPost,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "PUT with read_only blocked",
			method:     http.MethodPut,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "DELETE with read_only blocked",
			method:     http.MethodDelete,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "PATCH with read_only blocked",
			method:     http.MethodPatch,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "HEAD with read_only passes",
			method:     http.MethodHead,
			scope:      apitoken.ScopeReadOnly,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, "/test", nil)
			ctx := context.WithValue(req.Context(), ScopeKey, tc.scope)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			WriteScopeMiddleware(okHandler).ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status code = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}
