package endpoint

// SetName returns an UpdateSetter that sets the endpoint's name.
func SetName(name string) UpdateSetter {
	return func(e *Endpoint) error {
		if name == "" {
			return ErrInvalidEndpointName
		}
		e.Name = name
		return nil
	}
}

// SetURL returns an UpdateSetter that sets the endpoint's URL.
func SetURL(url string) UpdateSetter {
	return func(e *Endpoint) error {
		if url == "" {
			return ErrInvalidEndpointURL
		}
		e.URL = url
		return nil
	}
}

// SetCredentials returns an UpdateSetter that sets the endpoint's credentials.
func SetCredentials(creds Credentials) UpdateSetter {
	return func(e *Endpoint) error {
		e.Credentials = creds
		return nil
	}
}
