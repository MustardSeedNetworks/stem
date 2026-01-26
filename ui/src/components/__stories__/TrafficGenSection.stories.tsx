// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultTrafficGenConfig, type TrafficGenConfig } from '../TrafficGenConfigForm';
import { TrafficGenSection } from '../settings/tests/TrafficGenSection';
import { selectedTrafficGenTests } from './storyData';

const meta: Meta<typeof TrafficGenSection> = {
  title: 'Settings/Tests/TrafficGenSection',
  component: TrafficGenSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TrafficGenSection>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedTrafficGenTests);
    const [config, setConfig] = useState<TrafficGenConfig>(defaultTrafficGenConfig);

    return (
      <TrafficGenSection
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
