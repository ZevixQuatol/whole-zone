# 寰域 HuanYu

以寰点连接内容、域与人的沟通平台。

当前开发阶段：Account 模块邀请内测，包括邀请注册、安全登录会话、公开资料、密码管理、系统通知和平台邀请管理。

## 工程目录

- `api/`：Go + Gin API。
- `app/`：React + Next.js + TypeScript 桌面端。
- `docs/`：产品规格、实施计划和开发说明。

详细启动与验证步骤见 [docs/development.md](docs/development.md)。

后端首次配置后可直接启动：

```powershell
Copy-Item api/config.example.yml api/config.yml
# 编辑 api/config.yml
Set-Location api
go run ./cmd/api
```
