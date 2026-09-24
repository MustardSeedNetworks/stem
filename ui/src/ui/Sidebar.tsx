/**
 * Sidebar layout shell — persistent collapsible left navigation.
 *
 * Shared shell pattern — kept visually and behaviorally consistent across
 * seed / stem / niac by convention; each repo owns this file independently
 * (no master, no sync). All colors/spacing reference theme tokens;
 * per-product brand identity comes from each repo's index.css token values.
 *
 * Drawer triggers (help, settings, history) call up to the host App
 * via callback props so the actual drawer components stay mounted at
 * AppShell level alongside the existing test/state plumbing.
 */
import { type LucideIcon, Menu, X } from 'lucide-react';
import { createElement, type FC, type ReactNode, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router';
import { Tooltip } from '../components/ui/Tooltip';
import { iconSizes } from '../constants/sizes';
import { prefetchRoute } from '../utils/prefetch';
import { safeGetItem, safeSetItem } from '../utils/storage';
import { type RailStatus, SidebarFooter, SidebarHeader } from './SidebarChrome';

export type { RailStatus };

export interface SidebarNavItem {
  path: string;
  label: string;
  icon: LucideIcon;
  badge?: string;
}

export interface SidebarNavGroup {
  label: string;
  items: SidebarNavItem[];
}

interface SidebarLayoutProps {
  groups: SidebarNavGroup[];
  version?: string;
  children: ReactNode;
  status?: RailStatus;
  /** Rail-footer controls. Omit a callback and its button does not render. */
  onToggleTheme?: () => void;
  isDark?: boolean;
  onRefresh?: () => void;
  onLogout?: () => void;
  /** The role control, rendered in the footer when the rail is expanded. */
  roleControl?: ReactNode;
  /**
   * Drawer callbacks — all optional. Pass only the ones your product uses;
   * the corresponding footer button only renders when its callback is provided.
   * Stem typically uses help/settings/history; seed uses help/settings/profiles;
   * niac uses help/settings. Add more here if a new product needs another drawer.
   */
  onOpenHelp?: () => void;
  onOpenSettings?: () => void;
  onOpenProfiles?: () => void;
}

const STORAGE_KEY = 'stem-sidebar-collapsed';

interface NavItemButtonProps {
  item: SidebarNavItem;
  active: boolean;
  collapsed: boolean;
  onNavigate: (path: string) => void;
}

function badgeClass(badge: string): string {
  if (badge === 'New') return 'bg-status-success/20 text-status-success-strong';
  if (badge === 'Beta') return 'bg-status-warning/20 text-status-warning-strong';
  return 'bg-brand-primary/20 text-brand-primary-strong';
}

const NavItemButton: FC<NavItemButtonProps> = ({ item, active, collapsed, onNavigate }) => (
  <Tooltip text={collapsed ? item.label : undefined}>
    <button
      type="button"
      onClick={() => onNavigate(item.path)}
      onMouseEnter={() => prefetchRoute(item.path)}
      aria-current={active ? 'page' : undefined}
      aria-label={item.label}
      /* 44px minimum target, 11px radius, and a 3px left bar for the active
       route. The bar carries the state rather than a gradient fill: a filled
       row competes with status colour, and the rail is chrome. */
      className={`group relative flex items-center gap-default w-full min-h-11 px-3 py-2.5 rounded-[11px] text-sm font-medium transition-all duration-200 ${
        active
          ? 'bg-[color-mix(in_oklab,var(--color-brand-primary)_16%,transparent)] text-text-primary'
          : 'text-text-muted hover:text-text-primary hover:bg-surface-hover'
      }`}
    >
      {active ? (
        <span
          aria-hidden="true"
          className="absolute inset-y-1 left-0 w-[3px] rounded-full bg-brand-primary"
        />
      ) : null}
      {createElement(item.icon, {
        className: `${iconSizes.lg} flex-shrink-0 ${
          active ? 'text-brand-primary' : 'text-text-muted group-hover:text-text-secondary'
        }`,
      })}
      {!collapsed ? (
        <>
          <span className="flex-1 text-left truncate">{item.label}</span>
          {item.badge ? (
            <span className={`px-1.5 py-0.5 text-xs rounded font-medium ${badgeClass(item.badge)}`}>
              {item.badge}
            </span>
          ) : null}
        </>
      ) : null}
    </button>
  </Tooltip>
);

interface SidebarBodyProps {
  groups: SidebarNavGroup[];
  collapsed: boolean;
  version?: string;
  status?: RailStatus;
  onToggleTheme?: () => void;
  isDark?: boolean;
  onRefresh?: () => void;
  onLogout?: () => void;
  roleControl?: ReactNode;
  onCollapse: () => void;
  onExpand: () => void;
  onNavigate: (path: string) => void;
  isActive: (path: string) => boolean;
  onOpenHelp?: () => void;
  onOpenSettings?: () => void;
  onOpenProfiles?: () => void;
  // Forwarded to SidebarFooter — see comment there.
  surfaceTestIds: boolean;
}

const SidebarBody: FC<SidebarBodyProps> = ({
  groups,
  collapsed,
  version,
  status,
  onCollapse,
  onExpand,
  onNavigate,
  isActive,
  onOpenHelp,
  onOpenSettings,
  onOpenProfiles,
  onToggleTheme,
  isDark,
  onRefresh,
  onLogout,
  roleControl,
  surfaceTestIds,
}) => {
  const { t } = useTranslation();
  // group.label is either a plain display string ("Account") or an
  // i18n key ("common:sections.modules"). t() returns the translation
  // if the key resolves; otherwise the defaultValue (label itself).
  const translateLabel = (label: string): string =>
    label ? t(label, { defaultValue: label }) : '';
  return (
    <>
      <SidebarHeader
        collapsed={collapsed}
        onCollapse={onCollapse}
        status={status}
        surfaceTestIds={surfaceTestIds}
      />
      <nav className="flex-1 overflow-y-auto py-4 px-cell stack-xl">
        {groups.map((group, groupIndex) => (
          <div key={group.label || `nav-group-${String(groupIndex)}`}>
            {!collapsed && group.label ? (
              <h3 className="section-title font-semibold px-3 mb-2">
                {translateLabel(group.label)}
              </h3>
            ) : null}
            {collapsed ? <div className="h-px bg-surface-border mx-2 mb-2" /> : null}
            <div className="stack-xs">
              {group.items.map((item) => (
                <NavItemButton
                  key={item.path}
                  item={item}
                  active={isActive(item.path)}
                  collapsed={collapsed}
                  onNavigate={onNavigate}
                />
              ))}
            </div>
          </div>
        ))}
      </nav>
      <SidebarFooter
        collapsed={collapsed}
        version={version}
        onOpenHelp={onOpenHelp}
        onOpenSettings={onOpenSettings}
        onOpenProfiles={onOpenProfiles}
        onToggleTheme={onToggleTheme}
        isDark={isDark}
        onRefresh={onRefresh}
        onLogout={onLogout}
        roleControl={roleControl}
        onExpand={onExpand}
        surfaceTestIds={surfaceTestIds}
      />
    </>
  );
};

interface MobileTopBarProps {
  mobileOpen: boolean;
  toggleMobile: () => void;
}

const MobileTopBar: FC<MobileTopBarProps> = ({ mobileOpen, toggleMobile }) => {
  const { t } = useTranslation();
  return (
    <header className="lg:hidden fixed top-0 left-0 right-0 z-50 flex-between px-4 py-row-lg bg-surface-raised/95 backdrop-blur-xl border-b border-surface-border">
      <div className="flex items-center gap-compact">
        <div className="h-8 w-8 rounded-[11px] bg-brand-primary flex-center">
          <span className="figure text-xs font-extrabold tracking-tight text-on-brand">ST</span>
        </div>
        <span className="font-display font-bold text-text-primary">{t('app.title')}</span>
      </div>
      <Tooltip text={mobileOpen ? t('accessibility.closeMenu') : t('accessibility.openMenu')}>
        <button
          type="button"
          onClick={toggleMobile}
          className="pad-xs rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-hover transition-colors"
          // The only way into the navigation below the lg breakpoint, and it had
          // no test id — so mobile navigation could not be driven at all (#639).
          data-testid="mobile-nav-toggle"
          // Was hardcoded English in both attributes. `accessibility.closeMenu`
          // already existed and is used by the scrim two elements down; only
          // `openMenu` was missing.

          aria-label={mobileOpen ? t('accessibility.closeMenu') : t('accessibility.openMenu')}
        >
          {mobileOpen ? <X className={iconSizes.lg} /> : <Menu className={iconSizes.lg} />}
        </button>
      </Tooltip>
    </header>
  );
};

export const SidebarLayout: FC<SidebarLayoutProps> = ({
  groups,
  version,
  children,
  status,
  onOpenHelp,
  onOpenSettings,
  onOpenProfiles,
  onToggleTheme,
  isDark,
  onRefresh,
  onLogout,
  roleControl,
}) => {
  const location = useLocation();
  const navigate = useNavigate();
  const [collapsed, setCollapsed] = useState(() => safeGetItem(STORAGE_KEY) === 'true');
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    safeSetItem(STORAGE_KEY, String(collapsed));
  }, [collapsed]);

  useEffect(() => {
    setMobileOpen(false);
  }, []);

  const { t } = useTranslation();

  const isActive = (path: string) =>
    location.pathname === path || (path !== '/' && location.pathname.startsWith(path));

  // Both asides below stay in the DOM regardless of viewport (responsive
  // classes only toggle display, not mount). Only the desktop aside emits
  // sidebar-*-button testids — the mobile copy keeps them undefined so
  // getByTestId in tests resolves to exactly one element under strict
  // mode. Playwright's default 1280x720 viewport renders desktop.
  const body = (surfaceTestIds: boolean) => (
    <SidebarBody
      groups={groups}
      collapsed={collapsed}
      version={version}
      status={status}
      onCollapse={() => setCollapsed(true)}
      onExpand={() => setCollapsed(false)}
      onNavigate={(p) => navigate(p)}
      isActive={isActive}
      onOpenHelp={onOpenHelp}
      onOpenSettings={onOpenSettings}
      onOpenProfiles={onOpenProfiles}
      onToggleTheme={onToggleTheme}
      isDark={isDark}
      onRefresh={onRefresh}
      onLogout={onLogout}
      roleControl={roleControl}
      surfaceTestIds={surfaceTestIds}
    />
  );

  return (
    <div className="min-h-screen text-text-primary bg-gradient-to-br from-surface-base via-surface-raised to-surface-deep">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] focus:px-4 focus:py-row focus:rounded-lg focus:bg-brand-primary focus:text-text-inverse focus:outline-none"
      >
        {t('accessibility.skipToContent')}
      </a>

      <MobileTopBar mobileOpen={mobileOpen} toggleMobile={() => setMobileOpen(!mobileOpen)} />

      {mobileOpen ? (
        <button
          type="button"
          className="lg:hidden fixed inset-0 z-40 bg-scrim/60 backdrop-blur-sm"
          onClick={() => setMobileOpen(false)}
          aria-label={t('accessibility.closeMenu')}
        />
      ) : null}

      <aside
        // Both asides are always in the DOM; only CSS decides which is shown.
        // The desktop copy carries the `sidebar-*` ids (surfaceTestIds=true)
        // and this one carries none, which is what stops strict mode matching
        // two of everything. Naming the container itself is enough to scope a
        // mobile spec to the copy the operator can actually reach.
        //
        // Closed, it is `invisible` as well as off-canvas: parked at x -288..0
        // it was still a reachable landmark whose every link sat in the tab
        // order, and the fleet's 390px gate reads it as content leaving the
        // viewport. Visibility transitions with the slide, so it only hides
        // once the drawer is out of view.
        data-testid="mobile-sidebar"
        className={`lg:hidden fixed top-0 left-0 z-50 h-full w-72 bg-surface-raised/95 backdrop-blur-xl border-r border-surface-border transform transition-[transform,visibility] duration-300 ease-in-out ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full invisible'
        }`}
      >
        <div className="flex flex-col h-full">{body(false)}</div>
      </aside>

      <aside
        /* 224px, a vertical rail gradient, and a hairline right edge. The
           previous 1px solid surface-border drew a hard line down the page;
           the rail should read as a different plane, not a bordered box.
           The width and main's left offset below are one Tailwind step
           (w-56 / pl-56) so they cannot drift — they were 252px against
           256px before the density pass (UI-STEM-18). */
        // Named for the same reason as the mobile copy above: both asides are
        // always in the DOM, so a spec asserting on sidebar copy has to say
        // which surface it means rather than picking an index (#941).
        data-testid="desktop-sidebar"
        className={`hidden lg:flex fixed top-0 left-0 z-40 h-full flex-col bg-gradient-to-b from-rail-from to-rail-to backdrop-blur-xl border-r border-hairline transition-all duration-300 ease-in-out ${
          collapsed ? 'w-16' : 'w-56'
        }`}
      >
        {body(true)}
      </aside>

      <main
        id="main-content"
        className={`transition-all duration-300 ease-in-out pt-16 lg:pt-0 ${
          collapsed ? 'lg:pl-16' : 'lg:pl-56'
        }`}
      >
        <div className="pad sm:pad-lg lg:pad-xl">{children}</div>
      </main>
    </div>
  );
};
