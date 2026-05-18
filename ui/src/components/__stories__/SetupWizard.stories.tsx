import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { SetupWizard } from '../setup/SetupWizard';

/**
 * SetupWizard guides users through initial system setup.
 *
 * Features:
 * - Admin password creation with validation
 * - Generated password suggestion option
 * - Custom password entry mode
 * - Password visibility toggle
 * - Password confirmation requirement
 * - Automatic login after setup
 * - Copy to clipboard with feedback
 * - Full i18n support (English/Spanish)
 */
const meta: Meta<typeof SetupWizard> = {
  title: 'Setup/SetupWizard',
  component: SetupWizard,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'SetupWizard is a modal component that guides users through first-time setup, requiring them to set an admin password before accessing the application.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    username: {
      control: 'text',
      description: 'Admin username to display',
    },
    suggestedPassword: {
      control: 'text',
      description: 'Pre-generated password suggestion',
    },
    setupToken: {
      control: 'text',
      description: 'One-time setup token for CSRF protection',
    },
  },
};

export default meta;
type Story = StoryObj<typeof SetupWizard>;

// No-op function for story event handlers
const noop = (): void => {
  // intentionally empty
};

/**
 * Initial setup wizard state with custom password mode.
 * User will create their own password.
 */
export const CustomPasswordMode: Story = {
  args: {
    onComplete: noop,
    onLogin: async (_username: string, _password: string) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return true;
    },
    username: 'admin',
    setupToken: 'test-token-12345',
  },
};

/**
 * Setup wizard with suggested secure password option.
 * Shows generated password that user can accept or customize.
 */
export const WithSuggestedPassword: Story = {
  args: {
    onComplete: noop,
    onLogin: async (_username: string, _password: string) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return true;
    },
    username: 'admin',
    suggestedPassword: 'Xk9mP#2vL@q7Tn4w',
    setupToken: 'test-token-12345',
  },
};

/**
 * Setup submission in progress.
 * Shows loading state while password is being set on server.
 */
export const SubmittingSetup: Story = {
  args: {
    onComplete: noop,
    onLogin: (): Promise<boolean> => {
      // Simulate slow API
      return new Promise((resolve) => {
        setTimeout(() => resolve(true), 3000);
      });
    },
    username: 'admin',
    setupToken: 'test-token-12345',
  },
  parameters: {
    docs: {
      description: {
        story: 'Enter a valid password (12+ chars) and submit to see the loading state.',
      },
    },
  },
};

/**
 * Network error during setup.
 * Shows error message when API request fails.
 */
export const NetworkError: Story = {
  args: {
    onComplete: noop,
    onLogin: (): Promise<never> => Promise.reject(new Error('Network error')),
    username: 'admin',
    setupToken: 'test-token-12345',
  },
  parameters: {
    docs: {
      description: {
        story: 'This story simulates a network error. Submit the form to see the error state.',
      },
    },
  },
};

/**
 * Setup complete but login failed.
 * Shows scenario where password was set but automatic login didn't work.
 */
export const SetupCompleteLoginFailed: Story = {
  args: {
    onComplete: noop,
    onLogin: async () => false,
    username: 'admin',
    setupToken: 'test-token-12345',
  },
  parameters: {
    docs: {
      description: {
        story: 'Submit the form to see the scenario where setup completes but auto-login fails.',
      },
    },
  },
};

/**
 * Mobile viewport responsive layout.
 * Shows how the wizard adapts to smaller screens.
 */
export const MobileViewport: Story = {
  args: {
    onComplete: noop,
    onLogin: async () => true,
    username: 'admin',
    suggestedPassword: 'Xk9mP#2vL@q7Tn4w',
    setupToken: 'test-token-12345',
  },
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
};

/**
 * Tablet viewport responsive layout.
 * Shows how the wizard displays on tablet-sized screens.
 */
export const TabletViewport: Story = {
  args: {
    onComplete: noop,
    onLogin: async () => true,
    username: 'admin',
    suggestedPassword: 'Xk9mP#2vL@q7Tn4w',
    setupToken: 'test-token-12345',
  },
  parameters: {
    viewport: {
      defaultViewport: 'tablet',
    },
  },
};

/**
 * Interactive example: complete setup flow.
 * Demonstrates full user journey from password entry to completion.
 */
export const InteractiveSetupFlow: Story = {
  render: () => {
    const [setupComplete, setSetupComplete] = useState(false);

    if (setupComplete) {
      return (
        <div className="min-h-screen bg-[var(--color-surface-base)] flex items-center justify-center">
          <div className="text-center">
            <div className="mb-4 w-16 h-16 mx-auto bg-[var(--color-status-success)]/20 rounded-full flex items-center justify-center">
              <svg
                className="w-8 h-8 text-[var(--color-status-success)]"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-[var(--color-text-primary)] mb-2">
              Setup Complete!
            </h2>
            <p className="text-sm text-[var(--color-text-muted)]">You are now logged in.</p>
          </div>
        </div>
      );
    }

    return (
      <SetupWizard
        onComplete={() => setSetupComplete(true)}
        onLogin={async (_username: string, _password: string): Promise<boolean> => {
          // Simulate API delay
          await new Promise((resolve) => setTimeout(resolve, 1000));
          return true;
        }}
        username="admin"
        suggestedPassword="Xk9mP#2vL@q7Tn4w"
        setupToken="test-token-12345"
      />
    );
  },
  parameters: {
    docs: {
      description: {
        story: 'Fill out the form and submit to see the complete setup flow with success state.',
      },
    },
  },
};
