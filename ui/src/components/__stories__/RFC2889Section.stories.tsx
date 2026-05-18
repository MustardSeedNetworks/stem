import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC2889Config, type RFC2889Config } from '../RFC2889ConfigForm';
import { RFC2889Section } from '../settings/tests/RFC2889Section';
import { selectedRFC2889Tests } from './storyData';

const meta: Meta<typeof RFC2889Section> = {
  title: 'Settings/Tests/RFC2889Section',
  component: RFC2889Section,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC2889Section>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedRFC2889Tests);
    const [config, setConfig] = useState<RFC2889Config>(defaultRFC2889Config);

    return (
      <RFC2889Section
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
