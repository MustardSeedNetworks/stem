// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ReactElement } from 'react';
import { ErrorBoundary } from '../ErrorBoundary';

const meta: Meta<typeof ErrorBoundary> = {
  title: 'Components/ErrorBoundary',
  component: ErrorBoundary,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'ErrorBoundary catches JavaScript errors in child components and displays a fallback UI with retry and reload options.',
      },
    },
  },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ErrorBoundary>;

// Component that throws an error for testing
function BuggyComponent(): never {
  throw new Error('Test error: Something went wrong in the component!');
}

// Component that works normally
function WorkingComponent(): ReactElement {
  return (
    <div className="p-8 text-center">
      <h2 className="text-xl font-semibold text-text-primary mb-2">Everything is working!</h2>
      <p className="text-text-secondary">This component rendered successfully.</p>
    </div>
  );
}

// Default - with working content
export const Default: Story = {
  args: {
    children: <WorkingComponent />,
  },
};

// With error - shows error UI
export const WithError: Story = {
  args: {
    children: <BuggyComponent />,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shows the default error UI when an error is caught.',
      },
    },
  },
};

// With custom fallback
export const WithCustomFallback: Story = {
  args: {
    children: <BuggyComponent />,
    fallback: (
      <div className="min-h-screen flex items-center justify-center bg-surface-base">
        <div className="text-center p-8">
          <div className="text-6xl mb-4">🔧</div>
          <h2 className="text-xl font-semibold text-text-primary mb-2">Oops! Something broke</h2>
          <p className="text-text-secondary mb-4">We&apos;re working on fixing it.</p>
          <button
            type="button"
            className="px-4 py-2 bg-brand-primary text-white rounded-lg hover:bg-brand-accent"
            onClick={() => window.location.reload()}
          >
            Refresh Page
          </button>
        </div>
      </div>
    ),
  },
  parameters: {
    docs: {
      description: {
        story: 'Custom fallback UI can be provided via the fallback prop.',
      },
    },
  },
};

// With onError callback
export const WithErrorCallback: Story = {
  args: {
    children: <BuggyComponent />,
    // In real usage, this would send to Sentry or similar service
    onError: () => {
      // Error logged to external service
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'The onError callback can be used to send errors to external logging services like Sentry.',
      },
    },
  },
};
