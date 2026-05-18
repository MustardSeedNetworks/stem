import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from '@storybook/test';
import type { ComponentProps, ReactElement } from 'react';
import type { ModuleConfig, ModuleStatus, ModuleTestResults } from '../ModuleCard';
import { ModuleCard } from '../ModuleCard';

const meta: Meta<typeof ModuleCard> = {
  title: 'Components/ModuleCard',
  component: ModuleCard,
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        component:
          'ModuleCard displays a test module with enable/disable toggles, test selection, and execution controls. Supports different result types for RFC 2544, Y.1564, and Y.1731 tests.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    config: {
      description: 'Module configuration including tests and settings',
    },
    status: {
      description: 'Current execution status of the module',
    },
    results: {
      description: 'Test results if available',
    },
  },
  args: {
    onToggleModule: fn(),
    onToggleAutoStart: fn(),
    onToggleTest: fn(),
    onStart: fn(),
    onStop: fn(),
    onConfigure: fn(),
  },
};

export default meta;
type Story = StoryObj<typeof ModuleCard>;

// Sample configurations for different modules
const benchmarkConfig: ModuleConfig = {
  name: 'benchmark',
  displayName: 'Benchmark',
  description: 'RFC 2544 network benchmarking tests',
  color: '#dc2626',
  standard: 'RFC 2544',
  enabled: true,
  autoStart: false,
  tests: [
    {
      id: 'throughput',
      name: 'Throughput',
      description: 'Max rate with 0% loss',
      enabled: true,
    },
    {
      id: 'latency',
      name: 'Latency',
      description: 'Round-trip time measurement',
      enabled: true,
    },
    {
      id: 'frame_loss',
      name: 'Frame Loss',
      description: 'Loss vs offered load',
      enabled: false,
    },
    {
      id: 'back_to_back',
      name: 'Back-to-Back',
      description: 'Burst capacity test',
      enabled: false,
    },
  ],
};

const serviceTestConfig: ModuleConfig = {
  name: 'servicetest',
  displayName: 'Service Test',
  description: 'ITU-T Y.1564 service activation testing',
  color: '#ea580c',
  standard: 'Y.1564',
  enabled: true,
  autoStart: true,
  tests: [
    {
      id: 'config',
      name: 'Configuration Test',
      description: 'Service config validation',
      enabled: true,
    },
    {
      id: 'perf',
      name: 'Performance Test',
      description: 'Sustained 15+ min test',
      enabled: true,
    },
    {
      id: 'full',
      name: 'Full Test',
      description: 'Both config and perf',
      enabled: false,
    },
  ],
};

const measureConfig: ModuleConfig = {
  name: 'measure',
  displayName: 'Measure',
  description: 'ITU-T Y.1731 Ethernet OAM measurements',
  color: '#2563eb',
  standard: 'Y.1731',
  enabled: true,
  autoStart: false,
  tests: [
    {
      id: 'delay',
      name: 'Delay (DMM/DMR)',
      description: 'Frame delay measurement',
      enabled: true,
    },
    {
      id: 'loss',
      name: 'Loss (LMM/LMR)',
      description: 'Frame loss measurement',
      enabled: true,
    },
    {
      id: 'slm',
      name: 'Synthetic Loss',
      description: 'SLM measurement',
      enabled: false,
    },
  ],
};

const idleStatus: ModuleStatus = {
  status: 'idle',
  currentTest: null,
};

const runningStatus: ModuleStatus = {
  status: 'running',
  currentTest: 'Throughput',
  progress: 45,
};

const completedStatus: ModuleStatus = {
  status: 'completed',
  currentTest: 'Throughput',
};

const errorStatus: ModuleStatus = {
  status: 'error',
  currentTest: 'Throughput',
  message: 'Connection timeout',
};

// Idle state
export const Idle: Story = {
  args: {
    config: benchmarkConfig,
    status: idleStatus,
    results: null,
  },
};

// Running with RFC 2544 results
export const RunningWithResults: Story = {
  args: {
    config: benchmarkConfig,
    status: runningStatus,
    results: {
      testType: 'Throughput',
      startedAt: new Date().toISOString(),
      frameSizeResults: [
        {
          frameSize: 64,
          status: 'completed',
          txPackets: 1000000,
          rxPackets: 1000000,
          lossPercent: 0,
          throughputPps: 1488095,
        },
        {
          frameSize: 128,
          status: 'completed',
          txPackets: 800000,
          rxPackets: 799920,
          lossPercent: 0.01,
          throughputPps: 844594,
        },
        {
          frameSize: 256,
          status: 'running',
          txPackets: 450000,
          rxPackets: 449800,
          lossPercent: 0.04,
          progress: 65,
        },
        { frameSize: 512, status: 'pending' },
        { frameSize: 1024, status: 'pending' },
        { frameSize: 1518, status: 'pending' },
      ],
    } as ModuleTestResults,
  },
};

