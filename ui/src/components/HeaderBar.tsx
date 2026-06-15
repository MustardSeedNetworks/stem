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

import {
  Activity,
  Check,
  EthernetPort,
  Loader2,
  LogOut,
  Moon,
  RefreshCw,
  Settings,
  Sun,
  User,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { memo, type ReactElement, useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  cn,
  icon as iconTokens,
  layout,
  radius,
  spacing,
  status as statusColor,
} from '../styles/theme';

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
  onLogout?: () => void;
}

function ProfileDropdown({
  profiles,
  activeProfile,
  loading,
  onSelect,
  onManage,
  onLogout,
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
          <Loader2 className={cn(iconTokens.size.md, 'animate-spin')} />
        ) : (
          <User className={iconTokens.size.md} />
        )}
      </button>
      {isOpen ? (
        <div
          className={cn(
            'absolute top-full right-0 mt-tight w-56',
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
                  <div className="flex-between">
                    <span className="body-small text-text-primary truncate">{p.name}</span>
                    {p.id === activeProfile?.id ? (
                      <Check className={cn(iconTokens.size.sm, 'text-brand-primary')} />
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
                  'w-full flex-center',
                  spacing.gap.tight,
                  spacing.pad.sm,
                  'hover:bg-surface-hover text-brand-primary',
                )}
              >
                <Settings className={iconTokens.size.sm} />
                <span className="body-small font-medium">{t('profile.manage', 'Manage')}</span>
              </button>
            </div>
          ) : null}
          {onLogout ? (
            <div className="border-t border-surface-border">
              <button
                type="button"
                onClick={(): void => {
                  setIsOpen(false);
                  onLogout();
                }}
                className={cn(
                  'w-full flex-center',
                  spacing.gap.tight,
                  spacing.pad.sm,
                  'hover:bg-surface-hover text-status-error',
                )}
              >
                <LogOut className={iconTokens.size.sm} />
                <span className="body-small font-medium">{t('buttons.logout', 'Logout')}</span>
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
        <EthernetPort className={iconTokens.size.md} />
      </button>
      {isOpen ? (
        <div
          className={cn(
            'absolute top-full right-0 mt-tight w-64',
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
                  <div className="flex-between">
                    <div className="stack-xs">
                      <div className="flex items-center gap-tight">
                        <span className="body-small text-text-primary font-medium">
                          {getFriendlyInterfaceName(iface.name, iface.type)}
                        </span>
                        {iface.type === 'wifi' && (
                          <Wifi className={cn(iconTokens.size.xs, 'text-text-muted')} />
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
                      <Check className={cn(iconTokens.size.sm, 'text-brand-primary shrink-0')} />
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
        isConnected ? statusColor.badge.success : statusColor.badge.error,
      )}
    >
      {isConnected ? (
        <>
          <Wifi className={iconTokens.size.xs} />
          <span className="caption font-medium hidden sm:inline">
            {t('status.connected', 'Connected')}
          </span>
        </>
      ) : (
        <>
          <WifiOff className={cn(iconTokens.size.xs, isConnecting && 'animate-pulse')} />
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
      {darkMode ? <Sun className={iconTokens.size.md} /> : <Moon className={iconTokens.size.md} />}
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
          'py-row-lg',
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
                isConnected ? 'bg-brand-primary' : statusColor.bg.error,
                'text-text-inverse transition-colors',
                !isConnected && 'group-hover:opacity-80',
              )}
            >
              <Activity className={cn(iconTokens.size.md, isConnecting && 'animate-pulse')} />
            </div>
          </button>
          <div className="min-w-0">
            <h1 className="heading-4 text-text-primary truncate">{t('app.title', 'The Stem')}</h1>
          </div>
          <ConnectionBadge status={connectionStatus} />
        </div>

        {/* Right-slot: per-product context + theme toggle.
         * Help/Settings live in the sidebar footer; Logout lives in the
         * profile dropdown menu. Refresh/History are page-level concerns. */}
        <div className={cn('flex items-center', spacing.gap.tight)}>
          {hasInterfaces ? (
            <InterfaceDropdown
              interfaces={interfaces}
              currentInterface={currentInterface}
              onSelect={handleInterfaceSelect}
            />
          ) : null}
          {hasProfiles ? (
            <ProfileDropdown
              profiles={profiles}
              activeProfile={activeProfile}
              loading={profilesLoading}
              onSelect={handleProfileSelect}
              onManage={onProfileManage}
              onLogout={onLogout}
            />
          ) : null}
          <ThemeToggle darkMode={darkMode} onToggle={onToggleTheme} />
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
            title={
              isConnecting
                ? t('status.connecting', 'Connecting...')
                : t(
                    'status.tapToReconnectHint',
                    'Reconnect to the backend WebSocket and refresh live data',
                  )
            }
            aria-label={t('status.tapToReconnect', 'Tap to reconnect')}
            className={cn(
              'caption flex items-center gap-1.5',
              isConnecting ? statusColor.text.warning : statusColor.text.error,
            )}
          >
            {isConnecting ? (
              <>
                <RefreshCw className={cn(iconTokens.size.sm, 'animate-spin')} />
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
