import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import type { StorybookConfig } from '@storybook/react-vite';
import type { UserConfig } from 'vite';

const currentDir: string = dirname(fileURLToPath(import.meta.url));

const config: StorybookConfig = {
  stories: ['../src/**/*.mdx', '../src/**/*.stories.@(js|jsx|mjs|ts|tsx)'],
  addons: [
    '@chromatic-com/storybook',
    '@storybook/addon-vitest',
    '@storybook/addon-a11y',
    '@storybook/addon-docs',
    '@storybook/addon-onboarding',
  ],
  framework: '@storybook/react-vite',
  viteFinal: (viteConfig: UserConfig): UserConfig => {
    // OVERRIDE rather than merge, matching seed. vite.config.ts is the
    // function form and carries dev-server, visualizer and dedupe settings
    // that the story runner has no use for; merging them left aria-query
    // un-prebundled, so the browser runner died on "does not provide an export
    // named 'elementRoles'" before any story mounted — all 37 files, one
    // cause. What the stories actually need is the two aliases.
    return {
      ...viteConfig,
      resolve: {
        ...viteConfig.resolve,
        alias: {
          '@': resolve(currentDir, '../src'),
          '@locales': resolve(currentDir, '../locales'),
        },
      },
      optimizeDeps: {
        ...viteConfig.optimizeDeps,
        include: [...(viteConfig.optimizeDeps?.include ?? []), 'aria-query'],
      },
    };
  },
};
export default config;
