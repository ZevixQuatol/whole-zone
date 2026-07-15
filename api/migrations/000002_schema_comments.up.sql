COMMENT ON TABLE schema_migrations IS '数据库迁移记录';
COMMENT ON COLUMN schema_migrations.version IS '迁移版本号';
COMMENT ON COLUMN schema_migrations.name IS '迁移名称';
COMMENT ON COLUMN schema_migrations.applied_at IS '应用时间';

COMMENT ON TABLE users IS '用户账户';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.email IS '登录邮箱';
COMMENT ON COLUMN users.handle IS '用户名';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.role IS '平台角色';
COMMENT ON COLUMN users.status IS '账户状态';
COMMENT ON COLUMN users.display_name IS '显示名称';
COMMENT ON COLUMN users.bio IS '个人简介';
COMMENT ON COLUMN users.location IS '所在地区';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间';

COMMENT ON TABLE invites IS '注册邀请码';
COMMENT ON COLUMN invites.id IS '邀请码ID';
COMMENT ON COLUMN invites.code_hash IS '邀请码哈希';
COMMENT ON COLUMN invites.created_by IS '创建者用户ID';
COMMENT ON COLUMN invites.max_uses IS '最大使用次数';
COMMENT ON COLUMN invites.uses IS '已使用次数';
COMMENT ON COLUMN invites.expires_at IS '过期时间';
COMMENT ON COLUMN invites.disabled_at IS '停用时间';
COMMENT ON COLUMN invites.created_at IS '创建时间';

COMMENT ON TABLE sessions IS '用户登录会话';
COMMENT ON COLUMN sessions.id IS '会话ID';
COMMENT ON COLUMN sessions.user_id IS '用户ID';
COMMENT ON COLUMN sessions.token_hash IS '会话令牌哈希';
COMMENT ON COLUMN sessions.csrf_hash IS 'CSRF令牌哈希';
COMMENT ON COLUMN sessions.user_agent IS '客户端标识';
COMMENT ON COLUMN sessions.expires_at IS '过期时间';
COMMENT ON COLUMN sessions.last_seen_at IS '最后活跃时间';
COMMENT ON COLUMN sessions.revoked_at IS '撤销时间';
COMMENT ON COLUMN sessions.created_at IS '创建时间';

COMMENT ON TABLE password_resets IS '密码重置记录';
COMMENT ON COLUMN password_resets.id IS '重置记录ID';
COMMENT ON COLUMN password_resets.user_id IS '用户ID';
COMMENT ON COLUMN password_resets.token_hash IS '密码重置令牌哈希';
COMMENT ON COLUMN password_resets.expires_at IS '过期时间';
COMMENT ON COLUMN password_resets.used_at IS '使用时间';
COMMENT ON COLUMN password_resets.created_at IS '创建时间';

COMMENT ON TABLE system_notifications IS '系统通知';
COMMENT ON COLUMN system_notifications.id IS '通知ID';
COMMENT ON COLUMN system_notifications.user_id IS '接收用户ID';
COMMENT ON COLUMN system_notifications.kind IS '通知类型';
COMMENT ON COLUMN system_notifications.title IS '通知标题';
COMMENT ON COLUMN system_notifications.body IS '通知内容';
COMMENT ON COLUMN system_notifications.data IS '扩展数据';
COMMENT ON COLUMN system_notifications.read_at IS '阅读时间';
COMMENT ON COLUMN system_notifications.created_at IS '创建时间';

COMMENT ON TABLE account_events IS '账户审计事件';
COMMENT ON COLUMN account_events.id IS '事件ID';
COMMENT ON COLUMN account_events.user_id IS '关联用户ID';
COMMENT ON COLUMN account_events.kind IS '事件类型';
COMMENT ON COLUMN account_events.data IS '事件数据';
COMMENT ON COLUMN account_events.created_at IS '创建时间';
