import Link from "next/link";
import type { Me } from "@/types/account";

export function AccountNav({ role }: { role: Me["role"] }) {
  return (
    <nav>
      <Link href="/settings/profile">公开资料</Link>
      <Link href="/settings/account">账号安全</Link>
      <Link href="/notifications">系统通知</Link>
      {role === "platform_admin" ? <Link href="/admin/invites">邀请管理</Link> : null}
    </nav>
  );
}
