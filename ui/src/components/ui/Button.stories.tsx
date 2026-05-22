/**
 * Button primitive stories (Wave 5 / #236).
 *
 * Covers variant × tone × size matrix, the disabled and loading
 * states, and the leftIcon/rightIcon slots.
 */
import type { Meta, StoryObj } from '@storybook/react-vite';
import { ArrowRight, Check, Trash2 } from 'lucide-react';
import { Button } from './';

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  parameters: { layout: 'centered' },
  argTypes: {
    variant: { control: 'select', options: ['solid', 'outline', 'ghost', 'secondary'] },
    tone: { control: 'select', options: ['violet', 'red', 'green', 'blue', 'gray'] },
    size: { control: 'select', options: ['xs', 'sm', 'md', 'lg'] },
    disabled: { control: 'boolean' },
    loading: { control: 'boolean' },
  },
  args: { children: 'Button' },
};
export default meta;

type Story = StoryObj<typeof Button>;

export const Default: Story = {
  args: { variant: 'solid', tone: 'violet', size: 'md' },
};
export const Outline: Story = {
  args: { variant: 'outline', tone: 'violet', size: 'md' },
};
export const Ghost: Story = {
  args: { variant: 'ghost', tone: 'gray', size: 'md' },
};
export const Secondary: Story = {
  args: { variant: 'secondary', size: 'md' },
};

export const Destructive: Story = {
  args: {
    variant: 'solid',
    tone: 'red',
    size: 'md',
    leftIcon: <Trash2 className="w-4 h-4" />,
    children: 'Delete',
  },
};
export const Success: Story = {
  args: {
    variant: 'solid',
    tone: 'green',
    size: 'md',
    leftIcon: <Check className="w-4 h-4" />,
    children: 'Confirm',
  },
};
export const WithRightIcon: Story = {
  args: {
    variant: 'solid',
    tone: 'blue',
    size: 'md',
    rightIcon: <ArrowRight className="w-4 h-4" />,
    children: 'Next',
  },
};

export const Disabled: Story = { args: { disabled: true } };
export const Loading: Story = { args: { loading: true, children: 'Submitting…' } };

export const SizeMatrix: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button size="xs">XS</Button>
      <Button size="sm">SM</Button>
      <Button size="md">MD</Button>
      <Button size="lg">LG</Button>
    </div>
  ),
};

export const ToneMatrix: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button tone="violet">Violet</Button>
      <Button tone="red">Red</Button>
      <Button tone="green">Green</Button>
      <Button tone="blue">Blue</Button>
      <Button tone="gray">Gray</Button>
    </div>
  ),
};
