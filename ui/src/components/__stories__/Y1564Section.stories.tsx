import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Y1564Section } from '../settings/tests/Y1564Section';
import { defaultY1564Config, type Y1564Config } from '../Y1564ConfigForm';
import { selectedY1564Tests } from './storyData';

const meta: Meta<typeof Y1564Section> = {
  title: 'Settings/Tests/Y1564Section',
  component: Y1564Section,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Y1564Section>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedY1564Tests);
    const [config, setConfig] = useState<Y1564Config>(defaultY1564Config);

    return (
      <Y1564Section
        selectedTests={selectedTests}
        onToggleTest={(testId) =>
          setSelectedTests((prev) =>
            prev.includes(testId) ? prev.filter((t) => t !== testId) : [...prev, testId],
          )
        }
        config={config}
        onConfigChange={setConfig}
      />
    );
  },
};
