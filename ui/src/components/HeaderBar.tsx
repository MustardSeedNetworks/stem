// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * HeaderBar Component
 *
 * Primary application header with clean icon-based toolbar.
 * Displays app branding, connection status, interface/profile selectors, and utility buttons.
 *
 * Key Features:
 * - App logo/title with status indicator
 * - Connection status badge (connected/disconnected)
 * - Interface selector dropdown (ethernet/wifi)
 * - Profile selector dropdown (optional)
 * - Theme toggle (dark/light mode)
 * - Refresh, History, Help, Settings, Logout buttons
 * - Responsive design with mobile considerations
 * - Fully accessible with ARIA labels and keyboard navigation
 * - Uses theme tokens for consistent styling
 */

import { memo, type ReactElement, useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, icon as iconTokens, layout, radius, spacing } from '../styles/theme';

// =============================================================================
// Types
// =============================================================================

type ConnectionStatus = 'connected' | 'connecting' | 'disconnected' | 'error';

interface NetworkInterface {
  name: string;
  type: 'ethernet' | 'wifi' | 'loopback' | 'unknown';
  mac?: string;
  up: boolean;
}

interface Profile {
  id: string;
  name: string;
}

interface HeaderBarProps {
  connectionStatus: ConnectionStatus;
  darkMode: boolean;
  onReconnect?: () => void;
  onToggleTheme: () => void;
  onRefresh: () => void;
  onHistoryOpen: () => void;
  onHelpOpen: () => void;
  onSettingsOpen: () => void;
  onLogout: () => void;
  interfaces?: NetworkInterface[];
  currentInterface?: string;
  onInterfaceChange?: (interfaceName: string) => void;
  profiles?: Profile[];
  activeProfile?: Profile | null;
  onProfileSwitch?: (profileId: string) => Promise<boolean>;
  onProfileManage?: () => void;
  profilesLoading?: boolean;
}

// =============================================================================
// Icons
// =============================================================================

function ActivityIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <polyline
        points="22 12 18 12 15 21 9 3 6 12 2 12"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function WifiIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"
      />
    </svg>
  );
}

function WifiOffIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <line
        x1="1"
        y1="1"
        x2="23"
        y2="23"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M16.72 11.06A10.94 10.94 0 0119 12.55M5 12.55a10.94 10.94 0 015.17-2.39M10.71 5.05A16 16 0 0122.58 9M1.42 9a15.91 15.91 0 014.7-2.88M8.53 16.11a6 6 0 016.95 0M12 20h.01"
      />
    </svg>
  );
}

function SunIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg className={className} fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
      <path
        fillRule="evenodd"
        d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function MoonIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg className={className} fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
      <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
    </svg>
  );
}

function RefreshIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
  );
}

function HistoryIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
}

function HelpIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
}

function SettingsIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
      />
    </svg>
  );
}

function LogoutIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
      />
    </svg>
  );
}

function UserIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
      />
    </svg>
  );
}

function EthernetIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={className}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M8 4h8v4H8zM6 8h12v10a2 2 0 01-2 2H8a2 2 0 01-2-2V8zM9 12v4M12 12v4M15 12v4"
      />
    </svg>
  );
}

function CheckIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg className={className} fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
    </svg>
  );
}

function SpinnerIcon({ className }: { className?: string }): ReactElement {
  return (
    <svg
      className={cn(className, 'animate-spin')}
      fill="none"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  );
}

// =============================================================================
// Helpers
// =============================================================================

function getFriendlyInterfaceName(name: string, type: string): string {
  if (type === 'wifi') {
    const match = /\d+/.exec(name);
    if (match && Number.parseInt(match[0], 10) > 0) {
      return `Wi-Fi ${Number.parseInt(match[0], 10) + 1}`;
    }
    return 'Wi-Fi';
  }
  const numMatch = /(\d+)$/.exec(name);
  if (numMatch) {
    const num = Number.parseInt(numMatch[1], 10);
    if (num > 0) {
      return `Ethernet ${num + 1}`;
    }
  }
  return 'Ethernet';
}

const iconButtonClass: string = cn(
  radius.md,
  spacing.pad.sm,
  'hover:bg-surface-hover active:bg-surface-hover',
  'focus:outline-none focus:ring-2 focus:ring-brand-primary focus:ring-offset-1 focus:ring-offset-surface-raised',
  'touch-manipulation text-text-secondary hover:text-text-primary transition-colors',
);

// =============================================================================
// Sub-components
// =============================================================================

