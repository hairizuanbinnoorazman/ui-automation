package testprocedure

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProcedure_Validate(t *testing.T) {
	projectID := uuid.New()
	createdBy := uuid.New()
	tests := []struct {
		name          string
		testProcedure TestProcedure
		wantErr       error
	}{
		{
			name: "valid test procedure",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: projectID,
				CreatedBy: createdBy,
			},
			wantErr: nil,
		},
		{
			name: "valid test procedure with steps",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: projectID,
				CreatedBy: createdBy,
				Steps: Steps{
					{Name: "Click button", Instructions: "Click the submit button", ImagePaths: []string{}},
					{Name: "Type input", Instructions: "Type test into input", ImagePaths: []string{"path/image.png"}},
				},
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			testProcedure: TestProcedure{
				ProjectID: projectID,
				CreatedBy: createdBy,
			},
			wantErr: ErrInvalidTestProcedureName,
		},
		{
			name: "missing project_id",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				CreatedBy: createdBy,
			},
			wantErr: ErrInvalidProjectID,
		},
		{
			name: "missing created_by",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: projectID,
			},
			wantErr: ErrInvalidCreatedBy,
		},
		{
			name: "step without name",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: projectID,
				CreatedBy: createdBy,
				Steps: Steps{
					{Name: "", Instructions: "Missing name", ImagePaths: []string{}},
				},
			},
			wantErr: ErrInvalidStepName,
		},
		{
			name:          "missing all required fields",
			testProcedure: TestProcedure{},
			wantErr:       ErrInvalidTestProcedureName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testProcedure.Validate()
			if tt.wantErr != nil {
				if tt.wantErr == ErrInvalidStepName {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "step")
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSteps_MarshalUnmarshal(t *testing.T) {
	steps := Steps{
		{Name: "Navigate", Instructions: "Go to example.com", ImagePaths: []string{"nav.png"}},
		{Name: "Click", Instructions: "Click the button", ImagePaths: []string{}},
		{Name: "Type", Instructions: "Type test into input", ImagePaths: []string{"before.png", "after.png"}},
	}

	// Marshal to JSON
	data, err := json.Marshal(steps)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled Steps
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, len(steps), len(unmarshaled))
	assert.Equal(t, steps[0].Name, unmarshaled[0].Name)
	assert.Equal(t, steps[0].Instructions, unmarshaled[0].Instructions)
	assert.Equal(t, steps[0].ImagePaths, unmarshaled[0].ImagePaths)
	assert.Equal(t, steps[1].Name, unmarshaled[1].Name)
	assert.Equal(t, steps[2].ImagePaths, unmarshaled[2].ImagePaths)
}

func TestSteps_Value(t *testing.T) {
	t.Run("non-nil steps", func(t *testing.T) {
		steps := Steps{
			{Name: "Click", Instructions: "Click button", ImagePaths: []string{}},
		}
		value, err := steps.Value()
		require.NoError(t, err)
		assert.NotNil(t, value)

		// Verify it's valid JSON
		var result Steps
		err = json.Unmarshal(value.([]byte), &result)
		require.NoError(t, err)
		assert.Equal(t, steps, result)
	})

	t.Run("empty steps", func(t *testing.T) {
		steps := Steps{}
		value, err := steps.Value()
		require.NoError(t, err)
		assert.NotNil(t, value)

		// Should marshal to empty array
		var result Steps
		err = json.Unmarshal(value.([]byte), &result)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("nil steps", func(t *testing.T) {
		var steps Steps
		value, err := steps.Value()
		require.NoError(t, err)
		// Should marshal to empty array, not nil
		assert.NotNil(t, value)
	})
}

func TestSteps_Scan(t *testing.T) {
	t.Run("scan valid JSON", func(t *testing.T) {
		jsonData := []byte(`[{"name":"Click","instructions":"Click button","image_paths":[]}]`)
		var steps Steps
		err := steps.Scan(jsonData)
		require.NoError(t, err)
		assert.Len(t, steps, 1)
		assert.Equal(t, "Click", steps[0].Name)
		assert.Equal(t, "Click button", steps[0].Instructions)
		assert.Empty(t, steps[0].ImagePaths)
	})

	t.Run("scan with image paths", func(t *testing.T) {
		jsonData := []byte(`[{"name":"Step","instructions":"Do it","image_paths":["a.png","b.png"]}]`)
		var steps Steps
		err := steps.Scan(jsonData)
		require.NoError(t, err)
		assert.Len(t, steps, 1)
		assert.Len(t, steps[0].ImagePaths, 2)
		assert.Equal(t, "a.png", steps[0].ImagePaths[0])
	})

	t.Run("scan nil value", func(t *testing.T) {
		var steps Steps
		err := steps.Scan(nil)
		require.NoError(t, err)
		assert.Empty(t, steps)
	})

	t.Run("scan empty array", func(t *testing.T) {
		var steps Steps
		err := steps.Scan([]byte(`[]`))
		require.NoError(t, err)
		assert.Empty(t, steps)
	})

	t.Run("scan invalid type", func(t *testing.T) {
		var steps Steps
		err := steps.Scan("invalid")
		assert.Error(t, err)
	})

	t.Run("scan invalid JSON", func(t *testing.T) {
		var steps Steps
		err := steps.Scan([]byte(`invalid json`))
		assert.Error(t, err)
	})
}
