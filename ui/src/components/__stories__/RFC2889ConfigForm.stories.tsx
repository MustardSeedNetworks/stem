// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { RFC2889ConfigForm, defaultRFC2889Config, type RFC2889Config } from '../RFC2889ConfigForm';
import { selectedRFC2889Tests } from './storyData';

const meta: Meta<typeof RFC2889ConfigForm> = {
  title: 'Components/RFC2889ConfigForm',
  component: RFC2889ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC2889ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<RFC2889Config>(defaultRFC2889Config);
    return <RFC2889ConfigForm config={config} setConfig={setConfig} selectedTests={selectedRFC2889Tests} />;
  },
};
