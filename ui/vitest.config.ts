// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

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
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
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
    },
  },
});
