package project

// SetName returns an UpdateSetter that sets the project's name.
func SetName(name string) UpdateSetter {
	return func(p *Project) error {
		if name == "" {
			return ErrInvalidProjectName
		}
		p.Name = name
		return nil
	}
}

// SetDescription returns an UpdateSetter that sets the project's description.
func SetDescription(description string) UpdateSetter {
	return func(p *Project) error {
		p.Description = description
		return nil
	}
}

// SetActive returns an UpdateSetter that sets the project's active status.
func SetActive(active bool) UpdateSetter {
	return func(p *Project) error {
		p.IsActive = active
		return nil
	}
}
