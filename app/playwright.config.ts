import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.E2E_APP_URL ?? "http://127.0.0.1:3100";
const appPort = new URL(baseURL).port || "3100";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  reporter: "list",
  use: {
    baseURL,
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "desktop-chrome",
      use: {
        ...devices["Desktop Chrome"],
        channel: "chrome",
        viewport: { width: 1440, height: 900 },
      },
    },
  ],
  webServer: {
    command: `pnpm dev --hostname 127.0.0.1 --port ${appPort}`,
    env: {
      API_URL: process.env.E2E_API_URL ?? "http://127.0.0.1:18080",
    },
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
