import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC6349Config, type RFC6349Config, RFC6349ConfigForm } from '../RFC6349ConfigForm';
import { selectedRFC6349Tests } from './storyData';

const meta: Meta<typeof RFC6349ConfigForm> = {
  title: 'Components/RFC6349ConfigForm',
  component: RFC6349ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC6349ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<RFC6349Config>(defaultRFC6349Config);
    return (
      <RFC6349ConfigForm
        config={config}
        setConfig={setConfig}
        selectedTests={selectedRFC6349Tests}
      />
    );
  },
};
