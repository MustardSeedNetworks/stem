import type { Meta, StoryObj } from '@storybook/react-vite';
import { Activity, Gauge, Network } from 'lucide-react';
import { Card, CardDivider, CardRow, CardValue } from '../ui/Card';

const meta: Meta<typeof Card> = {
  title: 'Components/Card',
  component: Card,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'Card is the base container component for displaying information with status badges, headers, and structured content.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: ['success', 'warning', 'error', 'unknown', 'loading'],
      description: 'Status indicator shown in header',
    },
    title: {
      control: 'text',
      description: 'Card title',
    },
    subtitle: {
      control: 'text',
      description: 'Optional subtitle below title',
    },
    onClick: {
      action: 'clicked',
      description: 'Click handler - makes card interactive',
    },
  },
};

export default meta;
type Story = StoryObj<typeof Card>;

// Default card
export const Default: Story = {
  args: {
    title: 'Throughput Test',
    subtitle: 'RFC 2544',
    status: 'success',
    children: (
      <div className="space-y-2">
        <CardValue label="Result" value={942.5} unit=" Mbps" status="success" />
        <CardRow label="Duration" value="60s" />
        <CardRow label="Frame Size" value="1518 bytes" />
      </div>
    ),
  },
};

// Card with icon
export const WithIcon: Story = {
  args: {
    title: 'Network Status',
    subtitle: 'eth0',
    status: 'success',
    icon: <Network className="w-5 h-5" />,
    children: (
      <div className="space-y-2">
        <CardRow label="Speed" value="1 Gbps" />
        <CardRow label="Status" value="Connected" status="success" />
      </div>
    ),
  },
};

// Interactive card
export const Interactive: Story = {
  args: {
    title: 'Latency Test',
    subtitle: 'Click to view details',
    status: 'warning',
    icon: <Gauge className="w-5 h-5" />,
    children: (
      <div className="space-y-2">
        <CardValue label="Average" value={2.4} unit=" ms" status="warning" />
        <CardRow label="Min" value="0.8 ms" />
        <CardRow label="Max" value="12.3 ms" />
      </div>
    ),
  },
};

// Card showing error state
export const ErrorState: Story = {
  args: {
    title: 'Connection Test',
    subtitle: 'eth1',
    status: 'error',
    icon: <Network className="w-5 h-5" />,
    children: (
      <div className="space-y-2">
        <CardValue value="Connection Failed" status="error" />
        <CardRow label="Last Attempt" value="2 min ago" />
      </div>
    ),
  },
};

// Card in loading state
export const Loading: Story = {
  args: {
    title: 'Running Test',
    subtitle: 'Please wait...',
    status: 'loading',
    icon: <Activity className="w-5 h-5" />,
    children: (
      <div className="space-y-2">
        <CardRow label="Progress" value="45%" />
        <CardRow label="Elapsed" value="27s" />
      </div>
    ),
  },
};

// Card with divider
export const WithDivider: Story = {
  args: {
    title: 'Test Results',
    status: 'success',
    children: (
      <div>
        <CardValue label="Throughput" value={942.5} unit=" Mbps" size="lg" status="success" />
        <CardDivider />
        <div className="mt-3 space-y-1">
          <CardRow label="Frame Loss" value="0.00%" status="success" />
          <CardRow label="Latency" value="1.2 ms" status="success" />
          <CardRow label="Jitter" value="0.3 ms" status="success" />
        </div>
      </div>
    ),
  },
};

// CardValue sizes
export const CardValueSizes: Story = {
  render: () => (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <CardValue label="Small" value={100} unit=" Mbps" size="sm" />
      <CardValue label="Medium (default)" value={500} unit=" Mbps" size="md" />
      <CardValue label="Large" value={942.5} unit=" Mbps" size="lg" />
    </div>
  ),
};

// CardValue with status colors
export const CardValueStatus: Story = {
  render: () => (
    <div className="space-y-4 p-4 bg-gray-50 rounded-lg">
      <CardValue label="Success" value="Passed" status="success" />
      <CardValue label="Warning" value="Below threshold" status="warning" />
      <CardValue label="Error" value="Failed" status="error" />
      <CardValue label="Loading" value="Running..." status="loading" />
    </div>
  ),
};

// CardRow examples
export const CardRowExamples: Story = {
  render: () => (
    <div className="space-y-2 p-4 bg-gray-50 rounded-lg min-w-[300px]">
      <CardRow label="Interface" value="eth0" />
      <CardRow label="Status" value="Connected" status="success" />
      <CardRow label="Speed" value="1000 Mbps" mono={true} />
      <CardRow label="MAC Address" value="00:11:22:33:44:55" mono={true} />
      <CardRow
        label="Long Value"
        value="This is a very long value that might need wrapping"
        wrap={true}
      />
    </div>
  ),
};

// Multiple cards grid
export const CardGrid: Story = {
  render: () => (
    <div className="grid grid-cols-2 gap-4">
      <Card title="Throughput" subtitle="RFC 2544" status="success">
        <CardValue value={942.5} unit=" Mbps" size="lg" status="success" />
      </Card>
      <Card title="Latency" subtitle="RFC 2544" status="warning">
        <CardValue value={2.4} unit=" ms" size="lg" status="warning" />
      </Card>
      <Card title="Frame Loss" subtitle="RFC 2544" status="success">
        <CardValue value={0.0} unit="%" size="lg" status="success" />
      </Card>
      <Card title="Jitter" subtitle="RFC 2544" status="success">
        <CardValue value={0.3} unit=" ms" size="lg" status="success" />
      </Card>
    </div>
  ),
};
