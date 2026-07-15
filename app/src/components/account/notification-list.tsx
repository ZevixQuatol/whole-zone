"use client";

import { useEffect, useState } from "react";
import { accountApi } from "@/lib/api/account";
import type { Notification } from "@/types/account";

export function NotificationList() {
  const [items, setItems] = useState<Notification[]>();
  useEffect(() => { accountApi.notifications().then((result) => setItems(result.items)).catch(() => { window.location.href = "/login"; }); }, []);
  async function read(item: Notification) { if (!item.readAt) { await accountApi.readNotification(item.id); setItems((current) => current?.map((value) => value.id === item.id ? { ...value, readAt: new Date().toISOString() } : value)); } }
  if (!items) return <p className="page-lead">正在读取通知…</p>;
  if (!items.length) return <div className="panel"><p className="page-lead">暂时没有系统通知。</p></div>;
  return <div className="panel">{items.map((item) => <button className="row notification" key={item.id} onClick={() => read(item)}><span><strong>{item.title}</strong><small>{item.body}</small></span><time>{new Date(item.createdAt).toLocaleString("zh-CN")}</time></button>)}</div>;
}
