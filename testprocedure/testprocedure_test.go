package testprocedure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestProcedure_Validate(t *testing.T) {
	tests := []struct {
		name          string
		testProcedure TestProcedure
		wantErr       error
	}{
		{
			name: "valid test procedure",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: 1,
				CreatedBy: 1,
			},
			wantErr: nil,
		},
		{
			name: "valid test procedure with steps",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: 1,
				CreatedBy: 1,
				Steps: Steps{
					{"action": "click", "selector": "#button"},
					{"action": "type", "selector": "#input", "value": "test"},
				},
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			testProcedure: TestProcedure{
				ProjectID: 1,
				CreatedBy: 1,
			},
			wantErr: ErrInvalidTestProcedureName,
		},
		{
			name: "missing project_id",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				CreatedBy: 1,
			},
			wantErr: ErrInvalidProjectID,
		},
		{
			name: "missing created_by",
			testProcedure: TestProcedure{
				Name:      "Test Procedure",
				ProjectID: 1,
			},
			wantErr: ErrInvalidCreatedBy,
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
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSteps_MarshalUnmarshal(t *testing.T) {
	steps := Steps{
		{"action": "navigate", "url": "https://example.com"},
		{"action": "click", "selector": "#button"},
		{"action": "type", "selector": "#input", "value": "test"},
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
	assert.Equal(t, steps[0]["action"], unmarshaled[0]["action"])
	assert.Equal(t, steps[1]["selector"], unmarshaled[1]["selector"])
	assert.Equal(t, steps[2]["value"], unmarshaled[2]["value"])
}

func TestSteps_Value(t *testing.T) {
	t.Run("non-nil steps", func(t *testing.T) {
		steps := Steps{
			{"action": "click"},
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

	t.Run("nil steps", func(t *testing.T) {
		var steps Steps
		value, err := steps.Value()
		require.NoError(t, err)
		assert.Nil(t, value)
	})
}

func TestSteps_Scan(t *testing.T) {
	t.Run("scan valid JSON", func(t *testing.T) {
		jsonData := []byte(`[{"action":"click","selector":"#button"}]`)
		var steps Steps
		err := steps.Scan(jsonData)
		require.NoError(t, err)
		assert.Len(t, steps, 1)
		assert.Equal(t, "click", steps[0]["action"])
	})

	t.Run("scan nil value", func(t *testing.T) {
		var steps Steps
		err := steps.Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, steps)
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
