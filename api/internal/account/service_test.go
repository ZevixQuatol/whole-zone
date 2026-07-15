package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepo struct {
	registration *Registration
	users        map[string]User
	sessions     []Session
	registerErr  error
	authSession  *Session
	authUser     *User
	profile      *ProfileInput
	invite       *Invite
	reset        *PasswordReset
}

func (f *fakeRepo) Register(_ context.Context, registration Registration) error {
	if f.registerErr != nil {
		return f.registerErr
	}
	f.registration = &registration
	return nil
}

func (f *fakeRepo) UserByLogin(_ context.Context, login string) (User, error) {
	user, ok := f.users[login]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (f *fakeRepo) UserByID(context.Context, uuid.UUID) (User, error) {
	return User{}, ErrNotFound
}

func (f *fakeRepo) UserByHandle(_ context.Context, handle string) (User, error) {
	return f.UserByLogin(context.Background(), handle)
}

func (f *fakeRepo) CreateSession(_ context.Context, session Session) error {
	f.sessions = append(f.sessions, session)
	return nil
}

func (f *fakeRepo) SessionByHash(context.Context, []byte, time.Time) (Session, User, error) {
	if f.authSession == nil || f.authUser == nil {
		return Session{}, User{}, ErrSessionInvalid
	}
	return *f.authSession, *f.authUser, nil
}

func (f *fakeRepo) RevokeSession(context.Context, uuid.UUID, time.Time) error     { return nil }
func (f *fakeRepo) RevokeAllSessions(context.Context, uuid.UUID, time.Time) error { return nil }
func (f *fakeRepo) UpdateProfile(_ context.Context, _ uuid.UUID, displayName, bio, location string, _ time.Time) (User, error) {
	f.profile = &ProfileInput{DisplayName: displayName, Bio: bio, Location: location}
	user := User{DisplayName: displayName, Bio: bio, Location: location}
	if f.authUser != nil {
		user = *f.authUser
		user.DisplayName = displayName
		user.Bio = bio
		user.Location = location
	}
	return user, nil
}
func (f *fakeRepo) UpdatePassword(context.Context, uuid.UUID, string, time.Time) error { return nil }
func (f *fakeRepo) CreatePasswordReset(_ context.Context, reset PasswordReset) error {
	f.reset = &reset
	return nil
}
func (f *fakeRepo) UsePasswordReset(context.Context, []byte, string, time.Time) error { return nil }
func (f *fakeRepo) ListNotifications(context.Context, uuid.UUID, int) ([]Notification, error) {
	return nil, nil
}
func (f *fakeRepo) ReadNotification(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}
func (f *fakeRepo) ListInvites(context.Context, int) ([]Invite, error) { return nil, nil }
func (f *fakeRepo) CreateInvite(_ context.Context, invite Invite) error {
	f.invite = &invite
	return nil
}
func (f *fakeRepo) DisableInvite(context.Context, uuid.UUID, time.Time) error { return nil }

func TestRegisterCreatesMemberAndSession(t *testing.T) {
	repo := &fakeRepo{}
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	service := NewService(repo, 30*24*time.Hour, func() time.Time { return now })
	invite, _ := NewToken()

	result, err := service.Register(context.Background(), RegisterInput{
		InviteCode:  invite.Raw,
		Email:       " DEV@example.com ",
		Handle:      " Dev_User ",
		Password:    "long-enough-password",
		DisplayName: " 开发者 ",
	}, "test-agent")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if repo.registration == nil {
		t.Fatal("registration was not persisted")
	}
	if result.User.Email != "dev@example.com" || result.User.Handle != "dev_user" {
		t.Fatalf("user was not normalized: %+v", result.User)
	}
	if result.User.Role != RoleMember || result.User.Status != StatusActive {
		t.Fatalf("unexpected account defaults: %+v", result.User)
	}
	if result.ExpiresAt != now.Add(30*24*time.Hour) || result.SessionToken == "" || result.CSRFToken == "" {
		t.Fatalf("unexpected auth result: %+v", result)
	}
	if repo.registration.Session.UserID != result.User.ID {
		t.Fatal("session was not linked to the user")
	}
}

func TestRegisterRejectsInvalidInputBeforeRepository(t *testing.T) {
	service := NewService(&fakeRepo{}, time.Hour, time.Now)
	validInvite, _ := NewToken()
	tests := []RegisterInput{
		{InviteCode: validInvite.Raw, Email: "bad", Handle: "valid_name", Password: "long-enough-password", DisplayName: "name"},
		{InviteCode: validInvite.Raw, Email: "a@example.com", Handle: "?", Password: "long-enough-password", DisplayName: "name"},
		{InviteCode: validInvite.Raw, Email: "a@example.com", Handle: "valid_name", Password: "short", DisplayName: "name"},
		{InviteCode: validInvite.Raw, Email: "a@example.com", Handle: "valid_name", Password: "long-enough-password", DisplayName: ""},
		{InviteCode: "bad", Email: "a@example.com", Handle: "valid_name", Password: "long-enough-password", DisplayName: "name"},
	}
	for i, input := range tests {
		if _, err := service.Register(context.Background(), input, ""); err == nil {
			t.Fatalf("case %d: expected invalid input to fail", i)
		}
	}
}

func TestLoginUsesSamePublicErrorForUnknownAndWrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct-password")
	user := User{ID: uuid.New(), Email: "dev@example.com", Handle: "dev", PasswordHash: hash, Status: StatusActive}
	repo := &fakeRepo{users: map[string]User{"dev@example.com": user, "dev": user}}
	service := NewService(repo, time.Hour, time.Now)

	for _, input := range []LoginInput{
		{Login: "missing", Password: "anything"},
		{Login: "dev", Password: "wrong-password"},
	} {
		if _, err := service.Login(context.Background(), input, ""); !errors.Is(err, ErrAuth) {
			t.Fatalf("login error = %v, want ErrAuth", err)
		}
	}
}

