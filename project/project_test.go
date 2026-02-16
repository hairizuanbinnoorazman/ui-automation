package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		wantErr error
	}{
		{
			name: "valid project",
			project: Project{
				Name:    "Test Project",
				OwnerID: 1,
			},
			wantErr: nil,
		},
		{
			name: "valid project with description",
			project: Project{
				Name:        "Test Project",
				Description: "A test project description",
				OwnerID:     1,
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			project: Project{
				OwnerID: 1,
			},
			wantErr: ErrInvalidProjectName,
		},
		{
			name: "missing owner_id",
			project: Project{
				Name: "Test Project",
			},
			wantErr: ErrInvalidOwner,
		},
		{
			name:    "missing both name and owner_id",
			project: Project{},
			wantErr: ErrInvalidProjectName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
