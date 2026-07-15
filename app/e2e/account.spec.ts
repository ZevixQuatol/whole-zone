import { expect, test } from "@playwright/test";

const adminLogin = process.env.E2E_ADMIN_LOGIN;
const adminPassword = process.env.E2E_ADMIN_PASSWORD;

test("Account 模块完成受邀账号的桌面端主流程", async ({ page }) => {
  test.skip(!adminLogin || !adminPassword, "需要 E2E_ADMIN_LOGIN 和 E2E_ADMIN_PASSWORD");
  test.setTimeout(90_000);

  const stamp = Date.now().toString(36);
  const email = `e2e-${stamp}@example.com`;
  const handle = `e2e_${stamp}`;
  const displayName = `端到端用户 ${stamp}`;
  const password = `account-old-${stamp}`;
  const nextPassword = `account-new-${stamp}`;

  await login(page, adminLogin!, adminPassword!);
  await page.getByRole("link", { name: "邀请管理" }).click();
  await expect(page.getByRole("heading", { name: "邀请管理" })).toBeVisible();
  await page.getByLabel("可使用次数").fill("1");
  await page.getByLabel("有效天数").fill("7");
  await page.getByRole("button", { name: "创建邀请码" }).click();
  const code = (await page.getByLabel("新邀请码").textContent())?.trim();
  expect(code).toBeTruthy();

  await page.goto(`/register?invite=${encodeURIComponent(code!)}`);
  await page.getByLabel("邮箱").fill(email);
  await page.getByLabel("用户名").fill(handle);
  await page.getByLabel("显示名称").fill(displayName);
  await page.getByLabel("密码").fill(password);
  await page.getByRole("button", { name: "创建账号" }).click();
  await expect(page).toHaveURL(/\/settings\/profile$/);

  await page.getByLabel("个人简介").fill("专注 Go、React 与分布式系统");
  await page.getByLabel("所在地区").fill("上海");
  await page.getByRole("button", { name: "保存资料" }).click();
  await expect(page.getByRole("status")).toHaveText("公开资料已更新");

  await page.goto(`/users/${handle}`);
  await expect(page.getByRole("heading", { name: displayName })).toBeVisible();
  await expect(page.getByText("专注 Go、React 与分布式系统")).toBeVisible();

  await page.goto("/admin/invites");
  await expect(page.getByText("无权执行此操作", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "创建邀请码" })).not.toBeVisible();

  await page.goto("/notifications");
  await expect(page.getByText("暂时没有系统通知。")).toBeVisible();

  await page.goto("/settings/account");
  await page.getByLabel("当前密码").fill(password);
  await page.getByLabel("新密码").fill(nextPassword);
  await page.getByRole("button", { name: "更新并重新登录" }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.getByLabel("邮箱或用户名").fill(handle);
  await page.getByLabel("密码").fill(password);
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page.getByText("账号或密码错误", { exact: true })).toBeVisible();
  await page.getByLabel("密码").fill(nextPassword);
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page).toHaveURL(/\/settings\/profile$/);

  await page.goto("/settings/account");
  await page.getByRole("button", { name: "退出全部设备" }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.goto("/forgot-password");
  await page.getByLabel("邮箱").fill(email);
  await page.getByRole("button", { name: "发送重置邮件" }).click();
  await expect(page.getByRole("status")).toHaveText("如果账号存在，重置邮件已经发送。");
});

async function login(page: import("@playwright/test").Page, loginName: string, password: string) {
  await page.goto("/login");
  await page.getByLabel("邮箱或用户名").fill(loginName);
  await page.getByLabel("密码").fill(password);
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page).toHaveURL(/\/settings\/profile$/);
}
