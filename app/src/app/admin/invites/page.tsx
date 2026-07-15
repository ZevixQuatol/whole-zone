import { AccountShell } from "@/components/account/account-shell";
import { InvitePanel } from "@/components/account/invite-panel";

export default function InvitesPage() { return <AccountShell><h1 className="page-title">邀请管理</h1><p className="page-lead">只有平台管理员可以创建和停用内测邀请码。</p><InvitePanel /></AccountShell>; }
