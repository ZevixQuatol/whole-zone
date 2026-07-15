COMMENT ON TABLE schema_migrations IS NULL;
COMMENT ON COLUMN schema_migrations.version IS NULL;
COMMENT ON COLUMN schema_migrations.name IS NULL;
COMMENT ON COLUMN schema_migrations.applied_at IS NULL;

COMMENT ON TABLE users IS NULL;
COMMENT ON COLUMN users.id IS NULL;
COMMENT ON COLUMN users.email IS NULL;
COMMENT ON COLUMN users.handle IS NULL;
COMMENT ON COLUMN users.password_hash IS NULL;
COMMENT ON COLUMN users.role IS NULL;
COMMENT ON COLUMN users.status IS NULL;
COMMENT ON COLUMN users.display_name IS NULL;
COMMENT ON COLUMN users.bio IS NULL;
COMMENT ON COLUMN users.location IS NULL;
COMMENT ON COLUMN users.created_at IS NULL;
COMMENT ON COLUMN users.updated_at IS NULL;
COMMENT ON COLUMN users.deleted_at IS NULL;

COMMENT ON TABLE invites IS NULL;
COMMENT ON COLUMN invites.id IS NULL;
COMMENT ON COLUMN invites.code_hash IS NULL;
COMMENT ON COLUMN invites.created_by IS NULL;
COMMENT ON COLUMN invites.max_uses IS NULL;
COMMENT ON COLUMN invites.uses IS NULL;
COMMENT ON COLUMN invites.expires_at IS NULL;
COMMENT ON COLUMN invites.disabled_at IS NULL;
COMMENT ON COLUMN invites.created_at IS NULL;

COMMENT ON TABLE sessions IS NULL;
COMMENT ON COLUMN sessions.id IS NULL;
COMMENT ON COLUMN sessions.user_id IS NULL;
COMMENT ON COLUMN sessions.token_hash IS NULL;
COMMENT ON COLUMN sessions.csrf_hash IS NULL;
COMMENT ON COLUMN sessions.user_agent IS NULL;
COMMENT ON COLUMN sessions.expires_at IS NULL;
COMMENT ON COLUMN sessions.last_seen_at IS NULL;
COMMENT ON COLUMN sessions.revoked_at IS NULL;
COMMENT ON COLUMN sessions.created_at IS NULL;

COMMENT ON TABLE password_resets IS NULL;
COMMENT ON COLUMN password_resets.id IS NULL;
COMMENT ON COLUMN password_resets.user_id IS NULL;
COMMENT ON COLUMN password_resets.token_hash IS NULL;
COMMENT ON COLUMN password_resets.expires_at IS NULL;
COMMENT ON COLUMN password_resets.used_at IS NULL;
COMMENT ON COLUMN password_resets.created_at IS NULL;

COMMENT ON TABLE system_notifications IS NULL;
COMMENT ON COLUMN system_notifications.id IS NULL;
COMMENT ON COLUMN system_notifications.user_id IS NULL;
COMMENT ON COLUMN system_notifications.kind IS NULL;
COMMENT ON COLUMN system_notifications.title IS NULL;
COMMENT ON COLUMN system_notifications.body IS NULL;
COMMENT ON COLUMN system_notifications.data IS NULL;
COMMENT ON COLUMN system_notifications.read_at IS NULL;
COMMENT ON COLUMN system_notifications.created_at IS NULL;

COMMENT ON TABLE account_events IS NULL;
COMMENT ON COLUMN account_events.id IS NULL;
COMMENT ON COLUMN account_events.user_id IS NULL;
COMMENT ON COLUMN account_events.kind IS NULL;
COMMENT ON COLUMN account_events.data IS NULL;
COMMENT ON COLUMN account_events.created_at IS NULL;
