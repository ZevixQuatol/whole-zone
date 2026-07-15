package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PG struct {
	pool *pgxpool.Pool
}

func NewPG(pool *pgxpool.Pool) *PG {
	return &PG{pool: pool}
}

func (p *PG) Register(ctx context.Context, registration Registration) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var inviteID uuid.UUID
	err = tx.QueryRow(ctx, `
        UPDATE invites
        SET uses = uses + 1
        WHERE code_hash = $1
          AND disabled_at IS NULL
          AND expires_at > $2
          AND uses < max_uses
        RETURNING id`, registration.InviteHash, registration.User.CreatedAt).Scan(&inviteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInviteInvalid
	}
	if err != nil {
		return fmt.Errorf("consume invite: %w", err)
	}

	u := registration.User
	_, err = tx.Exec(ctx, `
        INSERT INTO users (
            id, email, handle, password_hash, role, status,
            display_name, bio, location, created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		u.ID, u.Email, u.Handle, u.PasswordHash, u.Role, u.Status,
		u.DisplayName, u.Bio, u.Location, u.CreatedAt, u.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	if err := insertSession(ctx, tx, registration.Session); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
        INSERT INTO account_events (id, user_id, kind, data, created_at)
        VALUES ($1, $2, 'registered', jsonb_build_object('inviteId', $3::text), $4)`,
		uuid.New(), u.ID, inviteID, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert account event: %w", err)
	}
	return tx.Commit(ctx)
}

func (p *PG) UserByLogin(ctx context.Context, login string) (User, error) {
	return queryUser(ctx, p.pool, `
        SELECT id,email,handle,password_hash,role,status,display_name,bio,location,created_at,updated_at
        FROM users
        WHERE deleted_at IS NULL AND (lower(email) = lower($1) OR lower(handle) = lower($1))`, login)
}

