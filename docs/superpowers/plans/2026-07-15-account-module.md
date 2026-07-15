# Account 模块实施计划

> 执行要求：逐任务测试先行；每个任务先看到针对目标行为的失败测试，再写最小实现，通过后再整理结构。不得以空目录、占位 handler 或未接通的页面作为功能完成。

## 目标

在项目根目录 `api/` 和 `app/` 中交付可运行的 Account 纵向模块，包含邀请注册、邮箱或用户名登录、安全会话、退出、退出全部设备、公开资料、密码修改与重置、最小系统通知、管理员邀请码管理和前后端联调。

## 架构

- `api/` 是 Go 1.25 + Gin 模块化单体的服务端。
- `app/` 是 Next.js 16 + React 19 + TypeScript 的桌面端。
- Account 核心事实保存在 PostgreSQL；Redis 不作为本模块的必需事实依赖。
- 会话使用高熵随机令牌；浏览器只保存原始令牌，数据库只保存 SHA-256 摘要。
- 密码使用 Argon2id。
- 状态变更使用 HttpOnly 会话 Cookie、同源校验和 CSRF Token。
- API 使用 `/api/v1`，错误响应遵循正式规格。
- PostgreSQL 集成测试读取 `TEST_DATABASE_URL`；未配置时明确跳过，不能伪装成通过。

## 依赖版本

- Go 1.25。
- Gin `v1.12.0`。
- pgx `v5.10.0`。
- Argon2id `v1.0.0`。
- google/uuid `v1.6.0`。
- Next.js `16.2.10`。
- React `19.2.7`。
- pnpm `10.32.1`。
- Vitest `4.1.10`。
- Playwright `1.61.1`。

## API 契约

### 公共接口

- `POST /api/v1/account/register`
- `POST /api/v1/account/login`
- `POST /api/v1/account/password/forgot`
- `POST /api/v1/account/password/reset`
- `GET /api/v1/users/:handle`

### 登录接口

- `GET /api/v1/account/me`
- `PATCH /api/v1/account/profile`
- `POST /api/v1/account/password/change`
- `POST /api/v1/account/logout`
- `POST /api/v1/account/logout-all`
- `GET /api/v1/account/notifications`
- `PATCH /api/v1/account/notifications/:id/read`

### 管理接口

- `GET /api/v1/admin/invites`
- `POST /api/v1/admin/invites`
- `POST /api/v1/admin/invites/:id/disable`

## 数据表

### `users`

- `id uuid primary key`
- `email text unique not null`
- `handle text unique not null`
- `password_hash text not null`
- `role text not null`
- `status text not null`
- `display_name text not null`
- `bio text not null default ''`
- `location text not null default ''`
- `created_at timestamptz not null`
- `updated_at timestamptz not null`
- `deleted_at timestamptz null`

### `invites`

- `id uuid primary key`
- `code_hash bytea unique not null`
- `created_by uuid null references users(id)`
- `max_uses integer not null`
- `uses integer not null default 0`
- `expires_at timestamptz not null`
- `disabled_at timestamptz null`
- `created_at timestamptz not null`

### `sessions`

- `id uuid primary key`
- `user_id uuid not null references users(id)`
- `token_hash bytea unique not null`
- `csrf_hash bytea not null`
- `user_agent text not null default ''`
- `expires_at timestamptz not null`
- `last_seen_at timestamptz not null`
- `revoked_at timestamptz null`
- `created_at timestamptz not null`

### `password_resets`

- `id uuid primary key`
- `user_id uuid not null references users(id)`
- `token_hash bytea unique not null`
- `expires_at timestamptz not null`
- `used_at timestamptz null`
- `created_at timestamptz not null`

### `system_notifications`

- `id uuid primary key`
- `user_id uuid not null references users(id)`
- `kind text not null`
- `title text not null`
- `body text not null`
- `data jsonb not null default '{}'`
- `read_at timestamptz null`
- `created_at timestamptz not null`

