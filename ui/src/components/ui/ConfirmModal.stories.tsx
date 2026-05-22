/**
 * ConfirmModal primitive stories (Wave 5 / #236).
 *
 * Covers the destructive (tone=red), neutral (tone=violet), and
 * info (tone=blue) variants, with a custom-icon example.
 */
import type { Meta, StoryObj } from '@storybook/react-vite';
import { AlertTriangle, Trash2 } from 'lucide-react';
import { ConfirmModal } from './';

const meta: Meta<typeof ConfirmModal> = {
  title: 'UI/ConfirmModal',
  component: ConfirmModal,
  parameters: { layout: 'fullscreen' },
  argTypes: {
    isOpen: { control: 'boolean' },
    confirmTone: { control: 'select', options: ['red', 'violet', 'blue', 'green'] },
  },
};
export default meta;

type Story = StoryObj<typeof ConfirmModal>;

const noop = () => undefined;

export const DestructiveDelete: Story = {
  args: {
    isOpen: true,
    title: 'Delete profile?',
    message: 'This permanently removes the "benchmark-prod" profile and all attached run history.',
    confirmLabel: 'Delete',
    cancelLabel: 'Cancel',
    confirmTone: 'red',
    icon: <Trash2 className="w-6 h-6" />,
    onConfirm: noop,
    onCancel: noop,
  },
};

export const Neutral: Story = {
  args: {
    isOpen: true,
    title: 'Restart reflector?',
    message: 'In-flight test sessions will be interrupted.',
    confirmLabel: 'Restart',
    confirmTone: 'violet',
    onConfirm: noop,
    onCancel: noop,
  },
};

export const Info: Story = {
  args: {
    isOpen: true,
    title: 'Apply settings?',
    message: 'The changes are saved immediately and survive restart.',
    confirmLabel: 'Apply',
    confirmTone: 'blue',
    onConfirm: noop,
    onCancel: noop,
  },
};

export const WithCustomIcon: Story = {
  args: {
    isOpen: true,
    title: 'Switch to TestMaster mode?',
    message: 'Reflector mode will be disabled. Existing reflector sessions will be dropped.',
    confirmLabel: 'Switch',
    confirmTone: 'violet',
    icon: <AlertTriangle className="w-6 h-6" />,
    onConfirm: noop,
    onCancel: noop,
  },
};

export const RichMessage: Story = {
  args: {
    isOpen: true,
    title: 'Bulk delete devices?',
    message: (
      <div className="space-y-2">
        <p>You're about to delete 18 devices:</p>
        <ul className="list-disc list-inside text-sm text-text-muted">
          <li>15 reflectors</li>
          <li>3 test masters</li>
        </ul>
        <p className="text-status-error text-sm font-medium">This cannot be undone.</p>
      </div>
    ),
    confirmLabel: 'Delete all',
    confirmTone: 'red',
    onConfirm: noop,
    onCancel: noop,
  },
};
