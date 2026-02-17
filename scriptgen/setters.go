package scriptgen

// SetStatus returns a setter that updates the generation status.
func SetStatus(status GenerationStatus) UpdateSetter {
	return func(gs *GeneratedScript) error {
		gs.GenerationStatus = status
		return nil
	}
}

// SetErrorMessage returns a setter that updates the error message.
func SetErrorMessage(message string) UpdateSetter {
	return func(gs *GeneratedScript) error {
		gs.ErrorMessage = &message
		return nil
	}
}

// SetScriptPath returns a setter that updates the script path and file size.
func SetScriptPath(path string, size int64) UpdateSetter {
	return func(gs *GeneratedScript) error {
		gs.ScriptPath = path
		gs.FileSize = size
		return nil
	}
}
