// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { HelpDrawer } from '../HelpDrawer';

const meta: Meta<typeof HelpDrawer> = {
  title: 'Components/HelpDrawer',
  component: HelpDrawer,
  parameters: { layout: 'fullscreen' },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof HelpDrawer>;

export const Default: Story = {
  args: {
    isOpen: true,
    onClose: () => {},
  },
};