### `account_events`

- `id uuid primary key`
- `user_id uuid null references users(id)`
- `kind text not null`
- `data jsonb not null default '{}'`
- `created_at timestamptz not null`

## 任务 1：同步规格和工程基础

**文件**

- 修改：`docs/superpowers/specs/2026-07-15-huanyu-platform-design.md`
- 新建：`.gitignore`
- 新建：`.editorconfig`
- 新建：`api/go.mod`
- 新建：`api/.env.example`
- 新建：`api/cmd/api/main.go`
- 新建：`api/internal/config/config.go`
- 新建：`api/internal/config/config_test.go`
- 新建：`app/package.json`
- 新建：`app/tsconfig.json`
- 新建：`app/next.config.ts`
- 新建：`app/.env.local.example`

**步骤**

1. 先写配置测试，覆盖必填 `DATABASE_URL`、默认监听地址、Cookie 安全开关和会话时长。
2. 运行 `cd api; go test ./internal/config`，确认因缺少实现失败。
3. 编写最小配置加载器并通过测试。
4. 初始化 Go 和 Next.js 依赖，固定本计划中的版本。
5. `.gitignore` 排除 `.superpowers/`、环境文件、构建产物和测试缓存，不排除示例配置。
6. 运行 `cd api; go test ./...` 与 `cd app; pnpm exec tsc --noEmit`。

## 任务 2：数据库迁移与存储连接

**文件**

- 新建：`api/migrations/000001_account.up.sql`
- 新建：`api/migrations/000001_account.down.sql`
- 新建：`api/internal/store/postgres.go`
- 新建：`api/internal/store/migrate.go`
- 新建：`api/internal/store/migrate_test.go`

**步骤**

1. 写迁移测试，断言六张 Account 表、唯一约束、外键和检查约束存在。
2. 在没有 `TEST_DATABASE_URL` 时测试明确 `Skip`；有地址时必须创建隔离 schema 后真实执行 up/down/up。
3. 写 SQL 迁移和迁移执行器。
4. 运行 `cd api; go test ./internal/store -v`。
5. 运行 `cd api; go test ./...`。

## 任务 3：密码、令牌和 Account 领域模型

**文件**

- 新建：`api/internal/account/model.go`
- 新建：`api/internal/account/errors.go`
- 新建：`api/internal/account/password.go`
- 新建：`api/internal/account/password_test.go`
- 新建：`api/internal/account/token.go`
- 新建：`api/internal/account/token_test.go`

**步骤**

1. 写密码测试：Argon2id 哈希不含明文、正确密码通过、错误密码失败、损坏哈希返回明确错误。
2. 写令牌测试：随机令牌具有足够熵、只持久化 SHA-256 摘要、CSRF 摘要验证正确。
3. 运行目标测试确认失败。
4. 实现密码和令牌原语；对参数选择和只存摘要的原因添加关键注释。
5. 运行 `cd api; go test ./internal/account -run 'Password|Token' -v`。

## 任务 4：Account Repository

**文件**

- 新建：`api/internal/account/repo.go`
- 新建：`api/internal/account/pg.go`
- 新建：`api/internal/account/pg_test.go`
- 新建：`api/internal/account/fake_test.go`

**步骤**

1. 定义面向用例的简洁 Repository 接口，不泄漏 Gin 或 pgx 类型。
2. 写 PostgreSQL 集成测试：邀请码并发消耗、邮箱/用户名唯一、会话撤销、重置令牌单次使用、通知归属。
3. 运行测试确认没有实现时失败。
4. 使用 pgx 实现事务化 Repository。
5. 邀请码消耗必须使用行锁或等效原子更新，避免超额注册。
6. 运行 `cd api; go test ./internal/account -run PG -v` 和全部 Go 测试。

## 任务 5：邀请注册和登录用例

**文件**

