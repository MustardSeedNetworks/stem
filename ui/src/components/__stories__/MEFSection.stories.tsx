// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { MEFSection } from '../settings/tests/MEFSection';
import { selectedMEFTests } from './storyData';

const meta: Meta<typeof MEFSection> = {
  title: 'Settings/Tests/MEFSection',
  component: MEFSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof MEFSection>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedMEFTests);
    return (
      <MEFSection
        selectedTests={selectedTests}
        onToggleTest={(testId) =>
          setSelectedTests((prev) =>
            prev.includes(testId) ? prev.filter((t) => t !== testId) : [...prev, testId],
          )
        }
      />
    );
  },
};
