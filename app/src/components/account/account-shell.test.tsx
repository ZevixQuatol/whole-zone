import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api/client";
import { AccountShell } from "./account-shell";

const mocks = vi.hoisted(() => ({
  me: vi.fn(),
  replace: vi.fn(),
}));

vi.mock("@/lib/api/account", () => ({
  accountApi: { me: mocks.me },
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mocks.replace }),
}));

describe("AccountShell", () => {
  beforeEach(() => {
    mocks.me.mockReset();
    mocks.replace.mockReset();
  });

  it("waits for authentication before rendering protected content", () => {
    mocks.me.mockReturnValue(new Promise(() => undefined));

    render(<AccountShell><div>账号内容</div></AccountShell>);

    expect(screen.getByText("正在验证登录状态…")).toBeInTheDocument();
    expect(screen.queryByText("账号内容")).not.toBeInTheDocument();
  });

  it("renders member navigation without the admin entry", async () => {
    mocks.me.mockResolvedValue({
      role: "member",
      user: { handle: "dev", displayName: "开发者", bio: "", location: "", createdAt: "2026-07-15T00:00:00Z" },
    });

    render(<AccountShell><div>账号内容</div></AccountShell>);

    expect(await screen.findByText("账号内容")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "账号安全" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "邀请管理" })).not.toBeInTheDocument();
  });

  it("shows the invite entry to platform administrators", async () => {
    mocks.me.mockResolvedValue({
      role: "platform_admin",
      user: { handle: "admin", displayName: "管理员", bio: "", location: "", createdAt: "2026-07-15T00:00:00Z" },
    });

    render(<AccountShell><div>账号内容</div></AccountShell>);

    expect(await screen.findByRole("link", { name: "邀请管理" })).toBeInTheDocument();
  });

  it("redirects an expired session to login", async () => {
    mocks.me.mockRejectedValue(new ApiError(401, {
      code: "UNAUTHENTICATED",
      message: "登录已失效",
      retryable: false,
    }));

    render(<AccountShell><div>账号内容</div></AccountShell>);

    await waitFor(() => expect(mocks.replace).toHaveBeenCalledWith("/login"));
    expect(screen.queryByText("账号内容")).not.toBeInTheDocument();
  });
});
