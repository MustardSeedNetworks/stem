import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC2544Config, type RFC2544Config, RFC2544ConfigForm } from '../RFC2544ConfigForm';
import { selectedRFC2544Tests } from './storyData';

const meta: Meta<typeof RFC2544ConfigForm> = {
  title: 'Components/RFC2544ConfigForm',
  component: RFC2544ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC2544ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<RFC2544Config>(defaultRFC2544Config);
    return (
      <RFC2544ConfigForm
        config={config}
        setConfig={setConfig}
        selectedTests={selectedRFC2544Tests}
      />
    );
  },
};
