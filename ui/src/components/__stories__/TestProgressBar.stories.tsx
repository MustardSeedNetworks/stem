import type { Meta, StoryObj } from '@storybook/react-vite';
import type { TestProgress } from '../TestProgressBar';
import { TestProgressBar } from '../TestProgressBar';

const meta: Meta<typeof TestProgressBar> = {
  title: 'Components/TestProgressBar',
  component: TestProgressBar,
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        component:
          'TestProgressBar displays test execution progress with elapsed time, ETA, and visual progress indicator. Automatically updates elapsed time while test is running.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    progress: {
      description: 'Test progress information including status, timing, and step details',
    },
  },
};

export default meta;
type Story = StoryObj<typeof TestProgressBar>;

// Helper to create progress with relative timestamps
function createProgress(overrides: Partial<TestProgress>): TestProgress {
  return {
    status: 'running',
    currentTest: 'Throughput Test',
    expectedDuration: 60,
    startedAt: Date.now() - 30000, // 30 seconds ago
    ...overrides,
  };
}

// Idle state - component returns null
export const Idle: Story = {
  args: {
    progress: {
      status: 'idle',
      currentTest: null,
      expectedDuration: 0,
      startedAt: null,
    },
  },
  parameters: {
    docs: {
      description: {
        story: 'When status is idle, the component renders nothing.',
      },
    },
  },
};

// Starting state
export const Starting: Story = {
  args: {
    progress: createProgress({
      status: 'starting',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now(),
      expectedDuration: 120,
    }),
  },
};

// Running - early progress
export const RunningEarly: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 15000, // 15 seconds ago
      expectedDuration: 120,
      currentStep: '1 of 7 frame sizes',
    }),
  },
};

// Running - mid progress
export const RunningMid: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 60000, // 60 seconds ago
      expectedDuration: 120,
      currentStep: '4 of 7 frame sizes',
    }),
  },
};

// Running - near completion
export const RunningNearEnd: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 110000, // 110 seconds ago
      expectedDuration: 120,
      currentStep: '7 of 7 frame sizes',
      progressPercent: 92,
    }),
  },
};

// Completed state
export const Completed: Story = {
  args: {
    progress: createProgress({
      status: 'completed',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 120000,
      expectedDuration: 120,
      progressPercent: 100,
    }),
  },
};

// Cancelled state
export const Cancelled: Story = {
  args: {
    progress: createProgress({
      status: 'cancelled',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 45000,
      expectedDuration: 120,
      currentStep: '3 of 7 frame sizes',
    }),
  },
};

// Error state
export const ErrorState: Story = {
  args: {
    progress: createProgress({
      status: 'error',
      currentTest: 'RFC 2544 Throughput',
      startedAt: Date.now() - 30000,
      expectedDuration: 120,
      currentStep: '2 of 7 frame sizes',
    }),
  },
};

// Y.1564 Service Test
export const Y1564Test: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'Y.1564 Configuration Test',
      startedAt: Date.now() - 180000, // 3 minutes ago
      expectedDuration: 300, // 5 minutes
      currentStep: 'Step 3: 75% CIR',
    }),
  },
};

// Long running test
export const LongRunningTest: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'Y.1564 Performance Test',
      startedAt: Date.now() - 600000, // 10 minutes ago
      expectedDuration: 900, // 15 minutes
      currentStep: 'Service validation at 100% CIR',
    }),
  },
};

// With explicit progress percentage
export const WithExplicitProgress: Story = {
  args: {
    progress: createProgress({
      status: 'running',
      currentTest: 'Custom Traffic Generation',
      startedAt: Date.now() - 30000,
      expectedDuration: 60,
      progressPercent: 75,
      currentStep: 'Generating 1Gbps traffic',
    }),
  },
  parameters: {
    docs: {
      description: {
        story: 'When progressPercent is provided, it takes precedence over time-based calculation.',
      },
    },
  },
};
