"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";

export function RegisterForm({ initialInvite = "" }: { initialInvite?: string }) {
  const router = useRouter();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      await accountApi.register({
        inviteCode: String(form.get("inviteCode")), email: String(form.get("email")),
        handle: String(form.get("handle")), password: String(form.get("password")),
        displayName: String(form.get("displayName")),
      });
      router.push("/settings/profile");
      router.refresh();
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "暂时无法注册，请稍后重试");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="form-panel" onSubmit={submit}>
      <h2>创建受邀账号</h2>
      <p>账号建立后，你可以添加寰点并逐步形成自己的可信名片。</p>
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      <div className="field"><label htmlFor="inviteCode">邀请码</label><input id="inviteCode" name="inviteCode" defaultValue={initialInvite} required /></div>
      <div className="field"><label htmlFor="email">邮箱</label><input id="email" name="email" type="email" autoComplete="email" required /></div>
      <div className="field"><label htmlFor="handle">用户名</label><input id="handle" name="handle" minLength={3} maxLength={32} pattern="[a-zA-Z0-9_]+" autoComplete="username" required /></div>
      <div className="field"><label htmlFor="displayName">显示名称</label><input id="displayName" name="displayName" maxLength={64} required /></div>
      <div className="field"><label htmlFor="password">密码</label><input id="password" name="password" type="password" minLength={12} maxLength={128} autoComplete="new-password" required /></div>
      <button className="button" disabled={busy} type="submit">{busy ? "创建中…" : "创建账号"}</button>
      <div className="form-links"><Link href="/login">已有账号，直接登录</Link></div>
    </form>
  );
}
