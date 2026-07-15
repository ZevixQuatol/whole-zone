package account

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	User       User
	InviteHash []byte
	Session    Session
}

type Repo interface {
	Register(context.Context, Registration) error
	UserByLogin(context.Context, string) (User, error)
	UserByID(context.Context, uuid.UUID) (User, error)
	UserByHandle(context.Context, string) (User, error)
	CreateSession(context.Context, Session) error
	SessionByHash(context.Context, []byte, time.Time) (Session, User, error)
	RevokeSession(context.Context, uuid.UUID, time.Time) error
	RevokeAllSessions(context.Context, uuid.UUID, time.Time) error
	UpdateProfile(context.Context, uuid.UUID, string, string, string, time.Time) (User, error)
	UpdatePassword(context.Context, uuid.UUID, string, time.Time) error
	CreatePasswordReset(context.Context, PasswordReset) error
	UsePasswordReset(context.Context, []byte, string, time.Time) error
	ListNotifications(context.Context, uuid.UUID, int) ([]Notification, error)
	ReadNotification(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	ListInvites(context.Context, int) ([]Invite, error)
	CreateInvite(context.Context, Invite) error
	DisableInvite(context.Context, uuid.UUID, time.Time) error
}
