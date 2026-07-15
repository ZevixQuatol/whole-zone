"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { accountApi } from "@/lib/api/account";
import { ApiError } from "@/lib/api/client";
import type { Me } from "@/types/account";
import { AccountNav } from "./account-nav";

export function AccountShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [me, setMe] = useState<Me>();
  const [error, setError] = useState("");

  useEffect(() => {
    accountApi.me()
      .then(setMe)
      .catch((cause) => {
        if (cause instanceof ApiError && cause.status === 401) {
          router.replace("/login");
          return;
        }
        setError(cause instanceof ApiError ? cause.message : "暂时无法读取账号信息");
      });
  }, [router]);

  return (
    <div className="shell">
      <aside className="sidebar">
        <Link className="brand" href="/">✦ 寰域</Link>
        {me ? <AccountNav role={me.role} /> : null}
      </aside>
      <main className="content">
        <div className="content-inner">
          {error ? <div className="panel"><p className="form-error" role="alert">{error}</p></div> : null}
          {!error && !me ? <p className="page-lead">正在验证登录状态…</p> : null}
          {me ? children : null}
        </div>
      </main>
    </div>
  );
}
