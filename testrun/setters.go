package testrun

import "github.com/google/uuid"

// SetStatus returns an UpdateSetter that sets the test run's status.
func SetStatus(status Status) UpdateSetter {
	return func(tr *TestRun) error {
		if !status.IsValid() {
			return ErrInvalidStatus
		}
		tr.Status = status
		return nil
	}
}

// SetNotes returns an UpdateSetter that sets the test run's notes.
func SetNotes(notes string) UpdateSetter {
	return func(tr *TestRun) error {
		tr.Notes = notes
		return nil
	}
}

// SetAssignedTo returns an UpdateSetter that assigns a user to the test run.
func SetAssignedTo(userID uuid.UUID) UpdateSetter {
	return func(tr *TestRun) error {
		tr.AssignedTo = &userID
		return nil
	}
}

// ClearAssignedTo returns an UpdateSetter that unassigns the user from the test run.
func ClearAssignedTo() UpdateSetter {
	return func(tr *TestRun) error {
		tr.AssignedTo = nil
		return nil
	}
}
