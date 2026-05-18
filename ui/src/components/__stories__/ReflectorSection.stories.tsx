import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { ReflectorSection } from '../settings/ReflectorSection';
import type { ReflectorProfile } from '../settings/types';

const meta: Meta<typeof ReflectorSection> = {
  title: 'Settings/ReflectorSection',
  component: ReflectorSection,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof ReflectorSection>;

export const Default: Story = {
  render: () => {
    const [profile, setProfile] = useState<ReflectorProfile>('msn');
    return <ReflectorSection profile={profile} onProfileChange={setProfile} />;
  },
};
