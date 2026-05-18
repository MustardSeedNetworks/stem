import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC2544Config, type RFC2544Config } from '../RFC2544ConfigForm';
import { RFC2544Section } from '../settings/tests/RFC2544Section';
import { selectedRFC2544Tests } from './storyData';

const meta: Meta<typeof RFC2544Section> = {
  title: 'Settings/Tests/RFC2544Section',
  component: RFC2544Section,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof RFC2544Section>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>(selectedRFC2544Tests);
    const [config, setConfig] = useState<RFC2544Config>(defaultRFC2544Config);

    return (
      <RFC2544Section
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
