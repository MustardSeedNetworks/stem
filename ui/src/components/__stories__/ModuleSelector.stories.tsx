// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { ModuleSelector } from '../ModuleSelector';

const meta: Meta<typeof ModuleSelector> = {
  title: 'Components/ModuleSelector',
  component: ModuleSelector,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ModuleSelector>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(['rfc2544_throughput']);
    return <ModuleSelector selectedTests={selectedTests} setSelectedTests={setSelectedTests} />;
  },
};
