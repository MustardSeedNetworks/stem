/**
 * SidebarChrome — the rail's header and footer.
 *
 * Split out of Sidebar.tsx when UI-STEM-9 moved the shell's chrome into the
 * rail and the file passed the 600-line gate. Sidebar.tsx keeps the layout and
 * the navigation; the two ends of the rail live here.
 */
import {
  ChevronLeft,
  ChevronRight,
  HelpCircle,
  LogOut,
  type LucideIcon,
  Moon,
  RefreshCw,
  Settings,
  Sun,
  Users,
} from 'lucide-react';
import { createElement, type FC, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Tooltip } from '../components/ui/Tooltip';
import { iconSizes } from '../constants/sizes';
import { MsnMark } from './MsnMark';

/**
 * The connection state the rail's product mark reports. The dot used to be a
 * hard-coded success green, so the rail said "connected" on a dead socket
 * (UI-STEM-9); the state now comes from the same query the rest of the shell
 * reads, and rides the lockup's accessible name so it is not colour alone.
 */
export interface RailStatus {
  state: 'connected' | 'disconnected';
  label: string;
}

interface FooterIconButtonProps {
  collapsed: boolean;
  onClick: () => void;
  icon: LucideIcon;
  label: string;
  title: string;
  /**
   * Accessible name, when it should differ from the tooltip. "Refresh
   * interfaces" names the control; "Rescan available network interfaces…"
   * describes what it does, and a screen reader wants both, not one twice.
   */
  ariaLabel?: string;
  'data-testid'?: string;
}

export const FooterIconButton: FC<FooterIconButtonProps> = ({
  collapsed,
  onClick,
  icon,
  label,
  title,
  ariaLabel,
  'data-testid': dataTestId,
}) => (
  <Tooltip text={title}>
    <button
      type="button"
      onClick={(event) => {
        event.currentTarget.focus();
        onClick();
      }}
      data-testid={dataTestId}
      className={`${collapsed ? 'w-full' : 'flex-1'} flex items-center ${
        collapsed ? 'justify-center' : 'gap-compact'
      } px-3 py-row rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-hover transition-colors text-sm font-medium`}
      aria-label={ariaLabel ?? title}
    >
      {createElement(icon, { className: `${iconSizes.md} flex-shrink-0` })}
      {!collapsed ? <span>{label}</span> : null}
    </button>
  </Tooltip>
);

export interface SidebarHeaderProps {
  collapsed: boolean;
  onCollapse: () => void;
  status?: RailStatus;
  surfaceTestIds: boolean;
}

export const SidebarHeader: FC<SidebarHeaderProps> = ({
  collapsed,
  onCollapse,
  status,
  surfaceTestIds,
}) => {
  const { t } = useTranslation();
  return (
    <div
      className={`flex items-center ${
        collapsed ? 'justify-center' : 'justify-between'
      } px-3 py-4 border-b border-hairline`}
    >
      <div className={`flex items-center gap-compact ${collapsed ? 'justify-center' : ''}`}>
        <div className="relative flex-shrink-0">
          <div className="h-9 w-9 rounded-[11px] bg-brand-primary flex-center">
            <span className="figure text-sm font-extrabold tracking-tight text-on-brand">ST</span>
          </div>
          {/* The dot is the status readout, so it is named rather than hidden:
              colour alone is not a status, and the name is what a screen
              reader and `toHaveAccessibleName` both read. `role="status"` also
              announces the change when the socket drops. */}
          <Tooltip text={status?.label}>
            <span
              role="status"
              aria-label={status?.label}
              data-testid={surfaceTestIds ? 'rail-status' : undefined}
              data-status={status?.state ?? 'connected'}
              className={`absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-surface-raised ${
                status?.state === 'disconnected' ? 'bg-status-error' : 'bg-status-success'
              }`}
            />
          </Tooltip>
        </div>
        {!collapsed ? (
          <span className="font-display font-bold text-lg text-text-primary tracking-tight">
            {t('app.title')}
          </span>
        ) : null}
      </div>
      {!collapsed ? (
        <Tooltip text={t('accessibility.collapseSidebar')}>
          <button
            type="button"
            onClick={onCollapse}
            className="p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-hover transition-colors lg:flex hidden"
            aria-label={t('accessibility.collapseSidebar')}
          >
            <ChevronLeft className={iconSizes.md} />
          </button>
        </Tooltip>
      ) : null}
    </div>
  );
};

export interface SidebarFooterProps {
  collapsed: boolean;
  version?: string;
  onOpenHelp?: () => void;
  onOpenSettings?: () => void;
  onOpenProfiles?: () => void;
  onToggleTheme?: () => void;
  isDark?: boolean;
  onRefresh?: () => void;
  onLogout?: () => void;
  roleControl?: ReactNode;
  onExpand: () => void;
  // SidebarLayout mounts SidebarBody twice (mobile + desktop asides) and
  // both stay in the DOM regardless of viewport — the responsive classes
  // only toggle display, not mount. Emitting the testids on both copies
  // makes every getByTestId('sidebar-*-button') resolve to 2 elements
  // and trip strict-mode. Layout passes `surfaceTestIds=true` only for
  // the desktop aside (the default Playwright viewport is 1280x720, lg+).
  surfaceTestIds: boolean;
}