func TestLoginCreatesSessionForActiveUser(t *testing.T) {
	hash, _ := HashPassword("correct-password")
	user := User{ID: uuid.New(), Email: "dev@example.com", Handle: "dev", PasswordHash: hash, Status: StatusActive}
	repo := &fakeRepo{users: map[string]User{"dev@example.com": user}}
	service := NewService(repo, time.Hour, time.Now)

	result, err := service.Login(context.Background(), LoginInput{Login: " DEV@example.com ", Password: "correct-password"}, "agent")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if len(repo.sessions) != 1 || result.SessionToken == "" || result.CSRFToken == "" {
		t.Fatalf("session was not created: result=%+v sessions=%d", result, len(repo.sessions))
	}
}

type fakeMailer struct {
	email string
	token string
	err   error
}

func (m *fakeMailer) SendPasswordReset(_ context.Context, email, token string) error {
	m.email = email
	m.token = token
	return m.err
}

func TestForgotPasswordCreatesExpiringTokenAndSendsMail(t *testing.T) {
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	user := User{ID: uuid.New(), Email: "dev@example.com", Status: StatusActive}
	repo := &fakeRepo{users: map[string]User{"dev@example.com": user}}
	mailer := &fakeMailer{}
	service := NewService(repo, time.Hour, func() time.Time { return now })
	service.SetMailer(mailer, 20*time.Minute)

	if err := service.ForgotPassword(context.Background(), " DEV@example.com "); err != nil {
		t.Fatalf("forgot password: %v", err)
	}
	if repo.reset == nil || repo.reset.UserID != user.ID || repo.reset.ExpiresAt != now.Add(20*time.Minute) {
		t.Fatalf("unexpected password reset: %+v", repo.reset)
	}
	if mailer.email != user.Email || mailer.token == "" {
		t.Fatalf("reset mail was not sent: %+v", mailer)
	}
	if string(repo.reset.TokenHash) == mailer.token {
		t.Fatal("raw reset token was stored")
	}
}

func TestForgotPasswordDoesNotRevealUnknownAccount(t *testing.T) {
	mailer := &fakeMailer{}
	service := NewService(&fakeRepo{users: map[string]User{}}, time.Hour, time.Now)
	service.SetMailer(mailer, 20*time.Minute)

	if err := service.ForgotPassword(context.Background(), "missing@example.com"); err != nil {
		t.Fatalf("unknown account returned an error: %v", err)
	}
	if mailer.email != "" || mailer.token != "" {
		t.Fatal("reset mail was sent for an unknown account")
	}
}
