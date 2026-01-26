// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import {
  TrafficGenConfigForm,
  defaultTrafficGenConfig,
  type TrafficGenConfig,
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
    return <TrafficGenConfigForm config={config} setConfig={setConfig} selectedTests={selectedTrafficGenTests} />;
  },
};
