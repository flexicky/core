package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestPublicMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"credentials", ErrInvalidCredentials, "invalid email or password"},
		{"wrapped credentials", fmt.Errorf("emailLogin: %w", ErrInvalidCredentials), "invalid email or password"},
		{"unauthorized", ErrUnauthorized, "unauthorized"},
		{"user exists", ErrUserAlreadyExists, "user already exists"},
		{"validation", NewValidationError("Email is required"), "invalid request: Email is required"},
		{"unknown is hidden", errors.New("pq: password authentication failed for user \"postgres\""), InternalMessage},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := PublicMessage(test.err); got != test.want {
				t.Fatalf("PublicMessage() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWrapKeepsCause(t *testing.T) {
	t.Parallel()

	err := Wrap("create session", ErrSessionNotFound)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("Wrap() = %v, want wrapped ErrSessionNotFound", err)
	}

	if Wrap("noop", nil) != nil {
		t.Fatal("Wrap(nil) must return nil")
	}
}