interface FullWidthDrawerButtonProps {
  onClick: () => void;
  icon: LucideIcon;
  label: string;
  title: string;
  'data-testid'?: string;
}

const FullWidthDrawerButton: FC<FullWidthDrawerButtonProps> = ({
  onClick,
  icon,
  label,
  title,
  'data-testid': dataTestId,
}) => (
  <Tooltip text={title}>
    <button
      type="button"
      onClick={(event) => {
        event.currentTarget.focus();
        onClick();
      }}
      data-testid={dataTestId}
      className="w-full mb-heading flex items-center gap-compact px-3 py-row rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-hover transition-colors text-sm font-medium"
      aria-label={title}
    >
      {createElement(icon, { className: `${iconSizes.md} flex-shrink-0` })}
      <span>{label}</span>
    </button>
  </Tooltip>
);

export const SidebarFooter: FC<SidebarFooterProps> = ({
  collapsed,
  version,
  onOpenHelp,
  onOpenSettings,
  onOpenProfiles,
  onToggleTheme,
  isDark = false,
  onRefresh,
  onLogout,
  roleControl,
  onExpand,
  surfaceTestIds,
}) => {
  const { t } = useTranslation();
  const themeLabel = isDark
    ? t('accessibility.switchToLightMode')
    : t('accessibility.switchToDarkMode');

  return (
    <div className={`px-3 py-4 border-t border-surface-border ${collapsed ? 'text-center' : ''}`}>
      {/* The role control is a two-value segmented chip, so it needs the
          expanded rail's width. Collapsed, the operator expands to switch —
          the same trade the profiles button below already makes. */}
      {roleControl && !collapsed ? <div className="mb-heading">{roleControl}</div> : null}

      <div className={`${collapsed ? 'stack-sm' : 'flex items-center gap-compact'} mb-heading`}>
        {onOpenHelp ? (
          <FooterIconButton
            collapsed={collapsed}
            onClick={onOpenHelp}
            icon={HelpCircle}
            label={t('labels.help')}
            title={t('tooltips.chrome.help')}
            data-testid={surfaceTestIds ? 'sidebar-help-button' : undefined}
          />
        ) : null}
        {onOpenSettings ? (
          <FooterIconButton
            collapsed={collapsed}
            onClick={onOpenSettings}
            icon={Settings}
            label={t('labels.settings')}
            title={t('tooltips.chrome.settings')}
            data-testid={surfaceTestIds ? 'sidebar-settings-button' : undefined}
          />
        ) : null}
      </div>

      {/* Theme, refresh and logout — the three controls the retired top strip
          carried. They are chrome, so they belong with the other chrome. */}
      <div className={`${collapsed ? 'stack-sm' : 'flex items-center gap-compact'} mb-heading`}>
        {onToggleTheme ? (
          <FooterIconButton
            collapsed={collapsed}
            onClick={onToggleTheme}
            icon={isDark ? Sun : Moon}
            label={isDark ? t('labels.lightMode') : t('labels.darkMode')}
            title={themeLabel}
            ariaLabel={themeLabel}
            data-testid={surfaceTestIds ? 'rail-theme-toggle' : undefined}
          />
        ) : null}
        {onRefresh ? (
          <FooterIconButton
            collapsed={collapsed}
            onClick={onRefresh}
            icon={RefreshCw}
            label={t('labels.refresh')}
            title={t('tooltips.chrome.refresh')}
            ariaLabel={t('accessibility.refreshInterfaces')}
            data-testid={surfaceTestIds ? 'rail-refresh' : undefined}
          />
        ) : null}
        {onLogout ? (
          <FooterIconButton
            collapsed={collapsed}
            onClick={onLogout}
            icon={LogOut}
            label={t('buttons.logout')}
            title={t('tooltips.chrome.logout')}
            ariaLabel={t('buttons.logout')}
            data-testid={surfaceTestIds ? 'rail-logout' : undefined}
          />
        ) : null}
      </div>

      {onOpenProfiles && !collapsed ? (
        <FullWidthDrawerButton
          onClick={onOpenProfiles}
          icon={Users}
          label={t('labels.profiles')}
          title={t('tooltips.chrome.profiles')}
        />
      ) : null}

      {version ? (
        <div
          className={`text-xs font-mono text-text-muted ${collapsed ? '' : 'flex-between gap-tight'}`}
        >
          {!collapsed ? <span className="shrink-0">{t('labels.version')}</span> : null}
          {/* A development build's version carries the commit and a -dirty
              suffix, which wrapped onto the label once the rail narrowed to
              224px (UI-STEM-18). Truncate with the full string on hover. */}
          <span className="truncate" title={version}>
            {version}
          </span>
        </div>
      ) : null}
      {/* Whose tool this is, under what it is. Quiet by design: the product mark
        at the top of the rail is the one that has to be recognised. */}
      <MsnMark collapsed={collapsed} className="mt-3" />
      {collapsed ? (
        <Tooltip text={t('accessibility.expandSidebar')}>
          <button
            type="button"
            onClick={onExpand}
            className="mt-inline p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-hover transition-colors"
            aria-label={t('accessibility.expandSidebar')}
          >
            <ChevronRight className={iconSizes.md} />
          </button>
        </Tooltip>
      ) : null}
    </div>
  );
};