func (p *PG) UserByID(ctx context.Context, id uuid.UUID) (User, error) {
	return queryUser(ctx, p.pool, `
        SELECT id,email,handle,password_hash,role,status,display_name,bio,location,created_at,updated_at
        FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
}

func (p *PG) UserByHandle(ctx context.Context, handle string) (User, error) {
	return queryUser(ctx, p.pool, `
        SELECT id,email,handle,password_hash,role,status,display_name,bio,location,created_at,updated_at
        FROM users WHERE lower(handle) = lower($1) AND deleted_at IS NULL`, handle)
}

func (p *PG) CreateSession(ctx context.Context, session Session) error {
	return insertSession(ctx, p.pool, session)
}

func (p *PG) SessionByHash(ctx context.Context, hash []byte, now time.Time) (Session, User, error) {
	var session Session
	var user User
	err := p.pool.QueryRow(ctx, `
        SELECT s.id,s.user_id,s.token_hash,s.csrf_hash,s.user_agent,s.expires_at,
               s.last_seen_at,s.revoked_at,s.created_at,
               u.id,u.email,u.handle,u.password_hash,u.role,u.status,u.display_name,
               u.bio,u.location,u.created_at,u.updated_at
        FROM sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > $2
          AND u.deleted_at IS NULL`, hash, now).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.CSRFHash,
		&session.UserAgent, &session.ExpiresAt, &session.LastSeenAt, &session.RevokedAt, &session.CreatedAt,
		&user.ID, &user.Email, &user.Handle, &user.PasswordHash, &user.Role, &user.Status,
		&user.DisplayName, &user.Bio, &user.Location, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, User{}, ErrSessionInvalid
	}
	return session, user, err
}

func (p *PG) RevokeSession(ctx context.Context, id uuid.UUID, now time.Time) error {
	_, err := p.pool.Exec(ctx, `UPDATE sessions SET revoked_at = COALESCE(revoked_at, $2) WHERE id = $1`, id, now)
	return err
}

func (p *PG) RevokeAllSessions(ctx context.Context, userID uuid.UUID, now time.Time) error {
	_, err := p.pool.Exec(ctx, `UPDATE sessions SET revoked_at = COALESCE(revoked_at, $2) WHERE user_id = $1`, userID, now)
	return err
}

func (p *PG) UpdateProfile(ctx context.Context, id uuid.UUID, displayName, bio, location string, now time.Time) (User, error) {
	_, err := p.pool.Exec(ctx, `
        UPDATE users SET display_name=$2,bio=$3,location=$4,updated_at=$5
        WHERE id=$1 AND deleted_at IS NULL`, id, displayName, bio, location, now)
	if err != nil {
		return User{}, err
	}
	return p.UserByID(ctx, id)
}

func (p *PG) UpdatePassword(ctx context.Context, id uuid.UUID, hash string, now time.Time) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `UPDATE users SET password_hash=$2,updated_at=$3 WHERE id=$1`, id, hash, now); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE user_id=$1`, id, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *PG) CreatePasswordReset(ctx context.Context, reset PasswordReset) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO password_resets (id,user_id,token_hash,expires_at,created_at)
        VALUES ($1,$2,$3,$4,$5)`, reset.ID, reset.UserID, reset.TokenHash, reset.ExpiresAt, reset.CreatedAt)
	return err
}

func (p *PG) UsePasswordReset(ctx context.Context, tokenHash []byte, passwordHash string, now time.Time) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
        UPDATE password_resets SET used_at=$2
        WHERE token_hash=$1 AND used_at IS NULL AND expires_at>$2
        RETURNING user_id`, tokenHash, now).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTokenInvalid
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET password_hash=$2,updated_at=$3 WHERE id=$1`, userID, passwordHash, now); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE user_id=$1`, userID, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *PG) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id,user_id,kind,title,body,data,read_at,created_at
        FROM system_notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifications []Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Kind, &notification.Title,
			&notification.Body, &notification.Data, &notification.ReadAt, &notification.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (p *PG) ReadNotification(ctx context.Context, userID, id uuid.UUID, now time.Time) error {
	tag, err := p.pool.Exec(ctx, `
        UPDATE system_notifications SET read_at=COALESCE(read_at,$3) WHERE id=$1 AND user_id=$2`, id, userID, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PG) ListInvites(ctx context.Context, limit int) ([]Invite, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id,code_hash,created_by,max_uses,uses,expires_at,disabled_at,created_at
        FROM invites ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invites []Invite
	for rows.Next() {
		var invite Invite
		if err := rows.Scan(&invite.ID, &invite.CodeHash, &invite.CreatedBy, &invite.MaxUses,
			&invite.Uses, &invite.ExpiresAt, &invite.DisabledAt, &invite.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (p *PG) CreateInvite(ctx context.Context, invite Invite) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO invites (id,code_hash,created_by,max_uses,uses,expires_at,created_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7)`, invite.ID, invite.CodeHash, invite.CreatedBy,
		invite.MaxUses, invite.Uses, invite.ExpiresAt, invite.CreatedAt)
	return err
}

func (p *PG) DisableInvite(ctx context.Context, id uuid.UUID, now time.Time) error {
	tag, err := p.pool.Exec(ctx, `UPDATE invites SET disabled_at=COALESCE(disabled_at,$2) WHERE id=$1`, id, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PG) BootstrapAdmin(ctx context.Context, user User) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO users (
            id,email,handle,password_hash,role,status,display_name,bio,location,created_at,updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		user.ID, user.Email, user.Handle, user.PasswordHash, user.Role, user.Status,
		user.DisplayName, user.Bio, user.Location, user.CreatedAt, user.UpdatedAt)
	if isUniqueViolation(err) {
		existing, findErr := p.UserByLogin(ctx, user.Email)
		if findErr == nil && existing.Role == RoleAdmin {
			return nil
		}
		return ErrConflict
	}
	return err
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type execQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func queryUser(ctx context.Context, db rowQuerier, query string, arg any) (User, error) {
	var user User
	err := db.QueryRow(ctx, query, arg).Scan(
		&user.ID, &user.Email, &user.Handle, &user.PasswordHash, &user.Role, &user.Status,
		&user.DisplayName, &user.Bio, &user.Location, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func insertSession(ctx context.Context, db execQuerier, session Session) error {
	_, err := db.Exec(ctx, `
        INSERT INTO sessions (
            id,user_id,token_hash,csrf_hash,user_agent,expires_at,last_seen_at,created_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		session.ID, session.UserID, session.TokenHash, session.CSRFHash,
		session.UserAgent, session.ExpiresAt, session.LastSeenAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
