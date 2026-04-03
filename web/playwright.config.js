import { defineConfig } from '@playwright/test';

const E2E_PORT = 3099;

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  retries: 0,
  workers: 1, // serialize tests to avoid shared-state issues on a single server
  reporter: 'list',
  use: {
    baseURL: `http://localhost:${E2E_PORT}`,
    headless: true,
    locale: 'zh-CN',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
  // Start a dedicated Go server on a non-conflicting port with a test-only SQLite DB.
  webServer: {
    command: [
      'cd ..',
      '&&',
      'eval "$(~/.local/bin/mise activate bash)"',
      '&&',
      `SQLITE_PATH=./e2e-test.db go run main.go --port ${E2E_PORT}`,
    ].join(' '),
    port: E2E_PORT,
    timeout: 120_000,
    reuseExistingServer: false,
  },
});
