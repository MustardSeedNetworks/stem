import type { Meta, StoryObj } from '@storybook/react-vite';
import { RecoveryForm } from '../recovery/RecoveryForm';

const meta: Meta<typeof RecoveryForm> = {
  title: 'Auth/RecoveryForm',
  component: RecoveryForm,
  parameters: { layout: 'fullscreen' },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RecoveryForm>;

export const Default: Story = {
  args: {
    onRecoveryComplete: () => {},
    onBackToLogin: () => {},
    remainingTime: 600,
    tokenFilePath: '/var/lib/stem/.recovery-token',
  },
};
