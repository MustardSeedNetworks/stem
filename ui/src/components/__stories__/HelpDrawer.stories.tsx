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
