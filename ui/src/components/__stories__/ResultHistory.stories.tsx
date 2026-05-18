import type { Meta, StoryObj } from '@storybook/react-vite';
import { ResultHistory } from '../ResultHistory';

const meta: Meta<typeof ResultHistory> = {
  title: 'Components/ResultHistory',
  component: ResultHistory,
  parameters: { layout: 'fullscreen' },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ResultHistory>;

export const Default: Story = {
  args: {
    isOpen: true,
    onClose: () => {},
    currentResult: {
      testType: 'RFC 2544 Throughput',
      module: 'Benchmark',
      status: 'completed',
      startedAt: new Date(Date.now() - 120000).toISOString(),
      completedAt: new Date().toISOString(),
      duration: 120000,
      success: true,
      metrics: { throughputMbps: 940, frameLoss: 0 },
    },
  },
};
