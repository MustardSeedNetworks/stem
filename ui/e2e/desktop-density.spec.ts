import { expect, type Page, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Desktop clarity and density (UI-STEM-18, stem#1308).
 *
 * The fleet directive "clear, concise, easy — everywhere" makes two properties
 * measurable on a wide monitor, and this spec is the ratchet for both.
 *
 *   1. Content fits its container. No element scrolls sideways at 1280 or
 *      1440 px unless it declares itself a deliberate horizontal region.
 *   2. Chrome is proportionate. The page header's own footprint and the
 *      idle results strip are held at the heights the density pass left them,
 *      so a later change that loosens the page fails instead of drifting.
 *
 * The overflow walk is PER ELEMENT, never `document.scrollWidth`. A shell with
 * `overflow-hidden` clips an over-wide child, so the document measures exactly
 * the viewport width while the card is unreadable — that is how the fleet's
 * earlier "no horizontal scroll" acceptances passed on the defect
 * (support-plan UI-FLEET-2, finding 1; trellis UI-TRL-2).
 *
 * The two exemptions are authored, never inferred. CSS forces `overflow-x` to
 * `auto` whenever `overflow-y` is set, so every `overflow-y` page body computes
 * as a horizontal scroller and a style-derived exemption would make the whole
 * walk vacuous (UI-FLEET-2, finding 2).
 */

const DESKTOP_WIDTHS = [1280, 1440] as const;
const VIEWPORT_HEIGHT = 900;

/** Every authenticated route in `pageRegistry.tsx`, plus the dashboard. */
const ROUTES = [
  '/',
  '/reflector',
  '/tests/benchmark',
  '/tests/servicetest',
  '/tests/trafficgen',
  '/tests/measure',
  '/tests/certify',
  '/history',
  '/account/security',
] as const;

/**
 * Ceilings measured at 1440×900 after the UI-STEM-18 density pass, chosen so
 * that the tree before it fails them. Measured on `origin/main` at f56f7c8,
 * then on this branch:
 *
 *   page header footprint   83 / 64 / 87 px  ->  76 / 57 / 80 px
 *   idle results card      182 px            ->  57 px
 *
 * (the three header figures are a page with a description, one without, and
 * History, whose description wraps). Ratchets: tightening further is welcome,
 * loosening needs a reason.
 */
const MAX_PAGE_HEADER_FOOTPRINT_PX = 80;
const MAX_IDLE_RESULTS_STRIP_PX = 90;

interface Overflow {
  selector: string;
  scrollWidth: number;
  clientWidth: number;
}

/**
 * Elements whose content is wider than the box that holds it. Rounding in
 * fractional layout widths costs a pixel or two, so only a real overflow
 * counts.
 */
async function findHorizontalOverflow(page: Page): Promise<Overflow[]> {
  return page.evaluate(() => {
    const TOLERANCE_PX = 2;
    // A visually-hidden element is clipped to a 1 px box on purpose, so its
    // text always "overflows" — that is the technique, not a defect.
    const VISUALLY_HIDDEN_PX = 2;
    const found: { selector: string; scrollWidth: number; clientWidth: number }[] = [];

    const describe = (el: Element): string => {
      const testId = el.getAttribute('data-testid');
      if (testId) return `[data-testid="${testId}"]`;
      const cls = el.className;
      const classes = typeof cls === 'string' ? cls.split(/\s+/).slice(0, 3).join('.') : '';
      return classes ? `${el.tagName.toLowerCase()}.${classes}` : el.tagName.toLowerCase();
    };

    for (const el of Array.from(document.body.querySelectorAll('*'))) {
      // The fleet attribute from UI-FLEET-2 — a region that scrolls sideways
      // by design says so in the markup.
      if (el.closest('[data-phone-width-exempt]')) continue;

      const rect = el.getBoundingClientRect();
      if (rect.width === 0 || rect.height === 0) continue;
      if (rect.width <= VISUALLY_HIDDEN_PX || rect.height <= VISUALLY_HIDDEN_PX) continue;

      // Truncation is the prescribed fit for a cell too long for its column,
      // and a truncated cell overflows its box by construction.
      if (getComputedStyle(el).textOverflow === 'ellipsis') continue;

      if (el.scrollWidth > el.clientWidth + TOLERANCE_PX) {
        found.push({
          selector: describe(el),
          scrollWidth: el.scrollWidth,
          clientWidth: el.clientWidth,
        });
      }
    }
    return found;
  });
}

interface Density {
  /** Height of the page header block itself, excluding the shell above it. */
  headerFootprint: number;
  /** Full scroll height of the document — the before/after density figure. */
  documentHeight: number;
}

async function measureDensity(page: Page): Promise<Density> {
  return page.evaluate(() => {
    const title = document.querySelector('[data-testid="page-header-title"]');
    // The header block is the title's nearest ancestor carrying the page
    // header's own bottom margin; walking up from the title keeps the
    // measurement independent of the bar above it (UI-STEM-9 owns that).
    const block = title?.closest('.animate-fade-in') ?? null;
    return {
      headerFootprint: block ? Math.round(block.getBoundingClientRect().height) : -1,
      documentHeight: document.documentElement.scrollHeight,
    };
  });
}

test.describe('Desktop clarity and density', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  for (const width of DESKTOP_WIDTHS) {
    test(`content fits its container at ${width}px on every route`, async ({ page }) => {
      await page.setViewportSize({ width, height: VIEWPORT_HEIGHT });

      const offenders: string[] = [];
      for (const route of ROUTES) {
        await page.goto(route);
        await expect(page.getByTestId('page-header-title')).toBeVisible();

        for (const overflow of await findHorizontalOverflow(page)) {
          offenders.push(
            `${route} → ${overflow.selector} (content ${overflow.scrollWidth}px in ${overflow.clientWidth}px)`,
          );
        }
      }

      expect(
        offenders,
        `no element may scroll sideways at ${width}px without data-phone-width-exempt`,
      ).toEqual([]);
    });
  }

  test('page chrome stays within its density budget at 1440px', async ({ page }, testInfo) => {
    await page.setViewportSize({ width: 1440, height: VIEWPORT_HEIGHT });

    const measured: Record<string, Density> = {};
    const tooTall: string[] = [];

    for (const route of ROUTES) {
      await page.goto(route);
      await expect(page.getByTestId('page-header-title')).toBeVisible();

      const density = await measureDensity(page);
      measured[route] = density;

      expect(density.headerFootprint, `${route} should render a page header`).toBeGreaterThan(0);
      if (density.headerFootprint > MAX_PAGE_HEADER_FOOTPRINT_PX) {
        tooTall.push(`${route} → page header ${density.headerFootprint}px`);
      }
    }

    await testInfo.attach('density-1440.json', {
      body: JSON.stringify(measured, null, 2),
      contentType: 'application/json',
    });

    expect(
      tooTall,
      `page header must stay within ${MAX_PAGE_HEADER_FOOTPRINT_PX}px (UI-STEM-18 ratchet)`,
    ).toEqual([]);
  });

  test('the idle results strip is a strip, not an empty panel', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: VIEWPORT_HEIGHT });
    await page.goto('/tests/benchmark');

    const idleMessage = page.getByText(/No tests running/);
    await expect(idleMessage).toBeVisible();

    const height = await idleMessage.evaluate((el: HTMLElement) => {
      const card = el.closest('.card');
      return card ? Math.round(card.getBoundingClientRect().height) : -1;
    });

    expect(
      height,
      `an idle results card showing one sentence must stay within ${MAX_IDLE_RESULTS_STRIP_PX}px (UI-STEM-18 ratchet)`,
    ).toBeLessThanOrEqual(MAX_IDLE_RESULTS_STRIP_PX);
  });
});
