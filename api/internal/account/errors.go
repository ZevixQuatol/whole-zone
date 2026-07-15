package account

import "errors"

var (
	ErrAuth           = errors.New("invalid credentials")
	ErrDisabled       = errors.New("account disabled")
	ErrConflict       = errors.New("account conflict")
	ErrInviteInvalid  = errors.New("invalid invite")
	ErrNotFound       = errors.New("account not found")
	ErrSessionInvalid = errors.New("invalid session")
	ErrTokenInvalid   = errors.New("invalid token")
	ErrForbidden      = errors.New("forbidden")
)

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

func invalid(message string) error { return ValidationError{Message: message} }
