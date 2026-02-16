package testprocedure

// SetName returns an UpdateSetter that sets the test procedure's name.
func SetName(name string) UpdateSetter {
	return func(tp *TestProcedure) error {
		if name == "" {
			return ErrInvalidTestProcedureName
		}
		tp.Name = name
		return nil
	}
}

// SetDescription returns an UpdateSetter that sets the test procedure's description.
func SetDescription(description string) UpdateSetter {
	return func(tp *TestProcedure) error {
		tp.Description = description
		return nil
	}
}

// SetSteps returns an UpdateSetter that sets the test procedure's steps.
func SetSteps(steps Steps) UpdateSetter {
	return func(tp *TestProcedure) error {
		tp.Steps = steps
		return nil
	}
}
