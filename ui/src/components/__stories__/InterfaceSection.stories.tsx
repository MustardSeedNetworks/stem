import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { InterfaceSection } from '../settings/InterfaceSection';
import { sampleInterfaces } from './storyData';

const meta: Meta<typeof InterfaceSection> = {
  title: 'Settings/InterfaceSection',
  component: InterfaceSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof InterfaceSection>;

export const Default: Story = {
  render: () => {
    const [selectedInterface, setSelectedInterface] = useState('eth0');
    return (
      <InterfaceSection
        interfaces={sampleInterfaces}
        selectedInterface={selectedInterface}
        onInterfaceChange={setSelectedInterface}
      />
    );
  },
};
