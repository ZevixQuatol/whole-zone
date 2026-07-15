"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";
import type { Invite } from "@/types/account";

export function InvitePanel() {
  const router = useRouter();
  const [items, setItems] = useState<Invite[]>([]);
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    accountApi.invites()
      .then((result) => setItems(result.items))
      .catch((cause) => {
        if (cause instanceof ApiError && cause.status === 401) {
          router.replace("/login");
          return;
        }
        setError(cause instanceof ApiError ? cause.message : "暂时无法读取邀请码");
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const result = await accountApi.createInvite(
        Number(form.get("maxUses")),
        Number(form.get("ttlDays")),
      );
      setItems((current) => [result.invite, ...current]);
      setCode(result.code);
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "暂时无法创建邀请码");
    }
  }

  async function disable(id: string) {
    setError("");
    try {
      await accountApi.disableInvite(id);
      setItems((current) => current.map((item) => (
        item.id === id ? { ...item, disabledAt: new Date().toISOString() } : item
      )));
    } catch (cause) {
      setError(cause instanceof ApiError ? cause.message : "暂时无法停用邀请码");
    }
  }

  if (loading) return <p className="page-lead">正在读取邀请码…</p>;
  if (error && !items.length) {
    return <div className="panel"><p className="form-error" role="alert">{error}</p></div>;
  }

  return (
    <>
      <form className="panel" onSubmit={create}>
        <h2>创建邀请码</h2>
        <div className="form-grid">
          <div className="field">
            <label htmlFor="maxUses">可使用次数</label>
            <input id="maxUses" name="maxUses" type="number" min={1} max={1000} defaultValue={1} required />
          </div>
          <div className="field">
            <label htmlFor="ttlDays">有效天数</label>
            <input id="ttlDays" name="ttlDays" type="number" min={1} max={365} defaultValue={7} required />
          </div>
        </div>
        {error ? <p className="form-error" role="alert">{error}</p> : null}
        <button className="button">创建邀请码</button>
        {code ? <p className="invite-code" aria-label="新邀请码">{code}</p> : null}
      </form>
      <section className="panel">
        <h2>邀请码记录</h2>
        {!items.length ? <p className="page-lead">还没有邀请码。</p> : null}
        {items.map((item) => (
          <div className="row" key={item.id}>
            <span>{item.uses} / {item.maxUses} 次 · 到期 {new Date(item.expiresAt).toLocaleDateString("zh-CN")}</span>
            <button className="button secondary" disabled={Boolean(item.disabledAt)} onClick={() => disable(item.id)}>
              {item.disabledAt ? "已停用" : "停用"}
            </button>
          </div>
        ))}
      </section>
    </>
  );
}
