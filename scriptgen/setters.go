package scriptgen

// SetStatus returns a setter that updates the generation status.
func SetStatus(status GenerationStatus) UpdateSetter {
	return func() map[string]interface{} {
		return map[string]interface{}{"generation_status": status}
	}
}

// SetErrorMessage returns a setter that updates the error message.
func SetErrorMessage(message string) UpdateSetter {
	return func() map[string]interface{} {
		return map[string]interface{}{"error_message": message}
	}
}

// SetScriptPath returns a setter that updates the script path and file size.
func SetScriptPath(path string, size int64) UpdateSetter {
	return func() map[string]interface{} {
		return map[string]interface{}{
			"script_path": path,
			"file_size":   size,
		}
	}
}
