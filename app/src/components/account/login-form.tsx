"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";

export function LoginForm() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      await accountApi.login({ login: String(form.get("login")), password: String(form.get("password")) });
      router.push("/settings/profile");
      router.refresh();
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "暂时无法登录，请稍后重试");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="form-panel" onSubmit={submit}>
      <h2>登录寰域</h2>
      <p>继续维护你的新时代名片，进入已经加入的域。</p>
      {error ? <p className="form-error" role="alert">{error}</p> : null}
      <div className="field"><label htmlFor="login">邮箱或用户名</label><input id="login" name="login" autoComplete="username" required /></div>
      <div className="field"><label htmlFor="password">密码</label><input id="password" name="password" type="password" autoComplete="current-password" required /></div>
      <button className="button" disabled={busy} type="submit">{busy ? "登录中…" : "登录"}</button>
      <div className="form-links"><Link href="/register">使用邀请码注册</Link><Link href="/forgot-password">忘记密码</Link></div>
    </form>
  );
}
