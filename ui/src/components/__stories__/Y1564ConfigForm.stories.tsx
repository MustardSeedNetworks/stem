import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultY1564Config, type Y1564Config, Y1564ConfigForm } from '../Y1564ConfigForm';
import { selectedY1564Tests } from './storyData';

const meta: Meta<typeof Y1564ConfigForm> = {
  title: 'Components/Y1564ConfigForm',
  component: Y1564ConfigForm,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Y1564ConfigForm>;

export const Default: Story = {
  render: () => {
    const [config, setConfig] = useState<Y1564Config>(defaultY1564Config);
    return (
      <Y1564ConfigForm config={config} setConfig={setConfig} selectedTests={selectedY1564Tests} />
    );
  },
};
