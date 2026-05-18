import type { Meta, StoryObj } from '@storybook/react-vite';
import { LicenseSection } from '../LicenseSection';

const meta: Meta<typeof LicenseSection> = {
  title: 'Components/LicenseSection',
  component: LicenseSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof LicenseSection>;

export const Default: Story = {};
