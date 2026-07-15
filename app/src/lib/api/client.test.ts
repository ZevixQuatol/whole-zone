import { afterEach, describe, expect, it, vi } from "vitest";
import { request } from "./client";

afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "hy_csrf=; max-age=0";
});

describe("request", () => {
  it("sends credentials and csrf for writes", async () => {
    document.cookie = "hy_csrf=csrf-token";
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await request("/account/profile", { method: "PATCH", body: JSON.stringify({ displayName: "林屿" }) });

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(init.credentials).toBe("include");
    expect(new Headers(init.headers).get("X-CSRF-Token")).toBe("csrf-token");
  });

  it("throws the server error contract", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ code: "INVALID", message: "输入错误", retryable: false }), {
          status: 422,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    await expect(request("/account/me")).rejects.toEqual(
      expect.objectContaining({ status: 422, message: "输入错误" }),
    );
  });
});
