import type { PartialStoryFn } from '@storybook/csf';
import type { ReactRenderer } from '@storybook/react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ReactElement } from 'react';
import { CollapsibleSection } from '../CollapsibleSection';
import { StatusBadge } from '../ui/StatusBadge';
import type { Status } from '../ui/StatusConfig';

/** Stand-in for the section body — the story is about the section, not its contents. */
function Row({
  label,
  value,
  status,
  mono = false,
}: {
  label: string;
  value: string;
  status?: Status;
  mono?: boolean;
}): ReactElement {
  return (
    <div className="flex items-center justify-between py-1">
      <span className="body-small shrink-0">{label}</span>
      <span
        className={`body-small inline-flex items-center gap-1 font-medium text-text-primary${
          mono ? ' font-mono tabular-nums' : ''
        }`}
      >
        {status ? <StatusBadge status={status} size="sm" /> : null}
        {value}
      </span>
    </div>
  );
}

const meta: Meta<typeof CollapsibleSection> = {
  title: 'Components/CollapsibleSection',
  component: CollapsibleSection,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'CollapsibleSection is an accordion component for organizing content. Supports default and compact variants with optional status and count indicators.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    title: {
      control: 'text',
      description: 'Section title - can be string or React node',
    },
    defaultOpen: {
      control: 'boolean',
      description: 'Start expanded',
    },
    count: {
      control: 'number',
      description: 'Number of items to display in header',
    },
    status: {
      control: 'select',
      options: [undefined, 'success', 'warning', 'error', 'unknown', 'loading'],
      description: 'Status indicator to show next to title',
    },
    variant: {
      control: 'radio',
      options: ['default', 'compact'],
      description: 'Visual variant - default has border, compact is for inside cards',
    },
  },
  decorators: [
    (StoryComponent: PartialStoryFn<ReactRenderer>): ReactElement => (
      <div className="min-w-[400px]">
        <StoryComponent />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof CollapsibleSection>;

// Default section
export const Default: Story = {
  args: {
    title: 'Advanced Options',
    defaultOpen: false,
    children: (
      <div className="space-y-2">
        <Row label="Frame Size" value="1518 bytes" />
        <Row label="Duration" value="60 seconds" />
        <Row label="Rate Limit" value="1 Gbps" />
      </div>
    ),
  },
};

// Default open
export const DefaultOpen: Story = {
  args: {
    title: 'Test Configuration',
    defaultOpen: true,
    children: (
      <div className="space-y-2">
        <Row label="Protocol" value="TCP" />
        <Row label="Port" value="5001" />
        <Row label="Buffer Size" value="128 KB" />
      </div>
    ),
  },
};

// With count
export const WithCount: Story = {
  args: {
    title: 'Selected Tests',
    count: 4,
    defaultOpen: true,
    children: (
      <ul className="list-disc list-inside space-y-1 text-sm">
        <li>Throughput Test</li>
        <li>Latency Test</li>
        <li>Frame Loss Test</li>
        <li>Back-to-Back Test</li>
      </ul>
    ),
  },
};

// With status
export const WithStatus: Story = {
  args: {
    title: 'Test Results',
    status: 'success',
    count: 3,
    defaultOpen: true,
    children: (
      <div className="space-y-2">
        <Row label="Throughput" value="942.5 Mbps" status="success" />
        <Row label="Latency" value="1.2 ms" status="success" />
        <Row label="Frame Loss" value="0.00%" status="success" />
      </div>
    ),
  },
};

// Warning status
export const WarningStatus: Story = {
  args: {
    title: 'Performance Metrics',
    status: 'warning',
    defaultOpen: true,
    children: (
      <div className="space-y-2">
        <Row label="Throughput" value="750 Mbps" status="warning" />
        <Row label="Note" value="Below expected threshold" />
      </div>
    ),
  },
};

// Error status
export const ErrorStatus: Story = {
  args: {
    title: 'Connection Status',
    status: 'error',
    defaultOpen: true,
    children: (
      <div className="space-y-2">
        <Row label="Status" value="Disconnected" status="error" />
        <Row label="Last Seen" value="5 min ago" />
      </div>
    ),
  },
};

// Compact variant
export const CompactVariant: Story = {
  args: {
    title: 'Server Details',
    variant: 'compact',
    defaultOpen: true,
    children: (
      <div className="space-y-1">
        <Row label="IP Address" value="192.168.1.100" mono={true} />
        <Row label="Port" value="5001" mono={true} />
        <Row label="Status" value="Active" status="success" />
      </div>
    ),
  },
};

// Compact with status
export const CompactWithStatus: Story = {
  args: {
    title: 'Endpoint A',
    variant: 'compact',
    status: 'success',
    count: 2,
    defaultOpen: true,
    children: (
      <div className="space-y-1">
        <Row label="Latency" value="1.2 ms" status="success" />
        <Row label="Jitter" value="0.3 ms" status="success" />
      </div>
    ),
  },
};

// Multiple sections
export const MultipleSections: Story = {
  render: () => (
    <div className="space-y-3">
      <CollapsibleSection title="RFC 2544 Tests" count={4} defaultOpen={true}>
        <ul className="list-disc list-inside space-y-1 text-sm">
          <li>Throughput Test</li>
          <li>Latency Test</li>
          <li>Frame Loss Test</li>
          <li>Back-to-Back Test</li>
        </ul>
      </CollapsibleSection>

      <CollapsibleSection title="Y.1564 Tests" count={2} status="warning">
        <ul className="list-disc list-inside space-y-1 text-sm">
          <li>Configuration Test</li>
          <li>Performance Test</li>
        </ul>
      </CollapsibleSection>

      <CollapsibleSection title="Advanced Settings">
        <div className="space-y-2">
          <Row label="Timeout" value="30s" />
          <Row label="Retries" value="3" />
        </div>
      </CollapsibleSection>
    </div>
  ),
};

// Complex title with custom React node
export const CustomTitle: Story = {
  args: {
    title: (
      <span className="flex items-center gap-2">
        <span className="w-2 h-2 rounded-full bg-green-500" />
        <span>Active Connections</span>
      </span>
    ),
    count: 3,
    defaultOpen: true,
    children: (
      <div className="space-y-1">
        <Row label="eth0" value="Connected" status="success" />
        <Row label="eth1" value="Connected" status="success" />
        <Row label="wlan0" value="Connected" status="success" />
      </div>
    ),
  },
};
