/**
 * Vite Build Configuration
 *
 * Purpose: Configures the Vite development server and build process for the The Seed web frontend.
 * Handles bundling, module resolution, and development server settings.
 *
 * Configuration:
 * - React plugin: Enables JSX/TSX transformation and fast refresh during development
 * - Path alias: @ resolves to src/ directory for cleaner imports
 * - Dev server: Runs on port 3000 with HMR support
 * - Build output: Compiled to dist/ directory with source maps for debugging
 * - Embedding: Compiled frontend is embedded in Go binary via //go:embed directive
 *
 * Build Process:
 * 1. TypeScript compilation and bundling
 * 2. CSS processing and minification
 * 3. Asset optimization and tree-shaking
 * 4. Source map generation for production debugging
 * 5. Output to dist/ for Go embedding
 *
 * Usage:
 * ```bash
 * npm run dev     # Start dev server on port 3000
 * npm run build   # Build for production
 * npm run preview # Preview production build locally
 * ```
 *
 * Dependencies: vite, @vitejs/plugin-react
 * See: web/embed.go for how dist/ is embedded in the Go binary
 */

import { fileURLToPath, URL } from 'node:url';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';
import { defineConfig, loadEnv, type PluginOption } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const analyze = env.ANALYZE === 'true';

  return {
    plugins: [
      react(),
      // Bundle treemap when ANALYZE=true (`npm run build:analyze`). Parity with
      // niac; output lands in the gitignored ui/dist/ so it never ships.
      analyze &&
        (visualizer({
          open: true,
          filename: 'dist/bundle-stats.html',
          gzipSize: true,
        }) as PluginOption),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@locales': fileURLToPath(new URL('../internal/i18n/locales', import.meta.url)),
      },
      // Force a single copy of these so duplicate transitive versions don't
      // bloat the bundle or break React's single-instance invariants.
      dedupe: [
        'react',
        'react-dom',
        'react-router-dom',
        'lucide-react',
        'i18next',
        'react-i18next',
      ],
    },
    server: {
      port: 3000,
      proxy: {
        '/api': {
          // Backend serves HTTPS on :8444 by default (Wave 1 task #81).
          // For local dev with the self-signed cert we let the proxy
          // accept it; secure:false is required for self-signed.
          target: 'https://localhost:8444',
          changeOrigin: true,
          secure: false,
        },
      },
    },
    build: {
      // Output directly into the Go embed directory — no copying or syncing.
      // Canonical path shared with niac and seed: internal/api/ui/.
      outDir: '../internal/api/ui',
      // emptyOutDir intentionally omitted: outDir is outside Vite's project
      // root, so Vite defaults to false and preserves the tracked .gitkeep
      // placeholder (CLAUDE.md mandate).
      sourcemap: true,
      // Modern browser target — matches niac. ES2022 covers all evergreen
      // browsers from 2023+; we don't support IE/legacy Safari.
      target: 'es2022',
      // CSS code splitting: allow per-route CSS bundles for better caching.
      cssCodeSplit: true,
      // Module preload polyfill: not needed for evergreen browsers (ES2022 target).
      modulePreload: { polyfill: false },
      // Real budget, not a cover-up: the app shell must stay under 500 kB after
      // the vendor split below. Tighten toward niac's 350 once that holds.
      chunkSizeWarningLimit: 500,
      // Never inline assets as data: URLs (Vite default is 4096 bytes). Required
      // because @fontsource-variable ships small metric-override shim fonts that
      // would otherwise be inlined and violate the production `font-src 'self'`
      // CSP. With this set to 0, every asset bundles as a file under /assets/,
      // served from same-origin and properly HTTP-cacheable.
      assetsInlineLimit: 0,
      rollupOptions: {
        output: {
          // Split stable third-party deps into long-lived vendor chunks so an
          // app-code change doesn't bust their browser cache. Only deps stem
          // actually ships are listed (no react-query / codemirror / xyflow —
          // those are niac-only).
          manualChunks: (id: string) => {
            if (
              id.includes('/node_modules/react/') ||
              id.includes('/node_modules/react-dom/') ||
              id.includes('/node_modules/react-router-dom/') ||
              id.includes('/node_modules/scheduler/')
            )
              return 'vendor-react';
            if (
              id.includes('/node_modules/i18next/') ||
              id.includes('/node_modules/react-i18next/') ||
              id.includes('/node_modules/i18next-browser-languagedetector/')
            )
              return 'vendor-i18n';
            if (id.includes('/node_modules/zustand/') || id.includes('/node_modules/immer/'))
              return 'vendor-state';
            if (
              id.includes('/node_modules/lucide-react/') ||
              id.includes('/node_modules/tailwind-merge/')
            )
              return 'vendor-ui';
            return undefined;
          },
        },
      },
    },
  };
});
