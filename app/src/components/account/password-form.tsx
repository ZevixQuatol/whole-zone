"use client";

import { FormEvent, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";

export function ForgotPasswordForm() {
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setBusy(true); setMessage("");
    const email = String(new FormData(event.currentTarget).get("email"));
    try {
      await accountApi.forgotPassword(email);
      setMessage("如果账号存在，重置邮件已经发送。");
    } catch (cause) {
      setMessage(cause instanceof ApiError ? cause.message : "请求失败，请稍后重试");
    } finally { setBusy(false); }
  }
  return <form className="form-panel" onSubmit={submit}><h2>找回密码</h2><p>输入注册邮箱。为保护账号隐私，无论账号是否存在，页面都会显示相同结果。</p>{message ? <p className="form-success" role="status">{message}</p> : null}<div className="field"><label htmlFor="email">邮箱</label><input id="email" name="email" type="email" required /></div><button className="button" disabled={busy}>{busy ? "提交中…" : "发送重置邮件"}</button></form>;
}

export function ResetPasswordForm({ token }: { token: string }) {
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setBusy(true); setMessage("");
    const password = String(new FormData(event.currentTarget).get("password"));
    try {
      await accountApi.resetPassword(token, password);
      setMessage("密码已更新，请返回登录页面。");
    } catch (cause) {
      setMessage(cause instanceof ApiError ? cause.message : "请求失败，请稍后重试");
    } finally { setBusy(false); }
  }
  return <form className="form-panel" onSubmit={submit}><h2>设置新密码</h2><p>新密码至少 12 个字符。重置成功后，已有会话全部失效。</p>{message ? <p className="form-success" role="status">{message}</p> : null}<div className="field"><label htmlFor="password">新密码</label><input id="password" name="password" type="password" minLength={12} maxLength={128} required /></div><button className="button" disabled={busy}>{busy ? "更新中…" : "更新密码"}</button></form>;
}
