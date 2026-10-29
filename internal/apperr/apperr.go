// Package apperr contains the domain errors of the service together with a
// translation layer that decides which messages may be exposed to API clients.
package apperr

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidCredentials is returned when an email/password pair does not match.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when a user with the same email already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = errors.New("session not found")
	// ErrUnauthorized is returned when the request carries no usable session.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidAuthType is returned for unsupported authentication methods.
	ErrInvalidAuthType = errors.New("invalid auth type")
	// ErrInvalidToken is returned when an access token cannot be validated.
	ErrInvalidToken = errors.New("invalid token")
)

// InternalMessage is returned to clients for every error that is not
// explicitly marked as safe to expose.
const InternalMessage = "internal error"

// ValidationError marks an error caused by malformed client input.
type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return "validation failed: " + e.Reason
}

// NewValidationError wraps a validation failure.
func NewValidationError(reason string) error {
	return &ValidationError{Reason: reason}
}

// IsValidation reports whether err describes invalid client input.
func IsValidation(err error) bool {
	var v *ValidationError
	return errors.As(err, &v)
}

// PublicMessage maps err to a message that is safe to return to API clients.
// Unknown errors are reported as InternalMessage so implementation details
// never leak over the wire.
func PublicMessage(err error) string {
	if err == nil {
		return ""
	}

	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return "invalid request: " + validationErr.Reason
	}

	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return "invalid email or password"
	case errors.Is(err, ErrUserNotFound):
		return "user not found"
	case errors.Is(err, ErrUserAlreadyExists):
		return "user already exists"
	case errors.Is(err, ErrSessionNotFound):
		return "session not found"
	case errors.Is(err, ErrUnauthorized):
		return "unauthorized"
	case errors.Is(err, ErrInvalidAuthType):
		return "unsupported authentication method"
	case errors.Is(err, ErrInvalidToken):
		return "invalid token"
	}

	return InternalMessage
}

// Wrap annotates err with the call that produced it, keeping the original
// error reachable through errors.Is/errors.As.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", op, err)
}
