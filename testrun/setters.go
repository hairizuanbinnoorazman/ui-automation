package testrun

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
