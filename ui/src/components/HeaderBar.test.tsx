/**
 * HeaderBar Component Tests
 *
 * Tests the HeaderBar component for correct rendering, connection status display,
 * button interactions, interface/profile selection, and accessibility.
 */

import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { HeaderBar } from './HeaderBar';

// Default props for HeaderBar
const defaultProps: Parameters<typeof HeaderBar>[0] = {
  connectionStatus: 'connected' as const,
  darkMode: false,
  onToggleTheme: vi.fn(),
  onLogout: vi.fn(),
};

describe('HeaderBar', () => {
  describe('rendering', () => {
    it('renders app title', () => {
      render(<HeaderBar {...defaultProps} />);
      expect(screen.getByText('The Stem')).toBeInTheDocument();
    });

    it('renders the theme toggle; refresh/history/help/settings are no longer in the header (#342)', () => {
      render(<HeaderBar {...defaultProps} />);

      expect(screen.getByRole('button', { name: /switch to dark mode/i })).toBeInTheDocument();
      // These moved to the sidebar footer / page level and logout to the
      // profile menu — the header was slimmed in #342.
      expect(screen.queryByRole('button', { name: /refresh interfaces/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /open test history/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /open help/i })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /open settings/i })).not.toBeInTheDocument();
    });
  });

  describe('connection status', () => {
    it('displays connected status when connected', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="connected" />);
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });

    it('displays disconnected status when disconnected', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="disconnected" />);
      expect(screen.getByText('Disconnected')).toBeInTheDocument();
    });

    it('displays connecting status when connecting', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="connecting" />);
      // "Connecting..." appears in both the status badge and mobile view
      const connectingElements = screen.getAllByText('Connecting...');
      expect(connectingElements.length).toBeGreaterThanOrEqual(1);
    });

    it('calls onReconnect when app icon clicked while disconnected', () => {
      const onReconnect = vi.fn();
      render(
        <HeaderBar {...defaultProps} connectionStatus="disconnected" onReconnect={onReconnect} />,
      );

      const reconnectButton = screen.getByRole('button', {
        name: /click to reconnect/i,
      });
      fireEvent.click(reconnectButton);
      expect(onReconnect).toHaveBeenCalled();
    });

    it('does not show reconnect option when connected', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="connected" />);
      expect(screen.queryByRole('button', { name: /click to reconnect/i })).not.toBeInTheDocument();
    });
  });

  describe('theme toggle', () => {
    it('shows moon icon when in light mode', () => {
      render(<HeaderBar {...defaultProps} darkMode={false} />);
      expect(screen.getByRole('button', { name: /switch to dark mode/i })).toBeInTheDocument();
    });

    it('shows sun icon when in dark mode', () => {
      render(<HeaderBar {...defaultProps} darkMode={true} />);
      expect(screen.getByRole('button', { name: /switch to light mode/i })).toBeInTheDocument();
    });

    it('calls onToggleTheme when clicked', () => {
      const onToggleTheme = vi.fn();
      render(<HeaderBar {...defaultProps} onToggleTheme={onToggleTheme} />);

      fireEvent.click(screen.getByRole('button', { name: /switch to dark mode/i }));
      expect(onToggleTheme).toHaveBeenCalled();
    });
  });

  describe('button interactions', () => {
    it('calls onLogout from the profile menu', () => {
      const onLogout = vi.fn();
      render(
        <HeaderBar
          {...defaultProps}
          profiles={[{ id: '1', name: 'Default' }]}
          onProfileSwitch={vi.fn()}
          onLogout={onLogout}
        />,
      );

      // Logout moved into the profile dropdown (#342).
      fireEvent.click(screen.getByRole('button', { name: /select profile/i }));
      fireEvent.click(screen.getByRole('button', { name: /logout/i }));
      expect(onLogout).toHaveBeenCalled();
    });
  });

  describe('interface selector', () => {
    const interfaces = [
      { name: 'eth0', type: 'ethernet' as const, up: true },
      { name: 'wlan0', type: 'wifi' as const, up: true },
      { name: 'lo', type: 'loopback' as const, up: true },
    ];

    it('does not render interface selector when interfaces not provided', () => {
      render(<HeaderBar {...defaultProps} />);
      expect(screen.queryByRole('button', { name: /select interface/i })).not.toBeInTheDocument();
    });

    it('renders interface selector when interfaces provided', () => {
      render(<HeaderBar {...defaultProps} interfaces={interfaces} onInterfaceChange={vi.fn()} />);
      expect(screen.getByRole('button', { name: /select interface/i })).toBeInTheDocument();
    });

    it('opens dropdown when interface button clicked', () => {
      render(<HeaderBar {...defaultProps} interfaces={interfaces} onInterfaceChange={vi.fn()} />);

      fireEvent.click(screen.getByRole('button', { name: /select interface/i }));
      expect(screen.getByText('Network Interfaces')).toBeInTheDocument();
    });

    it('shows ethernet and wifi interfaces but not loopback', () => {
      render(<HeaderBar {...defaultProps} interfaces={interfaces} onInterfaceChange={vi.fn()} />);

      fireEvent.click(screen.getByRole('button', { name: /select interface/i }));
      expect(screen.getByText('Ethernet')).toBeInTheDocument();
      expect(screen.getByText('Wi-Fi')).toBeInTheDocument();
      // loopback interface should not be shown
      expect(screen.queryByText('lo')).not.toBeInTheDocument();
    });

    it('calls onInterfaceChange when interface selected', () => {
      const onInterfaceChange = vi.fn();
      render(
        <HeaderBar
          {...defaultProps}
          interfaces={interfaces}
          onInterfaceChange={onInterfaceChange}
        />,
      );

      fireEvent.click(screen.getByRole('button', { name: /select interface/i }));
      fireEvent.click(screen.getByText('Ethernet'));
      expect(onInterfaceChange).toHaveBeenCalledWith('eth0');
    });

    it('highlights current interface', () => {
      render(
        <HeaderBar
          {...defaultProps}
          interfaces={interfaces}
          currentInterface="eth0"
          onInterfaceChange={vi.fn()}
        />,
      );

      fireEvent.click(screen.getByRole('button', { name: /select interface/i }));

      // Find the button containing "Ethernet" that has the highlight class
      const buttons = screen.getAllByRole('button');
      const ethernetButton = buttons.find((btn) => btn.textContent?.includes('Ethernet'));
      expect(ethernetButton?.className).toContain('bg-brand-primary/10');
    });
  });

  describe('profile selector', () => {
    const profiles = [
      { id: '1', name: 'Default' },
      { id: '2', name: 'High Speed' },
    ];

    it('does not render profile selector when profiles not provided', () => {
      render(<HeaderBar {...defaultProps} />);
      expect(screen.queryByRole('button', { name: /select profile/i })).not.toBeInTheDocument();
    });

    it('renders profile selector when profiles provided', () => {
      render(<HeaderBar {...defaultProps} profiles={profiles} onProfileSwitch={vi.fn()} />);
      expect(screen.getByRole('button', { name: /select profile/i })).toBeInTheDocument();
    });

    it('opens dropdown when profile button clicked', () => {
      render(<HeaderBar {...defaultProps} profiles={profiles} onProfileSwitch={vi.fn()} />);

      fireEvent.click(screen.getByRole('button', { name: /select profile/i }));
      expect(screen.getByText('Default')).toBeInTheDocument();
      expect(screen.getByText('High Speed')).toBeInTheDocument();
    });

    it('calls onProfileSwitch when profile selected', async () => {
      const onProfileSwitch = vi.fn().mockResolvedValue(true);
      render(<HeaderBar {...defaultProps} profiles={profiles} onProfileSwitch={onProfileSwitch} />);

      fireEvent.click(screen.getByRole('button', { name: /select profile/i }));
      fireEvent.click(screen.getByText('High Speed'));

      await waitFor(() => {
        expect(onProfileSwitch).toHaveBeenCalledWith('2');
      });
    });

    it('shows manage button when onProfileManage provided', () => {
      const onProfileManage = vi.fn();
      render(
        <HeaderBar
          {...defaultProps}
          profiles={profiles}
          onProfileSwitch={vi.fn()}
          onProfileManage={onProfileManage}
        />,
      );

      fireEvent.click(screen.getByRole('button', { name: /select profile/i }));
      expect(screen.getByText('Manage')).toBeInTheDocument();
    });

    it('calls onProfileManage when manage clicked', () => {
      const onProfileManage = vi.fn();
      render(
        <HeaderBar
          {...defaultProps}
          profiles={profiles}
          onProfileSwitch={vi.fn()}
          onProfileManage={onProfileManage}
        />,
      );

      fireEvent.click(screen.getByRole('button', { name: /select profile/i }));
      fireEvent.click(screen.getByText('Manage'));
      expect(onProfileManage).toHaveBeenCalled();
    });

    it('shows loading spinner when profilesLoading is true', () => {
      render(
        <HeaderBar
          {...defaultProps}
          profiles={profiles}
          onProfileSwitch={vi.fn()}
          profilesLoading={true}
        />,
      );

      // The spinner should be present (it has animate-spin class)
      const button = screen.getByRole('button', { name: /select profile/i });
      const spinner = button.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('uses semantic header element', () => {
      const { container } = render(<HeaderBar {...defaultProps} />);
      expect(container.querySelector('header')).toBeInTheDocument();
    });

    it('all buttons have accessible labels', () => {
      render(
        <HeaderBar
          {...defaultProps}
          interfaces={[{ name: 'eth0', type: 'ethernet' as const, up: true }]}
          onInterfaceChange={vi.fn()}
          profiles={[{ id: '1', name: 'Default' }]}
          onProfileSwitch={vi.fn()}
        />,
      );

      const buttons = screen.getAllByRole('button');
      for (const button of buttons) {
        expect(button).toHaveAttribute('aria-label');
      }
    });

    it('uses h1 for app title', () => {
      render(<HeaderBar {...defaultProps} />);
      const heading = screen.getByRole('heading', { level: 1 });
      expect(heading).toHaveTextContent('The Stem');
    });
  });

  describe('mobile view', () => {
    it('shows mobile reconnect prompt when disconnected', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="disconnected" onReconnect={vi.fn()} />);

      expect(screen.getByText('Tap to reconnect')).toBeInTheDocument();
    });

    it('does not show mobile reconnect prompt when connected', () => {
      render(<HeaderBar {...defaultProps} connectionStatus="connected" />);
      expect(screen.queryByText('Tap to reconnect')).not.toBeInTheDocument();
    });
  });
});
