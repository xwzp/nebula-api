import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './auth.js';

test.describe.serial('Subscription Plan Group Admin', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('navigate to subscription page and see create button', async ({ page }) => {
    await page.goto('/console/subscription');
    await expect(page.getByRole('button', { name: '新建套餐组' })).toBeVisible({ timeout: 15000 });
  });

  test('create a new plan group via UI', async ({ page }) => {
    await page.goto('/console/subscription');
    await expect(page.getByRole('button', { name: '新建套餐组' })).toBeVisible({ timeout: 15000 });

    // Click create button
    await page.getByRole('button', { name: '新建套餐组' }).click();

    // SideSheet opens — fill the form
    // Wait for the SideSheet to be visible
    await expect(page.locator('.semi-sidesheet')).toBeVisible({ timeout: 5000 });

    // Fill "套餐组标题"
    await page.locator('.semi-sidesheet').locator('input').first().fill('E2E 测试套餐');

    // Click the "创建" button in the SideSheet footer
    await page.locator('.semi-sidesheet-footer').getByRole('button', { name: '创建' }).click();

    // Wait for success — the group should appear in the table
    await expect(page.getByText('E2E 测试套餐')).toBeVisible({ timeout: 10000 });
  });

  test('edit a plan group via UI', async ({ page }) => {
    await page.goto('/console/subscription');
    // Wait for the group we just created
    await expect(page.getByText('E2E 测试套餐')).toBeVisible({ timeout: 15000 });

    // Click "编辑" on the row
    const row = page.locator('tr', { hasText: 'E2E 测试套餐' });
    await row.getByRole('button', { name: '编辑' }).click();

    // SideSheet opens for editing
    await expect(page.locator('.semi-sidesheet')).toBeVisible({ timeout: 5000 });

    // Clear and update the title
    const titleInput = page.locator('.semi-sidesheet').locator('input').first();
    await titleInput.clear();
    await titleInput.fill('E2E 测试套餐(已修改)');

    // Click "更新"
    await page.locator('.semi-sidesheet-footer').getByRole('button', { name: '更新' }).click();

    // Verify the updated name appears
    await expect(page.getByText('E2E 测试套餐(已修改)')).toBeVisible({ timeout: 10000 });
  });

  test('disable a plan group via UI', async ({ page }) => {
    await page.goto('/console/subscription');
    await expect(page.getByText('E2E 测试套餐(已修改)')).toBeVisible({ timeout: 15000 });

    // Click "禁用" on the row
    const row = page.locator('tr', { hasText: 'E2E 测试套餐(已修改)' });
    await row.getByRole('button', { name: '禁用' }).click();

    // Confirm the modal
    await page.getByRole('button', { name: '确认' }).click();

    // Wait for the status to change — should show "启用" button now
    await expect(row.getByRole('button', { name: '启用' })).toBeVisible({ timeout: 10000 });
  });

  test('re-enable a plan group via UI', async ({ page }) => {
    await page.goto('/console/subscription');
    await expect(page.getByText('E2E 测试套餐(已修改)')).toBeVisible({ timeout: 15000 });

    const row = page.locator('tr', { hasText: 'E2E 测试套餐(已修改)' });
    await row.getByRole('button', { name: '启用' }).click();

    // Confirm
    await page.getByRole('button', { name: '确认' }).click();

    // Should show "禁用" button again
    await expect(row.getByRole('button', { name: '禁用' })).toBeVisible({ timeout: 10000 });
  });

  test('delete a plan group via UI', async ({ page }) => {
    await page.goto('/console/subscription');
    await expect(page.getByText('E2E 测试套餐(已修改)')).toBeVisible({ timeout: 15000 });

    const row = page.locator('tr', { hasText: 'E2E 测试套餐(已修改)' });
    await row.getByRole('button', { name: '删除' }).click();

    // Confirm the delete modal
    await page.getByRole('button', { name: '确认' }).click();

    // The group should disappear from the table
    await expect(page.getByText('E2E 测试套餐(已修改)')).toBeHidden({ timeout: 10000 });
  });
});
