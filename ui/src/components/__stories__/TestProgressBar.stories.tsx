import type { Meta, StoryObj } from '@storybook/react-vite';
import { TestProgressBar } from '../TestProgressBar';

const meta = {
  title: 'Components/TestProgressBar',
  component: TestProgressBar,
  tags: ['autodocs'],
} satisfies Meta<typeof TestProgressBar>;

export default meta;
type Story = StoryObj<typeof meta>;

const running = {
  status: 'running' as const,
  currentTest: 'RFC 2544 Throughput',
  currentStep: 2,
  stepsTotal: 3,
  phase: 'Executing rfc2544_throughput',
  elapsedSeconds: 30,
  estimatedRemainingSeconds: 90,
  steps: [
    { testType: 'rfc2544_latency', module: 'benchmark', status: 'passed' as const },
    { testType: 'rfc2544_throughput', module: 'benchmark', status: 'running' as const },
    { testType: 'y1564', module: 'servicetest', status: 'pending' as const },
  ],
};

export const Determinate: Story = { args: { progress: running } };

export const Indeterminate: Story = {
  args: {
    progress: {
      ...running,
      currentTest: 'Y.1731 Delay',
      phase: 'Executing y1731_delay',
      estimatedRemainingSeconds: null,
    },
  },
};

export const Completed: Story = {
  args: {
    progress: {
      ...running,
      status: 'completed',
      currentStep: 3,
      elapsedSeconds: 120,
      estimatedRemainingSeconds: 0,
    },
  },
};

export const Failed: Story = {
  args: {
    progress: {
      ...running,
      status: 'error',
      estimatedRemainingSeconds: null,
    },
  },
};
