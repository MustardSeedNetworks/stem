// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { TSNConfigForm, defaultTSNConfig, type TSNConfig } from '../TSNConfigForm';
import { selectedTSNTests } from './storyData';

const meta: Meta<typeof TSNConfigForm> = {
  title: 'Components/TSNConfigForm',
  component: TSNConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TSNConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<TSNConfig>(defaultTSNConfig);
    return <TSNConfigForm config={config} setConfig={setConfig} selectedTests={selectedTSNTests} />;
  },
};
