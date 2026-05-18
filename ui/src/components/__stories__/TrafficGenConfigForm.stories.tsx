import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import {
  defaultTrafficGenConfig,
  type TrafficGenConfig,
  TrafficGenConfigForm,
} from '../TrafficGenConfigForm';
import { selectedTrafficGenTests } from './storyData';

const meta: Meta<typeof TrafficGenConfigForm> = {
  title: 'Components/TrafficGenConfigForm',
  component: TrafficGenConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TrafficGenConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<TrafficGenConfig>(defaultTrafficGenConfig);
    return (
      <TrafficGenConfigForm
        config={config}
        setConfig={setConfig}
        selectedTests={selectedTrafficGenTests}
      />
    );
  },
};
