import type { Meta, StoryObj } from '@storybook/react-vite';
import { Skeleton } from '../ui/Skeleton';

const meta: Meta<typeof Skeleton> = {
  title: 'UI/Skeleton',
  component: Skeleton,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof Skeleton>;

export const Default: Story = {
  args: {
    width: '100%',
    height: 16,
  },
};

export const CardBlock: Story = {
  render: () => (
    <div className="space-y-2 w-72">
      <Skeleton width="70%" height={18} />
      <Skeleton width="100%" />
      <Skeleton width="90%" />
      <Skeleton width="80%" />
    </div>
  ),
};
