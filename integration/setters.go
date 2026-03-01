package integration

// SetName returns an IntegrationSetter that sets the integration's name.
func SetName(name string) IntegrationSetter {
	return func(i *Integration) error {
		if name == "" {
			return ErrInvalidName
		}
		i.Name = name
		return nil
	}
}

// SetIsActive returns an IntegrationSetter that sets the integration's active status.
func SetIsActive(isActive bool) IntegrationSetter {
	return func(i *Integration) error {
		i.IsActive = isActive
		return nil
	}
}

// SetEncryptedCredentials returns an IntegrationSetter that sets the encrypted credentials.
func SetEncryptedCredentials(creds []byte) IntegrationSetter {
	return func(i *Integration) error {
		i.EncryptedCredentials = creds
		return nil
	}
}

// SetTitle returns an IssueLinkSetter that sets the issue link's title.
func SetTitle(title string) IssueLinkSetter {
	return func(il *IssueLink) error {
		il.Title = title
		return nil
	}
}

// SetStatus returns an IssueLinkSetter that sets the issue link's status.
func SetStatus(status string) IssueLinkSetter {
	return func(il *IssueLink) error {
		il.Status = status
		return nil
	}
}

// SetURL returns an IssueLinkSetter that sets the issue link's URL.
func SetURL(url string) IssueLinkSetter {
	return func(il *IssueLink) error {
		il.URL = url
		return nil
	}
}
