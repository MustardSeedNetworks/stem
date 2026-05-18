import type { Meta, StoryObj } from '@storybook/react-vite';
import { statusConfig } from '../ui/StatusConfig';

const meta: Meta = {
  title: 'UI/StatusConfig',
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj;

export const Overview: Story = {
  render: () => (
    <div className="grid grid-cols-2 gap-3">
      {Object.entries(statusConfig).map(([key, config]) => (
        <div
          key={key}
          className="flex items-center gap-3 p-3 border border-[var(--color-surface-border)] rounded-lg bg-[var(--color-surface-base)]"
        >
          <div className={`w-6 h-6 ${config.color}`}>{config.icon}</div>
          <div>
            <div className="text-sm font-medium text-[var(--color-text-primary)]">{key}</div>
            <div className="text-xs text-[var(--color-text-muted)]">{config.label}</div>
          </div>
        </div>
      ))}
    </div>
  ),
};
