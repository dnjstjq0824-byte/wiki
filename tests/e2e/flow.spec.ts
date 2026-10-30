// tests/e2e/flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Treewiki End-to-End User Flow', () => {
  test('Complete reader, editor, topic tree and diff journey', async ({ page }) => {
    // 1. Home Portal
    await page.goto('http://localhost:3000');
    await expect(page).toHaveTitle(/Treewiki/i);
    await expect(page.locator('h1')).toContainText('Welcome to Treewiki');

    // 2. Navigate to Featured Article
    const featuredLink = page.locator('a[href="/w/Quantum_Computing"]').first();
    if (await featuredLink.isVisible()) {
      await featuredLink.click();
      await expect(page).toHaveURL(/.*\/w\/Quantum_Computing/);
      await expect(page.locator('h1')).toContainText('Quantum Computing');
    }

    // 3. Open Live Editor
    await page.goto('http://localhost:3000/edit/Quantum_Computing');
    await expect(page.locator('h1')).toContainText('Editing: Quantum Computing');
    await expect(page.locator('textarea')).toBeVisible();

    // 4. Explore Topic Tree & Knowledge Network
    await page.goto('http://localhost:3000/tree');
    await expect(page.locator('h1')).toContainText('Topic Tree Knowledge Taxonomy');

    // 5. Inspect Revision History and Diffs
    await page.goto('http://localhost:3000/history/Quantum_Computing');
    await expect(page.locator('h1')).toContainText('Revision History');
  });
});
