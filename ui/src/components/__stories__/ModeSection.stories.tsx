// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { ModeSection } from '../settings/ModeSection';
import type { OperatingMode } from '../settings/types';

const meta: Meta<typeof ModeSection> = {
  title: 'Settings/ModeSection',
  component: ModeSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ModeSection>;

export const Default: Story = {
  render: () => {
    const [mode, setMode] = useState<OperatingMode>('test_master');
    return <ModeSection mode={mode} onModeChange={setMode} />;
  },
};
