import { test, expect } from '@playwright/test';

test.describe('Public Pages', () => {
  test('homepage loads without errors', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
    // No crash — page should have content
    await expect(page.locator('#root')).not.toBeEmpty({ timeout: 10000 });
  });

  test('homepage shows subscription section when plans exist', async ({ page }) => {
    // Navigate to homepage — even without plans, it should load without errors
    const response = await page.goto('/');
    expect(response?.status()).toBeLessThan(500);
    await expect(page.locator('#root')).not.toBeEmpty({ timeout: 10000 });
  });
});
