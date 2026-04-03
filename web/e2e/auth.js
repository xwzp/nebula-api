// Shared authentication helper for E2E tests.
// On a fresh DB, the system is in "setup" mode (4-step wizard).
// We complete setup via UI, then login via UI.

const ADMIN_USERNAME = 'e2eadmin';
const ADMIN_PASSWORD = 'e2epassword123';

/**
 * Ensure the system is set up and login as admin, all through the browser.
 * @param {import('@playwright/test').Page} page
 */
export async function loginAsAdmin(page) {
  await page.goto('/');
  await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {});

  // ---- Handle Setup Wizard (fresh DB) ----
  // The setup wizard shows "系统初始化" heading
  const setupVisible = await page.getByText('系统初始化').isVisible({ timeout: 3000 }).catch(() => false);

  if (setupVisible) {
    // Step 0: Database check — click "下一步"
    await page.getByRole('button', { name: '下一步' }).click();

    // Step 1: Admin account — fill username, password, confirmPassword
    await page.locator('input[name="username"], input[id*="username"]').first().fill(ADMIN_USERNAME);
    // Semi Design password inputs use mode='password', they render as <input type="password">
    const pwdInputs = page.locator('input[type="password"]');
    await pwdInputs.nth(0).fill(ADMIN_PASSWORD);
    await pwdInputs.nth(1).fill(ADMIN_PASSWORD);
    await page.getByRole('button', { name: '下一步' }).click();

    // Step 2: Usage mode — select "对外运营模式" (value='external')
    await page.getByLabel('使用模式选择').getByText('对外运营模式').click();
    await page.getByRole('button', { name: '下一步' }).click();

    // Step 3: Complete — click "初始化系统"
    await page.getByRole('button', { name: '初始化系统' }).click();

    // Wait for redirect after initialization
    await page.waitForTimeout(2000);
    await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {});
  }

  // ---- Login ----
  // After setup or if already initialized, we may be on login page or homepage
  // Navigate to login page explicitly
  await page.goto('/login');
  await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => {});

  // Check if we're already logged in (redirected to /console)
  if (page.url().includes('/console')) {
    return; // Already logged in
  }

  // Fill the login form
  // Semi Design Form.Input with field='username' — look for the actual input element
  const usernameInput = page.locator('input').filter({ has: page.locator('[name="username"]') }).first();
  const fallbackUsernameInput = page.locator('input').first();
  const targetUsername = await usernameInput.isVisible().catch(() => false) ? usernameInput : fallbackUsernameInput;
  await targetUsername.fill(ADMIN_USERNAME);

  // Password field
  const passwordInput = page.locator('input[type="password"]').first();
  await passwordInput.fill(ADMIN_PASSWORD);

  // Click "继续" (the login button)
  await page.getByRole('button', { name: '继续' }).click();

  // Wait for redirect to /console
  await page.waitForURL(/\/console/, { timeout: 15000 });
}
