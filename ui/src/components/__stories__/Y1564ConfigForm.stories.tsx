// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Y1564ConfigForm, defaultY1564Config, type Y1564Config } from '../Y1564ConfigForm';
import { selectedY1564Tests } from './storyData';

const meta: Meta<typeof Y1564ConfigForm> = {
  title: 'Components/Y1564ConfigForm',
  component: Y1564ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Y1564ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<Y1564Config>(defaultY1564Config);
    return <Y1564ConfigForm config={config} setConfig={setConfig} selectedTests={selectedY1564Tests} />;
  },
};
