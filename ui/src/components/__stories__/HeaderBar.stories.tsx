import type { PartialStoryFn } from '@storybook/csf';
import type { ReactRenderer } from '@storybook/react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from '@storybook/test';
import type { ReactElement } from 'react';
import { HeaderBar } from '../HeaderBar';

const meta: Meta<typeof HeaderBar> = {
  title: 'Components/HeaderBar',
  component: HeaderBar,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'HeaderBar is the main application header with branding, status indicators, theme toggle, and action buttons.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    title: {
      control: 'text',
      description: 'Application title',
    },
    connectionStatus: {
      control: 'select',
      options: ['connected', 'disconnected', 'reconnecting'],
      description: 'SSE connection status',
    },
    theme: {
      control: 'radio',
      options: ['light', 'dark'],
      description: 'Current theme',
    },
    hasError: {
      control: 'boolean',
      description: 'Whether an error exists',
    },
    showMobileSidebar: {
      control: 'boolean',
      description: 'Whether mobile sidebar is visible',
    },
  },
  decorators: [
    (StoryComponent: PartialStoryFn<ReactRenderer>): ReactElement => (
      <div className="bg-gray-100 min-h-[200px]">
        <StoryComponent />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof HeaderBar>;

// Mock handlers using Storybook actions
const mockHandlers: Record<string, ReturnType<typeof fn>> = {
  onToggleTheme: fn(),
  onRefresh: fn(),
  onHistoryOpen: fn(),
  onHelpOpen: fn(),
  onSettingsOpen: fn(),
  onLogout: fn(),
  onReconnect: fn(),
  onToggleMobileSidebar: fn(),
};

// Default header - connected state
export const Default: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    ...mockHandlers,
  },
};

// Disconnected state
export const Disconnected: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'disconnected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    ...mockHandlers,
  },
};

// Reconnecting state
export const Reconnecting: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'reconnecting',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    ...mockHandlers,
  },
};

// Dark theme
export const DarkTheme: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'dark',
    hasError: false,
    showMobileSidebar: false,
    ...mockHandlers,
  },
  decorators: [
    (StoryComponent: PartialStoryFn<ReactRenderer>): ReactElement => (
      <div className="bg-gray-900 min-h-[200px]">
        <StoryComponent />
      </div>
    ),
  ],
};

// With error state
export const WithError: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: true,
    showMobileSidebar: false,
    ...mockHandlers,
  },
};

// With interface selector
export const WithInterfaceSelector: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    interfaces: [
      { name: 'eth0', ip: '192.168.1.100', type: 'ethernet', speed: '1 Gbps' },
      { name: 'eth1', ip: '192.168.1.101', type: 'ethernet', speed: '10 Gbps' },
      { name: 'wlan0', ip: '192.168.1.102', type: 'wifi', speed: '867 Mbps' },
    ],
    selectedInterface: 'eth0',
    onInterfaceChange: fn(),
    ...mockHandlers,
  },
};

// With profile selector
export const WithProfileSelector: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    profiles: [
      { id: '1', name: 'Default Profile', isDefault: true },
      { id: '2', name: 'RFC 2544 Suite', isDefault: false },
      { id: '3', name: 'Y.1564 Testing', isDefault: false },
    ],
    activeProfileId: '1',
    onProfileSwitch: fn(),
    onProfileManage: fn(),
    ...mockHandlers,
  },
};

// Profiles loading
export const ProfilesLoading: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    profiles: [],
    profilesLoading: true,
    ...mockHandlers,
  },
};

// Full featured header
export const FullFeatured: Story = {
  args: {
    title: 'The Stem',
    connectionStatus: 'connected',
    theme: 'light',
    hasError: false,
    showMobileSidebar: false,
    interfaces: [
      { name: 'eth0', ip: '192.168.1.100', type: 'ethernet', speed: '1 Gbps' },
      { name: 'eth1', ip: '192.168.1.101', type: 'ethernet', speed: '10 Gbps' },
    ],
    selectedInterface: 'eth0',
    onInterfaceChange: fn(),
    profiles: [
      { id: '1', name: 'Default Profile', isDefault: true },
      { id: '2', name: 'RFC 2544 Suite', isDefault: false },
    ],
    activeProfileId: '1',
    onProfileSwitch: fn(),
    onProfileManage: fn(),
    ...mockHandlers,
  },
};
