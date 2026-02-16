package testrun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"pending is valid", StatusPending, true},
		{"running is valid", StatusRunning, true},
		{"passed is valid", StatusPassed, true},
		{"failed is valid", StatusFailed, true},
		{"skipped is valid", StatusSkipped, true},
		{"invalid status", Status("invalid"), false},
		{"empty status", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"passed is final", StatusPassed, true},
		{"failed is final", StatusFailed, true},
		{"skipped is final", StatusSkipped, true},
		{"pending is not final", StatusPending, false},
		{"running is not final", StatusRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsFinal())
		})
	}
}

func TestTestRun_Validate(t *testing.T) {
	tests := []struct {
		name    string
		testRun TestRun
		wantErr error
	}{
		{
			name: "valid test run",
			testRun: TestRun{
				TestProcedureID: 1,
				ExecutedBy:      1,
				Status:          StatusPending,
			},
			wantErr: nil,
		},
		{
			name: "missing test_procedure_id",
			testRun: TestRun{
				ExecutedBy: 1,
				Status:     StatusPending,
			},
			wantErr: ErrInvalidTestProcedureID,
		},
		{
			name: "missing executed_by",
			testRun: TestRun{
				TestProcedureID: 1,
				Status:          StatusPending,
			},
			wantErr: ErrInvalidExecutedBy,
		},
		{
			name: "invalid status",
			testRun: TestRun{
				TestProcedureID: 1,
				ExecutedBy:      1,
				Status:          Status("invalid"),
			},
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testRun.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTestRun_Start(t *testing.T) {
	t.Run("successfully start test run", func(t *testing.T) {
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusPending,
		}

		err := tr.Start()
		assert.NoError(t, err)
		assert.NotNil(t, tr.StartedAt)
		assert.Equal(t, StatusRunning, tr.Status)
		assert.WithinDuration(t, time.Now(), *tr.StartedAt, time.Second)
	})

	t.Run("cannot start already started test run", func(t *testing.T) {
		now := time.Now()
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusRunning,
			StartedAt:       &now,
		}

		err := tr.Start()
		assert.ErrorIs(t, err, ErrTestRunAlreadyStarted)
		assert.Equal(t, now, *tr.StartedAt) // StartedAt should not change
	})
}

func TestTestRun_Complete(t *testing.T) {
	t.Run("successfully complete test run with passed", func(t *testing.T) {
		now := time.Now()
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusRunning,
			StartedAt:       &now,
		}

		err := tr.Complete(StatusPassed, "All tests passed")
		assert.NoError(t, err)
		assert.NotNil(t, tr.CompletedAt)
		assert.Equal(t, StatusPassed, tr.Status)
		assert.Equal(t, "All tests passed", tr.Notes)
		assert.WithinDuration(t, time.Now(), *tr.CompletedAt, time.Second)
	})

	t.Run("successfully complete test run with failed", func(t *testing.T) {
		now := time.Now()
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusRunning,
			StartedAt:       &now,
		}

		err := tr.Complete(StatusFailed, "Test failed at step 3")
		assert.NoError(t, err)
		assert.Equal(t, StatusFailed, tr.Status)
		assert.Equal(t, "Test failed at step 3", tr.Notes)
	})

	t.Run("cannot complete non-running test run", func(t *testing.T) {
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusPending,
		}

		err := tr.Complete(StatusPassed, "")
		assert.ErrorIs(t, err, ErrTestRunNotRunning)
	})

	t.Run("cannot complete with non-final status", func(t *testing.T) {
		now := time.Now()
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusRunning,
			StartedAt:       &now,
		}

		err := tr.Complete(StatusPending, "")
		assert.ErrorIs(t, err, ErrInvalidStatus)

		err = tr.Complete(StatusRunning, "")
		assert.ErrorIs(t, err, ErrInvalidStatus)
	})

	t.Run("complete without notes", func(t *testing.T) {
		now := time.Now()
		tr := &TestRun{
			TestProcedureID: 1,
			ExecutedBy:      1,
			Status:          StatusRunning,
			StartedAt:       &now,
		}

		err := tr.Complete(StatusSkipped, "")
		assert.NoError(t, err)
		assert.Equal(t, StatusSkipped, tr.Status)
		assert.Empty(t, tr.Notes)
	})
}
