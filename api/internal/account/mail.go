package account

import (
	"context"
	"errors"
)

type Mailer interface {
	SendPasswordReset(context.Context, string, string) error
}

type unavailableMailer struct{}

func (unavailableMailer) SendPasswordReset(context.Context, string, string) error {
	return errors.New("password reset mailer is not configured")
}
