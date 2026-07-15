package account

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Principal struct {
	User    User
	Session Session
}

type ProfileInput struct {
	DisplayName string
	Bio         string
	Location    string
}

type InviteResult struct {
	Invite Invite
	Code   string
}

func (s *Service) SetMailer(mailer Mailer, resetTTL time.Duration) {
	if mailer == nil {
		mailer = unavailableMailer{}
	}
	s.mailer = mailer
	s.resetTTL = resetTTL
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (Principal, error) {
	hash, err := ParseToken(rawToken)
	if err != nil {
		return Principal{}, ErrSessionInvalid
	}
	session, user, err := s.repo.SessionByHash(ctx, hash, s.now().UTC())
	if err != nil {
		return Principal{}, err
	}
	if user.Status != StatusActive {
		return Principal{}, ErrDisabled
	}
	return Principal{User: user, Session: session}, nil
}

func (s *Service) CheckCSRF(session Session, rawToken string) error {
	hash, err := ParseToken(rawToken)
	if err != nil || !bytes.Equal(hash, session.CSRFHash) {
		return ErrForbidden
	}
	return nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.RevokeSession(ctx, sessionID, s.now().UTC())
}

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllSessions(ctx, userID, s.now().UTC())
}

func (s *Service) PublicProfile(ctx context.Context, handle string) (PublicUser, error) {
	user, err := s.repo.UserByHandle(ctx, strings.ToLower(strings.TrimSpace(handle)))
	if err != nil {
		return PublicUser{}, err
	}
	return user.Public(), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, input ProfileInput) (User, error) {
	displayName := strings.TrimSpace(input.DisplayName)
	bio := strings.TrimSpace(input.Bio)
	location := strings.TrimSpace(input.Location)
	if displayName == "" || len([]rune(displayName)) > 64 || len([]rune(bio)) > 500 || len([]rune(location)) > 80 {
		return User{}, invalid("公开资料字段超出允许范围")
	}
	return s.repo.UpdateProfile(ctx, userID, displayName, bio, location, s.now().UTC())
}

func (s *Service) ChangePassword(ctx context.Context, user User, current, next string) error {
	ok, err := CheckPassword(current, user.PasswordHash)
	if err != nil || !ok {
		return ErrAuth
	}
	if len(next) < 12 || len(next) > 128 {
		return invalid("密码长度必须为 12 至 128 个字符")
	}
	hash, err := HashPassword(next)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, user.ID, hash, s.now().UTC())
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.UserByLogin(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, ErrNotFound) || (err == nil && user.Status != StatusActive) {
		return nil
	}
	if err != nil {
		return err
	}
	token, err := NewToken()
	if err != nil {
		return err
	}
	now := s.now().UTC()
	reset := PasswordReset{ID: uuid.New(), UserID: user.ID, TokenHash: token.Hash, CreatedAt: now, ExpiresAt: now.Add(s.resetTTL)}
	if err := s.repo.CreatePasswordReset(ctx, reset); err != nil {
		return err
	}
	return s.mailer.SendPasswordReset(ctx, user.Email, token.Raw)
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, next string) error {
	if len(next) < 12 || len(next) > 128 {
		return invalid("密码长度必须为 12 至 128 个字符")
	}
	tokenHash, err := ParseToken(rawToken)
	if err != nil {
		return ErrTokenInvalid
	}
	passwordHash, err := HashPassword(next)
	if err != nil {
		return err
	}
	return s.repo.UsePasswordReset(ctx, tokenHash, passwordHash, s.now().UTC())
}

func (s *Service) Notifications(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListNotifications(ctx, userID, limit)
}

func (s *Service) ReadNotification(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.ReadNotification(ctx, userID, id, s.now().UTC())
}

func (s *Service) Invites(ctx context.Context, actor User, limit int) ([]Invite, error) {
	if actor.Role != RoleAdmin {
		return nil, ErrForbidden
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.ListInvites(ctx, limit)
}

func (s *Service) CreateInvite(ctx context.Context, actor User, maxUses int, ttl time.Duration) (InviteResult, error) {
	if actor.Role != RoleAdmin {
		return InviteResult{}, ErrForbidden
	}
	if maxUses <= 0 || maxUses > 1000 || ttl <= 0 || ttl > 365*24*time.Hour {
		return InviteResult{}, invalid("邀请码次数或有效期超出允许范围")
	}
	token, err := NewToken()
	if err != nil {
		return InviteResult{}, err
	}
	now := s.now().UTC()
	actorID := actor.ID
	invite := Invite{ID: uuid.New(), CodeHash: token.Hash, CreatedBy: &actorID, MaxUses: maxUses, ExpiresAt: now.Add(ttl), CreatedAt: now}
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return InviteResult{}, err
	}
	return InviteResult{Invite: invite, Code: token.Raw}, nil
}

func (s *Service) DisableInvite(ctx context.Context, actor User, id uuid.UUID) error {
	if actor.Role != RoleAdmin {
		return ErrForbidden
	}
	return s.repo.DisableInvite(ctx, id, s.now().UTC())
}
