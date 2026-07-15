import { AccountShell } from "@/components/account/account-shell";
import { SecurityPanel } from "@/components/account/security-panel";

export default function AccountPage() { return <AccountShell><h1 className="page-title">账号安全</h1><p className="page-lead">管理密码和有效登录会话。</p><SecurityPanel /></AccountShell>; }
