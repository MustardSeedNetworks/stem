import type { Preview } from '@storybook/react-vite';
import { RoleProvider } from '../src/contexts/RoleContext';
import '../src/index.css';

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
    (Story) => (
      <RoleProvider>
        <div className="font-sans antialiased">
          <Story />
        </div>
      </RoleProvider>
    ),
  ],

  tags: ['autodocs'],
};

export default preview;
