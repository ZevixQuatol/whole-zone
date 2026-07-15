package account

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "platform_admin"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusDeleted  Status = "deleted"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Handle       string
	PasswordHash string
	Role         Role
	Status       Status
	DisplayName  string
	Bio          string
	Location     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  []byte
	CSRFHash   []byte
	UserAgent  string
	ExpiresAt  time.Time
	LastSeenAt time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Kind      string
	Title     string
	Body      string
	Data      map[string]any
	ReadAt    *time.Time
	CreatedAt time.Time
}

type Invite struct {
	ID         uuid.UUID
	CodeHash   []byte
	CreatedBy  *uuid.UUID
	MaxUses    int
	Uses       int
	ExpiresAt  time.Time
	DisabledAt *time.Time
	CreatedAt  time.Time
}

type PasswordReset struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash []byte
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type PublicUser struct {
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		Handle:      u.Handle,
		DisplayName: u.DisplayName,
		Bio:         u.Bio,
		Location:    u.Location,
		CreatedAt:   u.CreatedAt,
	}
}
