import { notFound } from "next/navigation";
import type { PublicUser } from "@/types/account";

export const dynamic = "force-dynamic";

export default async function PublicProfilePage({ params }: { params: Promise<{ handle: string }> }) {
  const { handle } = await params;
  const apiUrl = process.env.API_URL ?? "http://localhost:8080";
  const response = await fetch(`${apiUrl}/api/v1/users/${encodeURIComponent(handle)}`, { cache: "no-store" });
  if (response.status === 404) notFound();
  if (!response.ok) throw new Error("无法读取公开资料");
  const result = (await response.json()) as { user: PublicUser };
  return <main className="content"><div className="content-inner"><p className="brand" style={{ color: "#3730a3" }}>✦ 寰域</p><section className="panel profile-card"><div className="profile-mark">{result.user.displayName.slice(0, 1)}</div><div><h1 className="page-title">{result.user.displayName}</h1><p className="page-lead">@{result.user.handle}{result.user.location ? ` · ${result.user.location}` : ""}</p><p>{result.user.bio || "这个用户还没有填写个人简介。"}</p></div></section></div></main>;
}
