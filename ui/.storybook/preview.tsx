import type { Preview } from '@storybook/react-vite';
import { I18nextProvider } from 'react-i18next';
import { RoleProvider } from '../src/contexts/RoleContext';
import i18n from '../src/i18n';
import '../src/index.css';

/* Storybook has no daemon behind it, so useBuildVersion's fetch of /__version
   404s in every story and logs a [STEM Warning] each time. Answer it with the
   canned payload the real endpoint returns. Scoped to that exact path -- every
   other request still goes through untouched, so a story that accidentally
   depends on a real API call still fails loudly rather than being silently
   served. */
const realFetch = globalThis.fetch;
globalThis.fetch = ((input: RequestInfo | URL, init?: RequestInit) => {
  const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
  if (url.endsWith('/__version')) {
    return Promise.resolve(
      new Response(
        JSON.stringify({
          version: 'storybook',
          commit: 'storybook',
          buildTime: '1970-01-01T00:00:00Z',
          uiBuildHash: 'storybook',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    );
  }
  return realFetch(input, init);
}) as typeof fetch;

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },

    a11y: {
      // 'todo' - show a11y violations in the test UI only
      // 'error' - fail CI on a11y violations
      // 'off' - skip a11y checks entirely
      test: 'todo',
    },

    backgrounds: {
      default: 'light',
      values: [
        { name: 'light', value: '#ffffff' },
        { name: 'dark', value: '#1a1a2e' },
        { name: 'surface', value: '#f8fafc' },
      ],
    },
  },

  decorators: [
    /* RoleProvider wraps every story because the shell components read the
       role, and useRole throws outside a provider rather than falling back.
       SettingsDrawer and all eight SetupWizard stories crashed on exactly
       that — invisible until the runner started rendering them, since a story
       that throws still typechecks, lints and builds. */
    /* I18nextProvider wraps every story for the same reason as RoleProvider.
       Without it react-i18next has no instance, so every component calling
       useTranslation logged NO_I18NEXT_INSTANCE and rendered raw keys -- the
       stories "passed" while showing text no user would ever see. Importing
       the app's own instance means stories render the real strings. */
    (Story) => (
      <I18nextProvider i18n={i18n}>
        <RoleProvider>
          <div className="font-sans antialiased">
            <Story />
          </div>
        </RoleProvider>
      </I18nextProvider>
    ),
  ],

  tags: ['autodocs'],
};

export default preview;
