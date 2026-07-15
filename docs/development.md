# 本地开发

## 目录

- `api/`：Go + Gin Account API。
- `app/`：Next.js 桌面端。

## 基础服务

项目使用 PostgreSQL 18 和 Redis。服务可以由外部 Docker Compose 管理；项目代码不依赖或修改共享 Compose 目录。

复制示例配置：

```powershell
Copy-Item api/.env.example api/.env
Copy-Item app/.env.local.example app/.env.local
```

把 `DATABASE_URL`、`REDIS_URL` 和 SMTP 参数改为本机实际值。不要提交真实密码。

SMTP 未配置或发送失败时，找回密码接口仍会返回统一结果，具体错误会写入 API 日志，避免泄露账号是否存在。要实际收到重置邮件，必须正确配置 `SMTP_HOST`、`SMTP_PORT` 和发件人信息。

## 初始化管理员

在 PowerShell 中设置管理员环境变量后执行：

```powershell
Set-Location api
$env:DATABASE_URL = 'postgres://postgres:password@localhost:5432/huanyu?sslmode=disable'
$env:ADMIN_EMAIL = 'admin@example.com'
$env:ADMIN_HANDLE = 'admin'
$env:ADMIN_PASSWORD = 'replace-with-a-strong-password'
$env:ADMIN_DISPLAY_NAME = '平台管理员'
go run ./cmd/admin bootstrap
```

命令会自动执行迁移，并且可以幂等重复运行。

## 启动 API

```powershell
Set-Location api
$env:DATABASE_URL = 'postgres://postgres:password@localhost:5432/huanyu?sslmode=disable'
go run ./cmd/api
```

存活检查：`http://localhost:8080/health/live`。

## 启动桌面端

```powershell
Set-Location app
pnpm install
pnpm dev
```

打开 `http://localhost:3000`。

## 验证

```powershell
Set-Location api
go test ./...
go vet ./...

Set-Location ../app
pnpm test --run
pnpm exec tsc --noEmit
pnpm lint
pnpm build
```

设置 `TEST_DATABASE_URL` 后，PostgreSQL 集成测试会运行；未设置时会明确跳过。

### 桌面端浏览器验收

API 与已初始化的平台管理员可用后，设置以下变量运行真实 Chrome 主流程：

```powershell
Set-Location app
$env:E2E_API_URL = 'http://127.0.0.1:18080'
$env:E2E_ADMIN_LOGIN = 'admin@example.com'
$env:E2E_ADMIN_PASSWORD = 'replace-with-a-strong-password'
pnpm test:e2e
```

测试会在 `http://127.0.0.1:3100` 自动启动独立 Next.js 开发服务，并使用 1440 × 900 桌面视口。测试账号和邀请码仅用于当前测试数据库。
