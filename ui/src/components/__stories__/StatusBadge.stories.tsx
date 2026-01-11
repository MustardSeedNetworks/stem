// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { StatusBadge } from '../ui/StatusBadge';

const meta: Meta<typeof StatusBadge> = {
  title: 'Components/StatusBadge',
  component: StatusBadge,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'StatusBadge displays system status with visual indicators. Supports icon and dot variants with consistent color-coding.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: ['success', 'warning', 'error', 'unknown', 'loading'],
      description: 'The status to display',
    },
    variant: {
      control: 'radio',
      options: ['icon', 'dot'],
      description: 'Visual variant - icon shows symbol, dot shows small indicator',
    },
    size: {
      control: 'radio',
      options: ['sm', 'md'],
      description: 'Size of the badge',
    },
  },
};

export default meta;
type Story = StoryObj<typeof StatusBadge>;

// Default story
export const Default: Story = {
  args: {
    status: 'success',
    variant: 'icon',
    size: 'sm',
  },
};

// All status types with icon variant
export const AllStatuses: Story = {
  render: () => (
    <div className="flex gap-4 items-center">
      <div className="text-center">
        <StatusBadge status="success" variant="icon" />
        <p className="text-xs text-gray-500 mt-1">Success</p>
      </div>
      <div className="text-center">
        <StatusBadge status="warning" variant="icon" />
        <p className="text-xs text-gray-500 mt-1">Warning</p>
      </div>
      <div className="text-center">
        <StatusBadge status="error" variant="icon" />
        <p className="text-xs text-gray-500 mt-1">Error</p>
      </div>
      <div className="text-center">
        <StatusBadge status="unknown" variant="icon" />
        <p className="text-xs text-gray-500 mt-1">Unknown</p>
      </div>
      <div className="text-center">
        <StatusBadge status="loading" variant="icon" />
        <p className="text-xs text-gray-500 mt-1">Loading</p>
      </div>
    </div>
  ),
};

// Dot variant
export const DotVariant: Story = {
  render: () => (
    <div className="flex gap-4 items-center">
      <StatusBadge status="success" variant="dot" />
      <StatusBadge status="warning" variant="dot" />
      <StatusBadge status="error" variant="dot" />
      <StatusBadge status="unknown" variant="dot" />
      <StatusBadge status="loading" variant="dot" />
    </div>
  ),
};

// Size comparison
export const Sizes: Story = {
  render: () => (
    <div className="flex gap-8 items-center">
      <div>
        <p className="text-sm text-gray-500 mb-2">Small (sm)</p>
        <StatusBadge status="success" variant="icon" size="sm" />
      </div>
      <div>
        <p className="text-sm text-gray-500 mb-2">Medium (md)</p>
        <StatusBadge status="success" variant="icon" size="md" />
      </div>
    </div>
  ),
};

// In context - inline with text
export const InlineWithText: Story = {
  render: () => (
    <div className="space-y-3">
      <p className="flex items-center gap-2">
        <StatusBadge status="success" variant="icon" size="sm" />
        <span>Test passed successfully</span>
      </p>
      <p className="flex items-center gap-2">
        <StatusBadge status="warning" variant="icon" size="sm" />
        <span>Performance below threshold</span>
      </p>
      <p className="flex items-center gap-2">
        <StatusBadge status="error" variant="icon" size="sm" />
        <span>Connection failed</span>
      </p>
    </div>
  ),
};

// Success variant
export const Success: Story = {
  args: {
    status: 'success',
    variant: 'icon',
    size: 'md',
  },
};

// Warning variant
export const Warning: Story = {
  args: {
    status: 'warning',
    variant: 'icon',
    size: 'md',
  },
};

// Error state variant (renamed from Error to avoid shadowing global)
export const ErrorStatus: Story = {
  args: {
    status: 'error',
    variant: 'icon',
    size: 'md',
  },
};

// Loading variant
export const Loading: Story = {
  args: {
    status: 'loading',
    variant: 'icon',
    size: 'md',
  },
};
