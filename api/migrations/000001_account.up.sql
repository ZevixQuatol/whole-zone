CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    handle text NOT NULL,
    password_hash text NOT NULL,
    role text NOT NULL DEFAULT 'member' CHECK (role IN ('member', 'platform_admin')),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'deleted')),
    display_name text NOT NULL,
    bio text NOT NULL DEFAULT '',
    location text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz,
    CHECK (char_length(email) BETWEEN 3 AND 254),
    CHECK (handle ~ '^[a-z0-9_]{3,32}$'),
    CHECK (char_length(display_name) BETWEEN 1 AND 64),
    CHECK (char_length(bio) <= 500),
    CHECK (char_length(location) <= 80)
);

CREATE UNIQUE INDEX users_email_uq ON users (lower(email)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_handle_uq ON users (lower(handle)) WHERE deleted_at IS NULL;

CREATE TABLE invites (
    id uuid PRIMARY KEY,
    code_hash bytea NOT NULL UNIQUE,
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    max_uses integer NOT NULL CHECK (max_uses > 0),
    uses integer NOT NULL DEFAULT 0 CHECK (uses >= 0 AND uses <= max_uses),
    expires_at timestamptz NOT NULL,
    disabled_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    csrf_hash bytea NOT NULL,
    user_agent text NOT NULL DEFAULT '',
    expires_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX sessions_user_idx ON sessions (user_id, created_at DESC);

CREATE TABLE password_resets (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX password_resets_user_idx ON password_resets (user_id, created_at DESC);

CREATE TABLE system_notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    read_at timestamptz,
    created_at timestamptz NOT NULL,
    CHECK (char_length(kind) BETWEEN 1 AND 50),
    CHECK (char_length(title) BETWEEN 1 AND 120),
    CHECK (char_length(body) <= 1000)
);

CREATE INDEX system_notifications_user_idx
    ON system_notifications (user_id, created_at DESC);

CREATE TABLE account_events (
    id uuid PRIMARY KEY,
    user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    kind text NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL,
    CHECK (char_length(kind) BETWEEN 1 AND 50)
);

CREATE INDEX account_events_user_idx ON account_events (user_id, created_at DESC);
