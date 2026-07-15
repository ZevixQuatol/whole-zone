"use client";

import { FormEvent, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";

export function SecurityPanel() {
  const [message, setMessage] = useState("");
  async function changePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); const form = new FormData(event.currentTarget);
    try { await accountApi.changePassword({ currentPassword: String(form.get("current")), newPassword: String(form.get("next")) }); window.location.href = "/login"; }
    catch (cause) { setMessage(cause instanceof ApiError ? cause.message : "修改失败"); }
  }
  async function end(all: boolean) { try { if (all) await accountApi.logoutAll(); else await accountApi.logout(); window.location.href = "/login"; } catch (cause) { setMessage(cause instanceof ApiError ? cause.message : "操作失败"); } }
  return <><form className="panel" onSubmit={changePassword}><h2>修改密码</h2><div className="field"><label htmlFor="current">当前密码</label><input id="current" name="current" type="password" required /></div><div className="field"><label htmlFor="next">新密码</label><input id="next" name="next" type="password" minLength={12} maxLength={128} required /></div>{message ? <p className="form-error">{message}</p> : null}<button className="button">更新并重新登录</button></form><section className="panel"><h2>登录会话</h2><p className="page-lead">结束当前设备，或撤销账号的全部有效会话。</p><div style={{ display: "flex", gap: 12 }}><button className="button secondary" onClick={() => end(false)}>退出当前设备</button><button className="button secondary" onClick={() => end(true)}>退出全部设备</button></div></section></>;
}
