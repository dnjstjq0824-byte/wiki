// tests/e2e/flow.spec.ts — full reader/editor/governance journey on seeded data.
import { test, expect } from '@playwright/test';

const GATEWAY_URL = process.env.TREEWIKI_GATEWAY_URL || 'http://localhost:8080';

test.describe('Treewiki End-to-End User Flow', () => {
  test('gateway health reports OK', async ({ request }) => {
    const res = await request.get(`${GATEWAY_URL}/healthz`);
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toBe('OK');
  });

  test('oauth providers endpoint lists google and github', async ({ request }) => {
    const res = await request.get(`${GATEWAY_URL}/api/v1/auth/oauth/providers`);
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    const names = (body.providers || []).map((p: any) => p.name).sort();
    expect(names).toEqual(['github', 'google']);
  });

  test('social callback page renders loading then fails cleanly without a code', async ({ page }) => {
    await page.goto('/auth/callback?provider=google');
    await expect(page.getByText('Sign-in failed')).toBeVisible();
    await expect(page.getByText('not completed')).toBeVisible();
  });

  test('home portal shows live stats and real articles', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Treewiki/i);
    await expect(page.locator('h1')).toContainText('Welcome to Treewiki');
    const stats = page.getByLabel('Platform statistics');
    await expect(stats).toContainText('Articles');
    await expect(stats).toContainText('Topic Nodes');
    await expect(page.locator('a[href="/w/Computer_Science"]').first()).toBeVisible();
  });

  test('reader journey: article, talk with live RFC tally, history', async ({ page }) => {
    // Article (page title h1 comes first; body headings follow)
    await page.goto('/w/Computer_Science');
    await expect(page.locator('h1').first()).toContainText('Computer Science', { timeout: 15000 });

    // Talk page with seeded RFC thread
    await page.goto('/talk/Treewiki');
    await expect(page.locator('h1')).toContainText('Talk:');
    await expect(page.getByText('RFC-1: Standardizing Wikitext')).toBeVisible({ timeout: 15000 });
    // First thread auto-expands; RFC tally shows 2 seeded support votes
    await expect(page.getByText('Active Community RFC')).toBeVisible({ timeout: 15000 });
    await expect(page.getByText('Support: 2')).toBeVisible();

    // Revision history
    await page.goto('/history/Treewiki');
    await expect(page.locator('h1')).toContainText('Revision History');
  });

  test('editor journey: open editor and save a new revision', async ({ page }) => {
    const title = `E2E_Article_${Date.now()}`;
    await page.goto(`/edit/${title}`);
    await expect(page.locator('h1')).toContainText(`Editing: ${title.replace(/_/g, ' ')}`);
    const editor = page.locator('textarea').first();
    await expect(editor).toBeVisible();
    await editor.fill(`= ${title.replace(/_/g, ' ')} =\n\nCreated by the Treewiki E2E suite.\n`);
    await page.getByPlaceholder(/Added definition/i).fill('E2E automated revision');
    await page.getByRole('button', { name: /Publish Revision/ }).click();
    await expect(page).toHaveURL(new RegExp(`/w/${title}`), { timeout: 15000 });
    await expect(page.locator('h1').first()).toContainText(title.replace(/_/g, ' '));
  });

  test('topic tree renders live taxonomy in all tabs', async ({ page }) => {
    await page.goto('/tree');
    await expect(page.locator('h1')).toContainText('Topic Tree Knowledge Taxonomy');
    // Directory tab lists real nodes
    await page.getByRole('button', { name: /Topic Directory/ }).click();
    await expect(page.getByText('Algorithms', { exact: false }).first()).toBeVisible({ timeout: 15000 });
    // Hierarchy tab nests the DAG
    await page.getByRole('button', { name: /Hierarchy Tree/ }).click();
    await expect(page.getByText('Computer Science').first()).toBeVisible({ timeout: 15000 });
  });

  test('search, random, recent changes, legal, admin gate, 404', async ({ page }) => {
    await page.goto('/search?q=Myers');
    await expect(page.getByText('Myers Diff').first()).toBeVisible({ timeout: 15000 });

    await page.goto('/special/random');
    await expect(page.getByText('Your random destination')).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('link', { name: /Read Article/ })).toBeVisible();

    await page.goto('/special/recent-changes');
    await expect(page.locator('h1')).toContainText('Recent Changes');

    await page.goto('/legal/terms');
    await expect(page.locator('h1')).toContainText('Terms of Service');
    await page.goto('/legal/privacy');
    await expect(page.locator('h1')).toContainText('Privacy Policy');

    // Anonymous visitors see the admin gate, not the dashboard
    await page.goto('/admin');
    await expect(page.getByText('Restricted area')).toBeVisible();

    // Unknown routes render the branded 404 page
    await page.goto('/this-route-does-not-exist-xyz');
    await expect(page.getByText('Page not found')).toBeVisible();
  });
});
