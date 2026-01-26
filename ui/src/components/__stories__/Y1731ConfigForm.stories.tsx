// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Y1731ConfigForm, defaultY1731Config, type Y1731Config } from '../Y1731ConfigForm';
import { selectedY1731Tests } from './storyData';

const meta: Meta<typeof Y1731ConfigForm> = {
  title: 'Components/Y1731ConfigForm',
  component: Y1731ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Y1731ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<Y1731Config>(defaultY1731Config);
    return <Y1731ConfigForm config={config} setConfig={setConfig} selectedTests={selectedY1731Tests} />;
  },
};
