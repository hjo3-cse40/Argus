package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when email/password do not match
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUserNotFound is returned when no user exists with the given identifier
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when registering with an email already in use
	ErrUserAlreadyExists = errors.New("email already in use")

	// ErrSessionExpired is returned when a session exists but has passed its expiry time
	ErrSessionExpired = errors.New("session has expired")

	// ErrSessionNotFound is returned when no session exists for the given session ID
	ErrSessionNotFound = errors.New("session not found")

	// ErrMissingSessionCookie is returned when no session cookie is present on the request
	ErrMissingSessionCookie = errors.New("missing session cookie")
)
