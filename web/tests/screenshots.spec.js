import { test, expect } from '@playwright/test';
import { setupMockApi } from './mocks/api.js';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const screenshotsDir = path.join(__dirname, '../../docs/images');

// Ensure screenshots directory exists before all tests
test.beforeAll(async () => {
  if (!fs.existsSync(screenshotsDir)) {
    fs.mkdirSync(screenshotsDir, { recursive: true });
  }
});

test.describe('Screenshot Generation', () => {
  test.beforeEach(async ({ page }) => {
    // Ensure consistent viewport
    await page.setViewportSize({ width: 1280, height: 800 });
  });

  test('login page', async ({ page }) => {
    // Clear any existing auth
    await page.goto('/');
    await page.evaluate(() => localStorage.clear());

    // Navigate to login
    await page.goto('/login');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount and render
    await page.waitForTimeout(1000);

    // Wait for the login form to be visible
    await page.waitForSelector('.login-card', { timeout: 10000 });

    await page.screenshot({
      path: path.join(screenshotsDir, 'login.png'),
      fullPage: true
    });

    console.log('Login screenshot saved');
  });

  test('dashboard - normal state', async ({ page }) => {
    await setupMockApi(page, 'default');

    // Set auth token before navigating
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-jwt-token-for-screenshots');
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount
    await page.waitForTimeout(1000);

    // Wait for dashboard content to load
    await page.waitForSelector('.status-cards', { timeout: 10000 });

    // Additional wait for data to render
    await page.waitForTimeout(500);

    // Scroll to bottom to ensure all content is loaded
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(300);
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);

    await page.screenshot({
      path: path.join(screenshotsDir, 'dashboard.png'),
      fullPage: true
    });

    console.log('Dashboard screenshot saved');
  });

  test('dashboard - blocked device', async ({ page }) => {
    await setupMockApi(page, 'blocked');

    // Set auth token
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-jwt-token-for-screenshots');
    });

    await page.goto('/dashboard');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount
    await page.waitForTimeout(1000);

    // Wait for dashboard content
    await page.waitForSelector('.status-cards', { timeout: 10000 });

    // Wait for blocked badge to appear
    await page.waitForTimeout(500);

    // Scroll to bottom to ensure all content is loaded
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(300);
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);

    await page.screenshot({
      path: path.join(screenshotsDir, 'dashboard-blocked.png'),
      fullPage: true
    });

    console.log('Dashboard-blocked screenshot saved');
  });

  test('devices list', async ({ page }) => {
    await setupMockApi(page, 'default');

    // Set auth token
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-jwt-token-for-screenshots');
    });

    await page.goto('/devices');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount
    await page.waitForTimeout(1000);

    // Wait for tables to load
    await page.waitForSelector('.card', { timeout: 10000 });
    await page.waitForTimeout(500);

    // Scroll to bottom to ensure all content is loaded
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(300);
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);

    await page.screenshot({
      path: path.join(screenshotsDir, 'devices-list.png'),
      fullPage: true
    });

    console.log('Devices-list screenshot saved');
  });

  test('device configuration', async ({ page }) => {
    await setupMockApi(page, 'default');

    // Set auth token
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-jwt-token-for-screenshots');
    });

    await page.goto('/devices/AA:BB:CC:DD:EE:01');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount
    await page.waitForTimeout(1000);

    // Wait for config to load
    await page.waitForSelector('.config-content', { timeout: 10000 });
    await page.waitForTimeout(500);

    // Scroll to bottom to ensure all content is loaded
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(300);
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.waitForTimeout(300);

    // Capture full page to show schedules
    await page.screenshot({
      path: path.join(screenshotsDir, 'device-config.png'),
      fullPage: true
    });

    console.log('Device-config screenshot saved');
  });

  test('device configuration - usage section', async ({ page }) => {
    await setupMockApi(page, 'default');

    // Set auth token
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-jwt-token-for-screenshots');
    });

    await page.goto('/devices/AA:BB:CC:DD:EE:01');
    await page.waitForLoadState('domcontentloaded');

    // Wait for Vue to mount
    await page.waitForTimeout(1000);

    // Wait for config and usage to load
    await page.waitForSelector('.config-content', { timeout: 10000 });
    await page.waitForTimeout(500);

    // Scroll to the current usage section and take full page screenshot
    const usageSection = page.locator('.current-usage');
    if (await usageSection.count() > 0) {
      await usageSection.scrollIntoViewIfNeeded();
      await page.waitForTimeout(300);

      // Take a full page screenshot with usage section visible
      await page.screenshot({
        path: path.join(screenshotsDir, 'device-config-usage.png'),
        fullPage: true
      });

      console.log('Device-config-usage screenshot saved');
    } else {
      console.log('Usage section not found, skipping screenshot');
    }
  });
});
