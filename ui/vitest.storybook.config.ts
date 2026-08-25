import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { playwright } from '@vitest/browser-playwright';
import { defineConfig } from 'vitest/config';

const currentDir = dirname(fileURLToPath(import.meta.url));

/**
 * Story files that do not yet pass the interaction/a11y run, excluded by path
 * so every other story is gated by default.
 *
 * A deny-list rather than a `tags: { include: ['test-ready'] }` allow-list, on
 * seed's evidence: under the allow-list exactly one of its 88 story files
 * carried the tag, so the job proved the harness worked while covering no real
 * component, and every story written afterwards was ungated by omission. This
 * way a new story is gated the moment it exists, and anything skipped is
 * visible here.
 *
 * Shrink this list; do not grow it.
 */
const NOT_YET_PASSING: string[] = [];

export default defineConfig({
  /* aria-query is CommonJS and reached only through @storybook/addon-a11y's
     runtime. Left un-prebundled, the browser runner imports it raw and dies
     with "does not provide an export named 'elementRoles'" before any story
     mounts — all 37 files, one cause. Forcing it through optimizeDeps applies
     the CJS interop. seed does not need this because its .storybook/main.ts
     replaces the Vite config wholesale; stem's is merged from the function
     form in vite.config.ts. */
  optimizeDeps: {
    include: ['aria-query'],
  },
  test: {
    projects: [
      {
        extends: true,
        plugins: [
          storybookTest({
            configDir: resolve(currentDir, '.storybook'),
          }),
        ],
        test: {
          name: 'storybook',
          exclude: NOT_YET_PASSING,
          /* Serial. Browser mode starts a Chromium per worker, and on a CI
             runner the contention showed up as two story files failing to
             *import* — "Vitest failed to find the current suite" — while all
             115 tests that did run passed. That is a resource race, not a
             defect in those stories, and it reproduced on no local run. Serial
             costs ~40s here and removes the whole class. */
          fileParallelism: false,
          browser: {
            enabled: true,
            headless: true,
            provider: playwright({}),
            instances: [{ browser: 'chromium' }],
          },
        },
      },
    ],
  },
});
