import type { Invite, Me, Notification, PublicUser } from "@/types/account";
import { request } from "./client";

export const accountApi = {
  register: (input: { inviteCode: string; email: string; handle: string; password: string; displayName: string }) =>
    request<{ user: PublicUser }>("/account/register", { method: "POST", body: JSON.stringify(input) }),
  login: (input: { login: string; password: string }) =>
    request<{ user: PublicUser }>("/account/login", { method: "POST", body: JSON.stringify(input) }),
  me: () => request<Me>("/account/me"),
  updateProfile: (input: { displayName: string; bio: string; location: string }) =>
    request<{ user: PublicUser }>("/account/profile", { method: "PATCH", body: JSON.stringify(input) }),
  changePassword: (input: { currentPassword: string; newPassword: string }) =>
    request<void>("/account/password/change", { method: "POST", body: JSON.stringify(input) }),
  forgotPassword: (email: string) =>
    request<void>("/account/password/forgot", { method: "POST", body: JSON.stringify({ email }) }),
  resetPassword: (token: string, password: string) =>
    request<void>("/account/password/reset", { method: "POST", body: JSON.stringify({ token, password }) }),
  logout: () => request<void>("/account/logout", { method: "POST" }),
  logoutAll: () => request<void>("/account/logout-all", { method: "POST" }),
  publicUser: (handle: string) => request<{ user: PublicUser }>(`/users/${encodeURIComponent(handle)}`),
  notifications: () => request<{ items: Notification[] }>("/account/notifications"),
  readNotification: (id: string) => request<void>(`/account/notifications/${id}/read`, { method: "PATCH" }),
  invites: () => request<{ items: Invite[] }>("/admin/invites"),
  createInvite: (maxUses: number, ttlDays: number) =>
    request<{ invite: Invite; code: string }>("/admin/invites", {
      method: "POST",
      body: JSON.stringify({ maxUses, ttlDays }),
    }),
  disableInvite: (id: string) => request<void>(`/admin/invites/${id}/disable`, { method: "POST" }),
};
