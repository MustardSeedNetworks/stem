import type { Meta, StoryObj } from '@storybook/react-vite';
import { LanguageSwitcher } from '../LanguageSwitcher';

const meta: Meta<typeof LanguageSwitcher> = {
  title: 'Components/LanguageSwitcher',
  component: LanguageSwitcher,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof LanguageSwitcher>;

export const Default: Story = {
  args: {
    showNative: true,
  },
};

export const EnglishLabels: Story = {
  args: {
    showNative: false,
  },
};
