import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './auth.js';

test.describe.serial('Topup Tier Admin', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('navigate to topup tiers page and see create button', async ({ page }) => {
    await page.goto('/console/topup-tiers');
    await expect(page.getByRole('button', { name: '创建档位' })).toBeVisible({ timeout: 15000 });
  });

  test('create a new topup tier via UI', async ({ page }) => {
    await page.goto('/console/topup-tiers');
    await expect(page.getByRole('button', { name: '创建档位' })).toBeVisible({ timeout: 15000 });

    // Click create button
    await page.getByRole('button', { name: '创建档位' }).click();

    // SideSheet opens — fill the form
    await expect(page.locator('.semi-sidesheet')).toBeVisible({ timeout: 5000 });

    // Fill "标题" (first text input in the SideSheet)
    const inputs = page.locator('.semi-sidesheet input[type="text"]');
    await inputs.first().fill('E2E 充值档位');

    // Fill "充值金额" — find the InputNumber for amount
    // The form has title, subtitle, tag (text inputs), then amount, discount, etc. (number inputs)
    const numberInputs = page.locator('.semi-sidesheet .semi-input-number input');
    // First number input should be "amount"
    await numberInputs.first().clear();
    await numberInputs.first().fill('50');

    // Click "创建" in the SideSheet footer
    await page.locator('.semi-sidesheet-footer').getByRole('button', { name: '创建' }).click();

    // Verify the tier appears in the table
    await expect(page.getByText('E2E 充值档位')).toBeVisible({ timeout: 10000 });
  });

  test('edit a topup tier via UI', async ({ page }) => {
    await page.goto('/console/topup-tiers');
    await expect(page.getByText('E2E 充值档位')).toBeVisible({ timeout: 15000 });

    // Click the edit icon button on the row
    const row = page.locator('tr', { hasText: 'E2E 充值档位' });
    // Edit button is an icon button (IconEdit)
    await row.locator('button').filter({ has: page.locator('.semi-icon-edit') }).click();

    // SideSheet opens for editing
    await expect(page.locator('.semi-sidesheet')).toBeVisible({ timeout: 5000 });

    // Update the title
    const titleInput = page.locator('.semi-sidesheet input[type="text"]').first();
    await titleInput.clear();
    await titleInput.fill('E2E 充值档位(已修改)');

    // Click "更新"
    await page.locator('.semi-sidesheet-footer').getByRole('button', { name: '更新' }).click();

    // Verify the updated name
    await expect(page.getByText('E2E 充值档位(已修改)')).toBeVisible({ timeout: 10000 });
  });

  test('toggle topup tier status via switch', async ({ page }) => {
    await page.goto('/console/topup-tiers');
    await expect(page.getByText('E2E 充值档位(已修改)')).toBeVisible({ timeout: 15000 });

    // The status column has an inline Switch
    const row = page.locator('tr', { hasText: 'E2E 充值档位(已修改)' });
    const toggle = row.locator('.semi-switch');
    await toggle.click();

    // Brief wait for the API call to complete
    await page.waitForTimeout(1000);

    // Click again to re-enable
    await toggle.click();
    await page.waitForTimeout(1000);
  });

  test('delete a topup tier via UI', async ({ page }) => {
    await page.goto('/console/topup-tiers');
    await expect(page.getByText('E2E 充值档位(已修改)')).toBeVisible({ timeout: 15000 });

    // Click the delete button — it uses Popconfirm
    const row = page.locator('tr', { hasText: 'E2E 充值档位(已修改)' });
    await row.getByRole('button', { name: '删除' }).click();

    // Popconfirm shows "确认删除？" — click confirm
    await page.getByRole('button', { name: '确定' }).click();

    // The tier should disappear
    await expect(page.getByText('E2E 充值档位(已修改)')).toBeHidden({ timeout: 10000 });
  });
});
