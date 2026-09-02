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
    include: [
      'aria-query',
      /* Reached only from inside stories, so Vite's initial scan of the entry
         graph does not see them. It discovers them once the first story that
         uses one mounts, re-optimizes, and tells the browser to reload:

           [vite] (client) dependencies optimized: @tanstack/react-query,
                           zustand, zustand/middleware
           [vite] (client) optimized dependencies changed. reloading

         Locally the runner survives that reload. On CI it did not: the run
         went silent at exactly that line and sat there until the job hit its
         25-minute timeout (#955). Because a timed-out job reports as
         `cancelled` rather than `failed`, and release-please is gated on
         `conclusion == 'success'`, that silently skipped the v0.24.49 release
         -- no red check anywhere, just a tag that never appeared.

         Pre-bundling them means there is nothing left to discover, so the
         mid-run reload cannot happen. Same reasoning as aria-query above. */
      '@tanstack/react-query',
      'zustand',
      'zustand/middleware',
    ],
  },
  server: {
    /* No watcher. This is a one-shot run, so nothing should be recompiling
       mid-suite -- but the watcher fired anyway on CI and invalidated the
       module graph underneath a running test:

         [vitest] Vite unexpectedly reloaded a test. This may cause tests to
                  fail, lead to flaky behaviour or duplicated test runs.
         Failed to import .../addon-vitest/dist/vitest-plugin/setup-file.js
         Caused by: Vitest failed to find the runner.

       One file (Button.stories.tsx) died while the other 37 passed, and it
       cleared on retry -- the signature of a race, not a broken story. It has
       never reproduced locally, so this targets the reported cause rather than
       a diagnosis: with no watcher there is no mid-run invalidation to lose the
       runner to. `fileParallelism: false` below already removed a separate
       contention class; this is not that one. */
    watch: null,
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