// Completed with results
export const CompletedWithResults: Story = {
  args: {
    config: benchmarkConfig,
    status: completedStatus,
    results: {
      testType: 'Throughput',
      startedAt: new Date(Date.now() - 120000).toISOString(),
      completedAt: new Date().toISOString(),
      duration: 120000,
      success: true,
      frameSizeResults: [
        {
          frameSize: 64,
          status: 'completed',
          txPackets: 1000000,
          rxPackets: 1000000,
          lossPercent: 0,
          throughputPps: 1488095,
        },
        {
          frameSize: 128,
          status: 'completed',
          txPackets: 800000,
          rxPackets: 800000,
          lossPercent: 0,
          throughputPps: 844594,
        },
        {
          frameSize: 256,
          status: 'completed',
          txPackets: 600000,
          rxPackets: 600000,
          lossPercent: 0,
          throughputPps: 452898,
        },
        {
          frameSize: 512,
          status: 'completed',
          txPackets: 400000,
          rxPackets: 400000,
          lossPercent: 0,
          throughputPps: 234962,
        },
        {
          frameSize: 1024,
          status: 'completed',
          txPackets: 300000,
          rxPackets: 300000,
          lossPercent: 0,
          throughputPps: 119731,
        },
        {
          frameSize: 1518,
          status: 'completed',
          txPackets: 250000,
          rxPackets: 250000,
          lossPercent: 0,
          throughputPps: 81274,
        },
      ],
    } as ModuleTestResults,
  },
};

// Error state
export const ErrorState: Story = {
  args: {
    config: benchmarkConfig,
    status: errorStatus,
    results: {
      testType: 'Throughput',
      startedAt: new Date(Date.now() - 30000).toISOString(),
      error: 'Connection to DUT lost during test execution',
      frameSizeResults: [
        {
          frameSize: 64,
          status: 'completed',
          txPackets: 1000000,
          rxPackets: 1000000,
          lossPercent: 0,
          throughputPps: 1488095,
        },
        { frameSize: 128, status: 'error' },
      ],
    } as ModuleTestResults,
  },
};

// Disabled module
export const Disabled: Story = {
  args: {
    config: { ...benchmarkConfig, enabled: false },
    status: idleStatus,
    results: null,
  },
};

// Y.1564 Service Test with flow results
export const ServiceTestWithFlows: Story = {
  args: {
    config: serviceTestConfig,
    status: {
      status: 'running',
      currentTest: 'Configuration Test',
      progress: 60,
    },
    results: {
      testType: 'Configuration Test',
      startedAt: new Date().toISOString(),
      serviceFlowResults: [
        {
          flowId: '1',
          flowName: 'Voice (EF)',
          status: 'completed',
          cir: 100,
          cirAchieved: 100,
          frameDelay: 2.1,
          frameDelayVariation: 0.3,
          frameLoss: 0,
        },
        {
          flowId: '2',
          flowName: 'Video (AF41)',
          status: 'running',
          cir: 500,
          cirAchieved: 498,
          frameDelay: 5.2,
          frameDelayVariation: 1.1,
        },
        { flowId: '3', flowName: 'Data (BE)', status: 'pending' },
      ],
    } as ModuleTestResults,
  },
};

// Y.1731 OAM with measurement results
export const OamMeasurements: Story = {
  args: {
    config: measureConfig,
    status: { status: 'running', currentTest: 'Delay (DMM/DMR)', progress: 80 },
    results: {
      testType: 'Delay Measurement',
      startedAt: new Date().toISOString(),
      oamResults: [
        {
          measurementType: 'Delay (DMM)',
          status: 'completed',
          delayMin: 1200,
          delayAvg: 1450,
          delayMax: 2100,
          jitter: 150,
        },
        {
          measurementType: 'Loss (LMM)',
          status: 'running',
          lossNear: 0.01,
          lossFar: 0.02,
        },
        { measurementType: 'Synthetic Loss (SLM)', status: 'pending' },
      ],
    } as ModuleTestResults,
  },
};

// Module with auto-start enabled
export const WithAutoStart: Story = {
  args: {
    config: serviceTestConfig,
    status: idleStatus,
    results: null,
  },
  parameters: {
    docs: {
      description: {
        story: 'Module configured to auto-start tests when link is detected.',
      },
    },
  },
};

// Grid of multiple modules
export const ModuleGrid: Story = {
  render: (args: ComponentProps<typeof ModuleCard>): ReactElement => (
    <div className="space-y-4">
      <ModuleCard {...args} config={benchmarkConfig} />
      <ModuleCard {...args} config={serviceTestConfig} />
      <ModuleCard {...args} config={measureConfig} />
    </div>
  ),
  args: {
    status: idleStatus,
    results: null,
  },
};
