import type { Meta, StoryObj } from '@storybook/react-vite';
import { Activity } from 'lucide-react';
import { BaseCard } from '../ui/BaseCard';
import { CardRow, CardValue } from '../ui/Card';

interface SampleData {
  name: string;
  durationMs: number;
  status: 'pass' | 'fail';
}

const sample: SampleData = {
  name: 'RFC 2544 Throughput',
  durationMs: 120000,
  status: 'pass',
};

const meta: Meta<typeof BaseCard<SampleData>> = {
  title: 'UI/BaseCard',
  component: BaseCard<SampleData>,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof BaseCard<SampleData>>;

export const WithData: Story = {
  args: {
    title: 'Test Result',
    subtitle: 'Last run',
    icon: <Activity className="w-4 h-4" />,
    data: sample,
    getStatus: (data) => (data.status === 'pass' ? 'success' : 'error'),
    children: (data) => (
      <>
        <CardValue value={data.name} size="md" />
        <CardRow label="Duration" value={`${Math.round(data.durationMs / 1000)}s`} />
      </>
    ),
  },
};

export const Loading: Story = {
  args: {
    title: 'Test Result',
    data: sample,
    loading: true,
    getStatus: () => 'loading',
    children: () => null,
  },
};

export const Error: Story = {
  args: {
    title: 'Test Result',
    data: sample,
    error: 'Failed to load results.',
    getStatus: () => 'error',
    children: () => null,
  },
};

export const Empty: Story = {
  args: {
    title: 'Test Result',
    data: null,
    getStatus: () => 'unknown',
    children: () => null,
  },
};
