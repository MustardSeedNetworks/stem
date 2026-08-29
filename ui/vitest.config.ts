/**
 * Vitest Configuration
 *
 * Purpose: Configures the Vitest test framework and test environment for The Stem frontend.
 * Handles test discovery, environment setup, and coverage reporting.
 *
 * Configuration:
 * - Globals: Enable global test functions (describe, it, expect) without imports
 * - Environment: jsdom - Simulates browser DOM for React component testing
 * - Setup files: Loads test/setup.ts for global mocks and utilities
 * - File discovery: Matches *.test.ts and *.spec.tsx patterns (recursive)
 * - Coverage: V8 provider with multiple report formats (text, json, html, lcov)
 *
 * Usage:
 * ```bash
 * npm test              # Run all tests
 * npm run test:watch   # Run with file watching
 * npm run test:coverage  # Generate coverage reports
 * npm test -- src/App.test.tsx  # Run specific test file
 * ```
 */

import { fileURLToPath, URL } from 'node:url';
import babel from '@rolldown/plugin-babel';
import react, { reactCompilerPreset } from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// Node's own experimental webstorage global is read once per test file while
// the jsdom environment is set up, and each read prints
// "localStorage is not available because --localstorage-file was not provided".
// Nothing here uses it -- `localStorage` in a test resolves to jsdom's
// MemoryStorage on a proper http://localhost origin -- so the feature is turned
// off rather than the message silenced. Workers do not inherit the parent's
// CLI flags and worker_threads ignores execArgv for process-level options, so
// NODE_OPTIONS is the only channel that reaches them; setting it here rather
// than in the npm script keeps it working on Windows shells too.
process.env.NODE_OPTIONS = [process.env.NODE_OPTIONS, '--no-experimental-webstorage']
  .filter(Boolean)
  .join(' ');

export default defineConfig({
  plugins: [
    react(),
    // The React Compiler, matching vite.config.ts. Without it the suite
    // exercises un-compiled components while the shipped bundle is compiled —
    // so a memo the compiler subsumes looks required here and a compiler
    // regression could never fail a test.
    babel({ presets: [reactCompilerPreset()] }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@locales': fileURLToPath(new URL('./locales', import.meta.url)),
    },
    // Force module deduplication for i18next + react-i18next. Vitest's
    // jsdom environment otherwise creates separate i18next instances
    // when one is imported directly (`import i18next from 'i18next'`)
    // vs through react-i18next's chain. That breaks the i18n hook
    // tests because changeLanguage on the global instance doesn't
    // propagate to the React-context-scoped instance.
    dedupe: ['i18next', 'react-i18next', 'react', 'react-dom'],
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
    exclude: ['src/components/__stories__/**', 'node_modules/'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: ['node_modules/', 'src/test/', '**/*.d.ts', '**/*.config.*', 'dist/'],
      // Anti-regression floor (set ~2pp below current measurement).
      // Already comfortably above CLAUDE.md's 50% minimum. Current:
      // lines 91, branches 84, functions 94, stmts 91.
      thresholds: {
        lines: 88,
        branches: 80,
        functions: 92,
        statements: 88,
      },
    },
  },
});
