// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultY1731Config, type Y1731Config } from '../Y1731ConfigForm';
import { Y1731Section } from '../settings/tests/Y1731Section';
import { selectedY1731Tests } from './storyData';

const meta: Meta<typeof Y1731Section> = {
  title: 'Settings/Tests/Y1731Section',
  component: Y1731Section,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Y1731Section>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedY1731Tests);
    const [config, setConfig] = useState<Y1731Config>(defaultY1731Config);

    return (
      <Y1731Section
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
