"use client";

import { FormEvent, useEffect, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";
import type { PublicUser } from "@/types/account";

export function ProfileForm() {
  const [user, setUser] = useState<PublicUser>();
  const [message, setMessage] = useState("");
  useEffect(() => { accountApi.me().then((result) => setUser(result.user)).catch(() => { window.location.href = "/login"; }); }, []);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setMessage("");
    const form = new FormData(event.currentTarget);
    try {
      const result = await accountApi.updateProfile({ displayName: String(form.get("displayName")), bio: String(form.get("bio")), location: String(form.get("location")) });
      setUser(result.user); setMessage("公开资料已更新");
    } catch (cause) { setMessage(cause instanceof ApiError ? cause.message : "更新失败"); }
  }
  if (!user) return <p className="page-lead">正在读取账号资料…</p>;
  return <form className="panel" onSubmit={submit}><div className="field"><label htmlFor="displayName">显示名称</label><input id="displayName" name="displayName" defaultValue={user.displayName} maxLength={64} required /></div><div className="field"><label htmlFor="bio">个人简介</label><textarea id="bio" name="bio" defaultValue={user.bio} maxLength={500} rows={6} /></div><div className="field"><label htmlFor="location">所在地区</label><input id="location" name="location" defaultValue={user.location} maxLength={80} /></div>{message ? <p className="form-success" role="status">{message}</p> : null}<button className="button">保存资料</button></form>;
}
