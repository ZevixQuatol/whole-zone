# 本地开发

## 目录

- `api/`：Go + Gin Account API。
- `app/`：Next.js 桌面端。

## 基础服务

项目使用 PostgreSQL 18 和 Redis。服务可以由外部 Docker Compose 管理；项目代码不依赖或修改共享 Compose 目录。

后端统一读取 YAML。首次使用时复制示例文件：

```powershell
Copy-Item api/config.example.yml api/config.yml
Copy-Item app/.env.local.example app/.env.local
```

编辑 `api/config.yml`，填写 PostgreSQL、Redis、SMTP 和管理员账号。该文件已被 Git 忽略，不会提交真实密码；可提交的字段模板保留在 `api/config.example.yml`。

API 和管理员命令默认读取当前 `api` 目录下的 `config.yml`。需要使用其他文件时，只设置一个路径变量即可：

```powershell
$env:HUANYU_CONFIG = 'D:\Config\huanyu.yml'
```

SMTP 的 `host` 留空或发送失败时，找回密码接口仍会返回统一结果，具体错误会写入 API 日志，避免泄露账号是否存在。要实际收到重置邮件，需要正确填写 `smtp` 配置。

## 初始化管理员

先在 `config.yml` 的 `admin` 节点填写管理员账号，然后直接执行：

```powershell
Set-Location api
go run ./cmd/admin bootstrap
```

命令会自动执行迁移，并且可以幂等重复运行。

## 只执行数据库迁移

API 启动时会自动迁移；如果只想更新数据库结构或注释而不启动服务，可以执行：

```powershell
Set-Location api
go run ./cmd/migrate
```

## 启动 API

```powershell
Set-Location api
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
