import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { TestCheckbox } from '../settings/TestCheckbox';
import type { TestDefinition } from '../settings/types';

const meta: Meta<typeof TestCheckbox> = {
  title: 'Settings/TestCheckbox',
  component: TestCheckbox,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TestCheckbox>;

const sampleTest: TestDefinition = {
  id: 'rfc2544_throughput',
  name: 'Throughput',
  desc: 'Max rate with 0% loss',
  tooltip: 'Find maximum rate at which the DUT forwards frames with zero loss.',
};

export const Default: Story = {
  render: () => {
    const [checked, setChecked] = useState(true);
    return (
      <TestCheckbox
        test={sampleTest}
        checked={checked}
        onChange={() => setChecked((prev) => !prev)}
      />
    );
  },
};