interface ProfileDropdownProps {
  profiles: Profile[];
  activeProfile: Profile | null | undefined;
  loading: boolean;
  onSelect: (id: string) => void;
  onManage?: () => void;
}

function ProfileDropdown({
  profiles,
  activeProfile,
  loading,
  onSelect,
  onManage,
}: ProfileDropdownProps): ReactElement {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent): void => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return (): void => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (id: string): void => {
    onSelect(id);
    setIsOpen(false);
  };

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        className={iconButtonClass}
        onClick={(): void => setIsOpen(!isOpen)}
        aria-label={t('accessibility.selectProfile', 'Select profile')}
        title={
          activeProfile
            ? `${t('profile.current', 'Profile')}: ${activeProfile.name}`
            : t('profile.select', 'Select Profile')
        }
      >
        {loading ? (
          <SpinnerIcon className={iconTokens.size.md} />
        ) : (
          <UserIcon className={iconTokens.size.md} />
        )}
      </button>
      {isOpen ? (
        <div
          className={cn(
            'absolute top-full right-0 mt-1 w-56',
            radius.lg,
            'border border-surface-border bg-surface-raised shadow-lg z-50 overflow-hidden',
          )}
        >
          <div className="max-h-60 overflow-y-auto">
            {profiles.length === 0 ? (
              <div className={cn(spacing.pad.default, 'text-center')}>
                <span className="caption text-text-muted">
                  {t('profile.noProfiles', 'No profiles')}
                </span>
              </div>
            ) : (
              profiles.map((p) => (
                <button
                  type="button"
                  key={p.id}
                  onClick={(): void => handleSelect(p.id)}
                  className={cn(
                    'w-full text-left',
                    spacing.pad.sm,
                    'hover:bg-surface-hover focus:bg-surface-hover focus:outline-none',
                    p.id === activeProfile?.id && 'bg-brand-primary/10',
                  )}
                >
                  <div className="flex items-center justify-between">
                    <span className="body-small text-text-primary truncate">{p.name}</span>
                    {p.id === activeProfile?.id ? (
                      <CheckIcon className={cn(iconTokens.size.sm, 'text-brand-primary')} />
                    ) : null}
                  </div>
                </button>
              ))
            )}
          </div>
          {onManage ? (
            <div className="border-t border-surface-border">
              <button
                type="button"
                onClick={(): void => {
                  setIsOpen(false);
                  onManage();
                }}
                className={cn(
                  'w-full flex items-center justify-center',
                  spacing.gap.tight,
                  spacing.pad.sm,
                  'hover:bg-surface-hover text-brand-primary',
                )}
              >
                <SettingsIcon className={iconTokens.size.sm} />
                <span className="body-small font-medium">{t('profile.manage', 'Manage')}</span>
              </button>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

interface InterfaceDropdownProps {
  interfaces: NetworkInterface[];
  currentInterface: string | undefined;
  onSelect: (name: string) => void;
}

function InterfaceDropdown({
  interfaces,
  currentInterface,
  onSelect,
}: InterfaceDropdownProps): ReactElement {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent): void => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return (): void => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (name: string): void => {
    onSelect(name);
    setIsOpen(false);
  };

  const filtered = interfaces.filter((i): boolean => i.type !== 'loopback');

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        className={cn(
          iconButtonClass,
          currentInterface && 'ring-2 ring-brand-primary ring-offset-1 ring-offset-surface-raised',
        )}
        onClick={(): void => setIsOpen(!isOpen)}
        aria-label={t('accessibility.selectInterface', 'Select interface')}
        title={currentInterface || t('interface.select', 'Select Interface')}
      >
        <EthernetIcon className={iconTokens.size.md} />
      </button>
      {isOpen ? (
        <div
          className={cn(
            'absolute top-full right-0 mt-1 w-64',
            radius.lg,
            'border border-surface-border bg-surface-raised shadow-lg z-50 overflow-hidden',
          )}
        >
          <div className={cn(spacing.pad.sm, 'border-b border-surface-border bg-surface-base')}>
            <span className="caption font-medium text-text-muted uppercase tracking-wide">
              {t('interface.networkInterfaces', 'Network Interfaces')}
            </span>
          </div>
          <div className="max-h-60 overflow-y-auto">
            {filtered.length === 0 ? (
              <div className={cn(spacing.pad.default, 'text-center')}>
                <span className="caption text-text-muted">
                  {t('interface.noInterfaces', 'No interfaces found')}
                </span>
              </div>
            ) : (
              filtered.map((iface) => (
                <button
                  type="button"
                  key={iface.name}
                  onClick={(): void => handleSelect(iface.name)}
                  className={cn(
                    'w-full text-left',
                    spacing.pad.sm,
                    'hover:bg-surface-hover focus:bg-surface-hover focus:outline-none',
                    iface.name === currentInterface && 'bg-brand-primary/10',
                  )}
                >
                  <div className="flex items-center justify-between">
                    <div className="stack-xs">
                      <div className="flex items-center gap-1">
                        <span className="body-small text-text-primary font-medium">
                          {getFriendlyInterfaceName(iface.name, iface.type)}
                        </span>
                        {iface.type === 'wifi' && (
                          <WifiIcon className={cn(iconTokens.size.xs, 'text-text-muted')} />
                        )}
                      </div>
                      <span
                        className={cn(
                          'caption text-text-muted',
                          spacing.chip.sm,
                          radius.default,
                          'bg-surface-base inline-block',
                        )}
                      >
                        {iface.name}
                      </span>
                    </div>
                    {iface.name === currentInterface ? (
                      <CheckIcon
                        className={cn(iconTokens.size.sm, 'text-brand-primary shrink-0')}
                      />
                    ) : null}
                  </div>
                </button>
              ))
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}

interface ConnectionBadgeProps {
  status: ConnectionStatus;
}

function ConnectionBadge({ status }: ConnectionBadgeProps): ReactElement {
  const { t } = useTranslation();
  const isConnected = status === 'connected';
  const isConnecting = status === 'connecting';

  return (
    <div
      className={cn(
        'inline-flex items-center',
        spacing.gap.tight,
        spacing.chip.sm,
        radius.full,
        isConnected
          ? 'bg-status-success/10 text-status-success'
          : 'bg-status-error/10 text-status-error',
      )}
    >
      {isConnected ? (
        <>
          <WifiIcon className={iconTokens.size.xs} />
          <span className="caption font-medium hidden sm:inline">
            {t('status.connected', 'Connected')}
          </span>
        </>
      ) : (
        <>
          <WifiOffIcon className={cn(iconTokens.size.xs, isConnecting && 'animate-pulse')} />
          <span className="caption font-medium hidden sm:inline">
            {isConnecting
              ? t('status.connecting', 'Connecting...')
              : t('status.disconnected', 'Disconnected')}
          </span>
        </>
      )}
    </div>
  );
}

interface ThemeToggleProps {
  darkMode: boolean;
  onToggle: () => void;
}

function ThemeToggle({ darkMode, onToggle }: ThemeToggleProps): ReactElement {
  const { t } = useTranslation();
  const label = darkMode
    ? t('accessibility.switchToLightMode', 'Switch to light mode')
    : t('accessibility.switchToDarkMode', 'Switch to dark mode');

  return (
    <button
      type="button"
      className={iconButtonClass}
      onClick={onToggle}
      aria-label={label}
      title={label}
    >
      {darkMode ? (
        <SunIcon className={iconTokens.size.md} />
      ) : (
        <MoonIcon className={iconTokens.size.md} />
      )}
    </button>
  );
}

// =============================================================================
// Main Component
// =============================================================================

export const HeaderBar: React.FC<HeaderBarProps> = memo(function HeaderBarComponent({
  connectionStatus,
  darkMode,
  onReconnect,
  onToggleTheme,
  onRefresh,
  onHistoryOpen,
  onHelpOpen,
  onSettingsOpen,
  onLogout,
  interfaces = [],
  currentInterface,
  onInterfaceChange,
  profiles = [],
  activeProfile,
  onProfileSwitch,
  onProfileManage,
  profilesLoading = false,
}: HeaderBarProps): ReactElement {
  const { t } = useTranslation();

  const isConnected = connectionStatus === 'connected';
  const isConnecting = connectionStatus === 'connecting';
  const hasInterfaces = interfaces.length > 0 && onInterfaceChange;
  const hasProfiles = profiles.length > 0 && onProfileSwitch;

  const getStatusTooltip = useCallback((): string => {
    const statusMap: Record<ConnectionStatus, string> = {
      connected: t('status.connected', 'Connected'),
      connecting: t('status.connecting', 'Connecting...'),
      disconnected: t('status.disconnected', 'Disconnected'),
      error: t('status.error', 'Connection Error'),
    };
    return statusMap[connectionStatus];
  }, [connectionStatus, t]);

  const handleProfileSelect = useCallback(
    (id: string): void => {
      if (onProfileSwitch) {
        onProfileSwitch(id).catch(() => {
          // Handle profile switch error silently
        });
      }
    },
    [onProfileSwitch],
  );

  const handleInterfaceSelect = useCallback(
    (name: string): void => {
      onInterfaceChange?.(name);
    },
    [onInterfaceChange],
  );

  return (
    <header className="border-b border-surface-border bg-surface-raised">
      <div
        className={cn(
          'mx-auto max-w-7xl',
          spacing.mainPadding.x,
          'py-3',
          layout.flex.between,
          spacing.gap.default,
        )}
      >
        {/* Logo and title */}
        <div className={cn(layout.inline.default, 'min-w-0')}>
          <button
            type="button"
            className={cn(layout.inline.default, 'group', !isConnected && 'cursor-pointer')}
            onClick={isConnected ? undefined : onReconnect}
            title={getStatusTooltip()}
            aria-label={
              isConnected ? getStatusTooltip() : t('status.clickToReconnect', 'Click to reconnect')
            }
          >
            <div
              className={cn(
                'flex h-8 w-8 items-center justify-center rounded-lg',
                isConnected ? 'bg-brand-primary' : 'bg-status-error',
                'text-text-inverse transition-colors',
                !isConnected && 'group-hover:opacity-80',
              )}
            >
              <ActivityIcon className={cn(iconTokens.size.md, isConnecting && 'animate-pulse')} />
            </div>
          </button>
          <div className="min-w-0">
            <h1 className="heading-4 text-text-primary truncate">{t('app.title', 'The Stem')}</h1>
            <p className="caption text-text-muted hidden xs:block">
              {t('app.tagline', 'Mustard Seed Networks')}
            </p>
          </div>
          <ConnectionBadge status={connectionStatus} />
        </div>

        {/* Icon toolbar */}
        <div className={cn('flex items-center', spacing.gap.tight)}>
          {hasProfiles ? (
            <ProfileDropdown
              profiles={profiles}
              activeProfile={activeProfile}
              loading={profilesLoading}
              onSelect={handleProfileSelect}
              onManage={onProfileManage}
            />
          ) : null}
          {hasInterfaces ? (
            <InterfaceDropdown
              interfaces={interfaces}
              currentInterface={currentInterface}
              onSelect={handleInterfaceSelect}
            />
          ) : null}
          <ThemeToggle darkMode={darkMode} onToggle={onToggleTheme} />
          <button
            type="button"
            className={iconButtonClass}
            onClick={onRefresh}
            aria-label={t('accessibility.refreshInterfaces', 'Refresh interfaces')}
            title={t('accessibility.refreshInterfaces', 'Refresh interfaces')}
          >
            <RefreshIcon className={iconTokens.size.md} />
          </button>
          <button
            type="button"
            className={iconButtonClass}
            onClick={onHistoryOpen}
            aria-label={t('accessibility.openHistory', 'Open test history')}
            title={t('history.title', 'Test History')}
          >
            <HistoryIcon className={iconTokens.size.md} />
          </button>
          <button
            type="button"
            className={iconButtonClass}
            onClick={onHelpOpen}
            aria-label={t('accessibility.openHelp', 'Open help')}
            title={t('help.title', 'Help & Documentation')}
          >
            <HelpIcon className={iconTokens.size.md} />
          </button>
          <button
            type="button"
            className={iconButtonClass}
            onClick={onSettingsOpen}
            aria-label={t('accessibility.openSettings', 'Open settings')}
            title={t('settings.title', 'Settings')}
          >
            <SettingsIcon className={iconTokens.size.md} />
          </button>
          <button
            type="button"
            className={iconButtonClass}
            onClick={onLogout}
            aria-label={t('buttons.logout', 'Logout')}
            title={t('buttons.logout', 'Logout')}
          >
            <LogoutIcon className={iconTokens.size.md} />
          </button>
        </div>
      </div>

      {/* Mobile connection status */}
      {!isConnected && (
        <div
          className={cn(
            'sm:hidden',
            spacing.mainPadding.x,
            spacing.padding.bottom.inline,
            layout.flex.center,
          )}
        >
          <button
            type="button"
            onClick={onReconnect}
            className={cn(
              'caption flex items-center gap-1.5',
              isConnecting ? 'text-status-warning' : 'text-status-error',
            )}
          >
            {isConnecting ? (
              <>
                <RefreshIcon className={cn(iconTokens.size.sm, 'animate-spin')} />
                {t('status.connecting', 'Connecting...')}
              </>
            ) : (
              <>
                <span>●</span>
                {t('status.tapToReconnect', 'Tap to reconnect')}
              </>
            )}
          </button>
        </div>
      )}
    </header>
  );
});

export default HeaderBar;
