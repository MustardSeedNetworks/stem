import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

const routes = [
  { path: '/reflector', topic: 'Reflector' },
  { path: '/tests/benchmark', topic: 'Throughput Test' },
  { path: '/tests/servicetest', topic: 'Y.1564 Service Configuration Test' },
  { path: '/tests/trafficgen', topic: 'Custom Traffic Stream' },
  { path: '/tests/measure', topic: 'Frame Delay Measurement' },
  { path: '/tests/certify', topic: 'TSN Full Validation Suite' },
  { path: '/history', topic: 'History', spanish: 'Historial' },
  { path: '/account/security', topic: 'Security', spanish: 'Seguridad' },
];

for (const language of ['en', 'es']) {
  for (const width of [390, 1440]) {
    test(`route help and keyboard restoration ${language} at ${width}px`, async ({
      page,
    }, testInfo) => {
      await page.setViewportSize({ width, height: 900 });
      await skipSetupWizard(page);
      await useRole(page, 'test_master');
      await page.addInitScript((locale) => {
        localStorage.setItem('stem-language', locale);
        localStorage.setItem('stem-theme', 'dark');
      }, language);
      for (const route of routes) {
        await page.goto(route.path);
        const opener = page.getByTestId('page-help-button');
        await opener.click();
        const drawer = page.getByTestId('help-drawer');
        await expect(drawer).toBeVisible();
        await expect(drawer).toContainText(
          language === 'es' ? (route.spanish ?? route.topic) : route.topic,
        );
        await expect(page.getByTestId('help-search')).toBeFocused();
        const bounds = await drawer.boundingBox();
        expect(bounds).not.toBeNull();
        expect(bounds?.x).toBeGreaterThanOrEqual(0);
        expect((bounds?.x ?? 0) + (bounds?.width ?? 0)).toBeLessThanOrEqual(width);
        if (route.path === '/tests/benchmark') {
          await expect(drawer.getByTestId('help-drawer-version')).not.toContainText('vv');
          await expect(drawer.getByTestId('help-standard-link')).toHaveAttribute(
            'href',
            'https://www.rfc-editor.org/rfc/rfc2544.html',
          );
          if (language === 'es')
            await expect(drawer).toContainText('Validar equipos antes de instalarlos');
          await page.screenshot({ path: testInfo.outputPath(`help-${language}-${width}.png`) });
        }
        await page.keyboard.press('Escape');
        await expect(drawer).not.toBeVisible();
        await expect(opener).toBeFocused();
        if (width === 1440) {
          const footer = page.getByTestId('sidebar-help-button');
          await footer.click();
          await expect(drawer).toContainText(
            language === 'es' ? (route.spanish ?? route.topic) : route.topic,
          );
          await page.getByTestId('help-search').press('Escape');
          await expect(footer).toBeFocused();
        }
      }
    });
  }
}

test('Spanish form help works on keyboard focus, hover and Escape', async ({ page }) => {
  await skipSetupWizard(page);
  await useRole(page, 'test_master');
  await page.addInitScript(() => localStorage.setItem('stem-language', 'es'));
  await page.goto('/tests/benchmark');
  const help = page.getByTestId('rfc2544-config-form').getByTestId('help-icon').first();
  await help.focus();
  const id = await help.getAttribute('aria-describedby');
  expect(id).toBeTruthy();
  const tooltip = page.locator(`[id="${id}"]`);
  await expect(tooltip).toBeVisible();
  await expect(tooltip).toContainText(/duración|tiempo/i);
  await page.keyboard.press('Escape');
  await expect(tooltip).not.toBeVisible();
  await page.getByTestId('page-help-button').focus();
  await help.hover();
  await expect(tooltip).toBeVisible();
  await tooltip.hover();
  await expect(tooltip).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(tooltip).not.toBeVisible();
  await expect(page.getByTestId('page-help-button')).toBeFocused();
});

test('help follows trailing-slash routes and stays closed after browser history changes', async ({
  page,
}) => {
  await skipSetupWizard(page);
  await useRole(page, 'test_master');
  await page.goto('/tests/benchmark/');
  await page.getByTestId('page-help-button').click();
  await expect(page.getByTestId('help-drawer')).toContainText('Throughput Test');
  await page.getByTestId('help-search').press('Escape');
  await page
    .getByTestId('desktop-sidebar')
    .getByRole('button', { name: 'Measure', exact: true })
    .click();
  await page.getByTestId('page-help-button').click();
  await expect(page.getByTestId('help-drawer')).toContainText('Frame Delay Measurement');
  await page.goBack();
  await expect(page).toHaveURL(/\/tests\/benchmark\/$/);
  await expect(page.getByTestId('help-drawer')).not.toBeVisible();
  await page.goForward();
  await expect(page).toHaveURL(/\/tests\/measure$/);
  await expect(page.getByTestId('help-drawer')).not.toBeVisible();
  await page.getByTestId('page-help-button').click();
  await expect(page.getByTestId('help-drawer')).toContainText('Frame Delay Measurement');
});

/**
 * The regression behind #1305: a help click that reaches React before the
 * navigation it follows has been rendered used to be swallowed. The drawer
 * opened, the route-change effect closed it again on the next commit, and
 * nothing re-opened it, so the user's click did nothing. On CI that window is
 * whatever the runner's scheduling latency happens to be — 80 ms was enough to
 * eject PR #1351 from the merge queue. Firing both clicks in one task pins the
 * ordering instead of leaving it to the machine.
 */
test('a help click batched with a navigation opens help for the destination', async ({ page }) => {
  await skipSetupWizard(page);
  await useRole(page, 'test_master');
  await page.goto('/tests/benchmark/');
  await page.getByTestId('page-help-button').waitFor();
  await page.evaluate(() => {
    const sidebar = document.querySelector('[data-testid="desktop-sidebar"]');
    const measure = [...(sidebar?.querySelectorAll('button') ?? [])].find(
      (button) => button.textContent?.trim() === 'Measure',
    );
    if (!measure) {
      throw new Error('Measure nav button not found');
    }
    measure.click();
    document.querySelector<HTMLElement>('[data-testid="page-help-button"]')?.click();
  });
  await expect(page).toHaveURL(/\/tests\/measure$/);
  await expect(page.getByTestId('help-drawer')).toContainText('Frame Delay Measurement');
});

test('embedded help build reports complete metadata', async ({ request }, testInfo) => {
  const response = await request.get('/__version');
  expect(response.ok()).toBe(true);
  const text = await response.text();
  const version: { version: string; commit: string; buildTime: string; uiBuildHash: string } =
    JSON.parse(text);
  expect(version.uiBuildHash).toMatch(/^[a-f0-9]{32}$/);
  expect(version.commit).toMatch(/^[a-f0-9]{7,40}$/);
  expect(version.version).not.toBe('unknown');
  expect(Number.isNaN(Date.parse(version.buildTime))).toBe(false);
  await testInfo.attach('build-version', { body: text, contentType: 'application/json' });
});
