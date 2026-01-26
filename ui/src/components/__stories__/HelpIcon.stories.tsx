// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { HelpIcon } from '../HelpIcon';

const meta: Meta<typeof HelpIcon> = {
  title: 'Components/HelpIcon',
  component: HelpIcon,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof HelpIcon>;

export const Default: Story = {
  args: {
    tooltip: 'This is a helpful tooltip.',
  },
};

export const Clickable: Story = {
  args: {
    tooltip: 'Click to open help.',
    onClick: () => {},
    size: 'md',
  },
};
