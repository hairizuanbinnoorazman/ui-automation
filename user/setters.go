package user

// SetEmail returns an UpdateSetter that sets the user's email.
func SetEmail(email string) UpdateSetter {
	return func(u *User) error {
		if email == "" {
			return ErrInvalidEmail
		}
		u.Email = email
		return nil
	}
}

// SetUsername returns an UpdateSetter that sets the user's username.
func SetUsername(username string) UpdateSetter {
	return func(u *User) error {
		if username == "" {
			return ErrInvalidUsername
		}
		u.Username = username
		return nil
	}
}

// SetPassword returns an UpdateSetter that sets the user's password.
func SetPassword(password string) UpdateSetter {
	return func(u *User) error {
		return u.SetPassword(password)
	}
}

// SetActive returns an UpdateSetter that sets the user's active status.
func SetActive(active bool) UpdateSetter {
	return func(u *User) error {
		u.IsActive = active
		return nil
	}
}
