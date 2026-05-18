import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { TSNSection } from '../settings/tests/TSNSection';
import { defaultTSNConfig, type TSNConfig } from '../TSNConfigForm';
import { selectedTSNTests } from './storyData';

const meta: Meta<typeof TSNSection> = {
  title: 'Settings/Tests/TSNSection',
  component: TSNSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TSNSection>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedTSNTests);
    const [config, setConfig] = useState<TSNConfig>(defaultTSNConfig);

    return (
      <TSNSection
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
