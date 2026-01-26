// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC6349Config, type RFC6349Config } from '../RFC6349ConfigForm';
import { RFC6349Section } from '../settings/tests/RFC6349Section';
import { selectedRFC6349Tests } from './storyData';

const meta: Meta<typeof RFC6349Section> = {
  title: 'Settings/Tests/RFC6349Section',
  component: RFC6349Section,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC6349Section>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedRFC6349Tests);
    const [config, setConfig] = useState<RFC6349Config>(defaultRFC6349Config);

    return (
      <RFC6349Section
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
