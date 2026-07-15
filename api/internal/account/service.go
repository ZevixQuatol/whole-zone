package account

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var handlePattern = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

type Service struct {
	repo       Repo
	now        func() time.Time
	sessionTTL time.Duration
	resetTTL   time.Duration
	mailer     Mailer
}

type RegisterInput struct {
	InviteCode  string
	Email       string
	Handle      string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Login    string
	Password string
}

type AuthResult struct {
	User         User
	SessionToken string
	CSRFToken    string
	ExpiresAt    time.Time
}

func NewService(repo Repo, sessionTTL time.Duration, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{
		repo:       repo,
		now:        now,
		sessionTTL: sessionTTL,
		resetTTL:   30 * time.Minute,
		mailer:     unavailableMailer{},
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput, userAgent string) (AuthResult, error) {
	email, handle, displayName, err := NormalizeRegistration(input)
	if err != nil {
		return AuthResult{}, err
	}
	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}
	sessionToken, csrfToken, session, err := s.newSession(uuid.Nil, userAgent)
	if err != nil {
		return AuthResult{}, err
	}

	now := s.now().UTC()
	user := User{
		ID:           uuid.New(),
		Email:        email,
		Handle:       handle,
		PasswordHash: passwordHash,
		Role:         RoleMember,
		Status:       StatusActive,
		DisplayName:  displayName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	session.UserID = user.ID

	inviteHash, err := ParseToken(strings.TrimSpace(input.InviteCode))
	if err != nil {
		return AuthResult{}, ErrInviteInvalid
	}
	if err := s.repo.Register(ctx, Registration{User: user, InviteHash: inviteHash, Session: session}); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: user, SessionToken: sessionToken, CSRFToken: csrfToken, ExpiresAt: session.ExpiresAt}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput, userAgent string) (AuthResult, error) {
	login := strings.ToLower(strings.TrimSpace(input.Login))
	user, err := s.repo.UserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResult{}, ErrAuth
		}
		return AuthResult{}, err
	}
	if user.Status != StatusActive {
		return AuthResult{}, ErrDisabled
	}
	ok, err := CheckPassword(input.Password, user.PasswordHash)
	if err != nil || !ok {
		return AuthResult{}, ErrAuth
	}

	sessionToken, csrfToken, session, err := s.newSession(user.ID, userAgent)
	if err != nil {
		return AuthResult{}, err
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: user, SessionToken: sessionToken, CSRFToken: csrfToken, ExpiresAt: session.ExpiresAt}, nil
}

func (s *Service) newSession(userID uuid.UUID, userAgent string) (string, string, Session, error) {
	sessionToken, err := NewToken()
	if err != nil {
		return "", "", Session{}, err
	}
	csrfToken, err := NewToken()
	if err != nil {
		return "", "", Session{}, err
	}
	now := s.now().UTC()
	return sessionToken.Raw, csrfToken.Raw, Session{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  sessionToken.Hash,
		CSRFHash:   csrfToken.Hash,
		UserAgent:  userAgent,
		ExpiresAt:  now.Add(s.sessionTTL),
		LastSeenAt: now,
		CreatedAt:  now,
	}, nil
}

// NormalizeRegistration validates the shared account identity rules used by registration and admin bootstrap.
func NormalizeRegistration(input RegisterInput) (string, string, string, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 254 {
		return "", "", "", invalid("邮箱格式无效")
	}
	handle := strings.ToLower(strings.TrimSpace(input.Handle))
	if !handlePattern.MatchString(handle) {
		return "", "", "", invalid("用户名只能包含字母、数字和下划线，长度为 3 至 32")
	}
	if len(input.Password) < 12 || len(input.Password) > 128 {
		return "", "", "", invalid("密码长度必须为 12 至 128 个字符")
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" || len([]rune(displayName)) > 64 {
		return "", "", "", invalid("显示名称不能为空且最多 64 个字符")
	}
	return email, handle, displayName, nil
}