- 新建：`api/internal/account/service.go`
- 新建：`api/internal/account/register_test.go`
- 新建：`api/internal/account/login_test.go`

**步骤**

1. 注册测试覆盖：邀请码有效、过期、停用、次数耗尽；邮箱和用户名规范化；重复冲突；弱密码拒绝；成功后生成会话。
2. 登录测试覆盖：邮箱登录、用户名登录、错误密码统一错误、停用账号拒绝、成功后生成新会话。
3. 运行目标测试确认失败。
4. 实现最小 Service；注册在单个事务中创建用户、消耗邀请码、写 Account 事件和创建会话。
5. 登录错误不能泄漏账号是否存在。
6. 运行 `cd api; go test ./internal/account -run 'Register|Login' -v`。

## 任务 6：认证中间件、退出和资料

**文件**

- 新建：`api/internal/account/auth.go`
- 新建：`api/internal/account/auth_test.go`
- 新建：`api/internal/account/profile_test.go`

**步骤**

1. 测试会话 Cookie、过期/撤销会话、同源校验、CSRF、退出当前会话和退出全部会话。
2. 测试读取本人、更新 displayName/bio/location、按 handle 读取公开资料。
3. 运行测试确认失败。
4. 实现认证和资料用例；更新资料使用乐观时间戳或等效冲突保护。
5. 运行 `cd api; go test ./internal/account -run 'Auth|Logout|Profile' -v`。

## 任务 7：密码修改与重置

**文件**

- 新建：`api/internal/account/mail.go`
- 新建：`api/internal/account/password_flow_test.go`
- 新建：`api/internal/account/smtp.go`

**步骤**

1. 测试登录用户修改密码后撤销其他会话。
2. 测试忘记密码始终返回相同外部响应，存在账号时生成短期单次令牌并调用 Mailer。
3. 测试重置令牌过期、已使用、无效和成功路径。
4. 运行测试确认失败。
5. 实现 Mailer 接口、SMTP 适配器和开发环境安全日志适配器；日志只能输出邮件 ID，不能输出原始重置令牌。
6. 运行 `cd api; go test ./internal/account -run PasswordFlow -v`。

## 任务 8：通知和管理员邀请码

**文件**

- 新建：`api/internal/account/notification_test.go`
- 新建：`api/internal/account/invite_test.go`
- 新建：`api/cmd/admin/main.go`

**步骤**

1. 测试通知分页、只读本人通知、标记已读幂等。
2. 测试只有平台管理员可列出、创建和停用邀请码。
3. 测试管理员 CLI 可以幂等创建首个管理员和首个邀请码。
4. 运行测试确认失败。
5. 实现用例和 CLI；CLI 密码从交互输入或环境读取，不接受命令行明文参数。
6. 运行 `cd api; go test ./internal/account -run 'Notification|Invite|Admin' -v`。

## 任务 9：Gin Handler、统一错误和 OpenAPI

**文件**

- 新建：`api/internal/account/handler.go`
- 新建：`api/internal/account/handler_test.go`
- 新建：`api/internal/web/response.go`
- 新建：`api/internal/web/request.go`
- 新建：`api/internal/web/origin.go`
- 新建：`api/openapi.yaml`
- 修改：`api/cmd/api/main.go`

**步骤**

1. 使用 `httptest` 写全部 Account 路由的请求、响应、Cookie、CSRF、错误码和字段错误测试。
2. 运行 handler 测试确认失败。
3. 实现 handler 和统一错误协议；handler 只做协议转换，不承载业务规则。
4. 写 OpenAPI Account 契约并校验路径与实现一致。
5. 运行 `cd api; go test ./...`、`go vet ./...`。

## 任务 10：Next.js 基础和类型安全 API 客户端

**文件**

- 新建：`app/src/app/layout.tsx`
- 新建：`app/src/app/globals.css`
- 新建：`app/src/lib/api/client.ts`
- 新建：`app/src/lib/api/account.ts`
- 新建：`app/src/lib/api/client.test.ts`
- 新建：`app/src/types/account.ts`
- 新建：`app/vitest.config.ts`
- 新建：`app/src/test/setup.ts`

**步骤**

1. 写客户端测试：成功 JSON、统一错误、401、CSRF Header、Cookie 凭据和网络失败。
2. 运行 `cd app; pnpm test --run` 确认失败。
3. 实现最小 API 客户端和 Account 类型。
4. 建立桌面端基础样式，不添加移动端断点。
5. 运行 `pnpm test --run`、`pnpm exec tsc --noEmit`。

## 任务 11：登录、注册和密码页面

**文件**

- 新建：`app/src/app/(auth)/login/page.tsx`
- 新建：`app/src/app/(auth)/register/page.tsx`
- 新建：`app/src/app/(auth)/forgot-password/page.tsx`
- 新建：`app/src/app/(auth)/reset-password/page.tsx`
- 新建：`app/src/components/account/auth-form.tsx`
- 新建：`app/src/components/account/auth-form.test.tsx`

**步骤**

1. 写组件测试：字段校验、服务端字段错误、提交禁用、防重复提交、成功跳转、草稿保留。
2. 运行测试确认失败。
3. 实现页面和共享表单。
4. 页面使用明确中文文案，登录错误不泄漏账号存在性。
5. 运行前端测试、类型检查和 lint。

## 任务 12：账号设置、公开资料、通知和邀请管理页面

**文件**

- 新建：`app/src/app/settings/account/page.tsx`
- 新建：`app/src/app/settings/profile/page.tsx`
- 新建：`app/src/app/notifications/page.tsx`
- 新建：`app/src/app/users/[handle]/page.tsx`
- 新建：`app/src/app/admin/invites/page.tsx`
- 新建：`app/src/components/account/profile-form.tsx`
- 新建：`app/src/components/account/notification-list.tsx`
- 新建：`app/src/components/account/invite-panel.tsx`
- 新建：对应 `*.test.tsx`

**步骤**

1. 先写资料更新、密码修改、公开资料、通知已读和管理员邀请码测试。
2. 运行测试确认失败。
3. 实现桌面页面和受保护布局。
4. 非管理员不能看到或访问邀请管理。
5. 运行 `pnpm test --run`、`pnpm exec tsc --noEmit`、`pnpm lint`。

## 任务 13：端到端联调与文档

**文件**

- 新建：`app/playwright.config.ts`
- 新建：`app/e2e/account.spec.ts`
- 新建：`docs/development.md`
- 新建或重写：`README.md`

**步骤**

1. E2E 覆盖：管理员创建邀请码、用户注册、登录、修改资料、公开资料、退出、重新登录、退出全部设备、密码重置和通知已读。
2. 使用专用测试数据库执行迁移；测试数据不写入开发数据库。
3. 写 Windows/PowerShell 友好的启动、迁移、管理员初始化和测试命令。
4. 运行完整验证：

```text
cd api
go test ./...
go vet ./...

cd ../app
pnpm test --run
pnpm exec tsc --noEmit
pnpm lint
pnpm build
pnpm exec playwright test
```

5. 对照 Account 功能点逐项手工审计，不以仅有页面或仅有 API 作为完成。

## 完成标准

- `api/` 和 `app/` 目录结构与规格一致。
- Account 所有 API 有自动化测试和 OpenAPI 契约。
- 邀请、注册、登录、会话、退出、资料、密码、通知和管理员邀请页面全部接通真实 API。
- 数据库集成测试在提供 PostgreSQL 地址时通过；缺少地址时有明确说明。
- 前后端单元测试、类型检查、lint、构建和 Account E2E 全部通过。
- 关键安全逻辑有简洁而有意义的注释。
- 变量、文件、表和字段命名符合简洁命名约束。
