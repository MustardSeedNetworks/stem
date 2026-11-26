/**
 * @fileoverview AppShell — the authenticated application shell.
 * @description Sidebar layout + routed pages + the pinned TestResults card, plus
 *              the Settings / Help / History drawers. The shell is the rail and
 *              the page header; nothing sits above the page (UI-STEM-9). Reads drawer state from the
 *              shell-store and test config from the test-store directly. Mounted
 *              only once signed in. Extracted from App.tsx during the W5.5
 *              providers+routing decomposition.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  lazy,
  type ReactElement,
  type ReactNode,
  Suspense,
  useCallback,
  useEffect,
  useState,
} from 'react';
import { useTranslation } from 'react-i18next';
import { matchPath, Navigate, Route, Routes, useLocation } from 'react-router';
import { useRoleSwitch } from './components/RoleSwitch';
import { TestResults } from './components/TestResults';
import { TestRunControls } from './components/TestRunControls';
import { useConnectionStatus } from './hooks/useConnectionStatus';
import { useNavGroups } from './navGroups';
import { type PageConfig, usePages } from './pageRegistry';
import { documentTitle } from './shared/documentTitle';
import { useRecordTestResult } from './stores/history-store';
import { useShellStore } from './stores/shell-store';
import { useTestStore } from './stores/test-store';
import type { Stats, TestResult } from './types/api';
import { DaemonUnreachableBanner } from './ui/DaemonUnreachableBanner';
import { NotFoundPanel } from './ui/NotFoundPanel';
import { PageActionsSlotContext } from './ui/PageActions';
import { PageLoader } from './ui/PageLoader';
import { RoutedPageHeader } from './ui/RoutedPageHeader';
import { type RailStatus, SidebarLayout } from './ui/Sidebar';
import { routeForPathname, useOpenHelp } from './useOpenHelp';

const HelpDrawer = lazy(() =>
  import('./components/HelpDrawer').then(({ HelpDrawer: component }) => ({ default: component })),
);
const SettingsDrawer = lazy(() =>
  import('./components/SettingsDrawer').then(({ SettingsDrawer: component }) => ({
    default: component,
  })),
);

export interface AppShellProps {
  version?: string;
  testResult: TestResult | null;
  testStatus: Stats['testStatus'];
  status: RailStatus;
  isDark: boolean;
  onToggleTheme: () => void;
  onLogout: () => void;
}

export function AppShell({
  version,
  testResult,
  testStatus,
  status,
  isDark,
  onToggleTheme,
  onLogout,
}: AppShellProps): ReactElement {
  // A run is recorded because it ended, not because a view is open.
  useRecordTestResult(testResult, testStatus);
  const navGroups = useNavGroups();
  const roleSwitch = useRoleSwitch();
  const connection = useConnectionStatus();
  const pages = usePages();
  const location = useLocation();
  const routePath = pages.find((page) => matchPath(page.path, location.pathname))?.path;
  const helpTopic = pages.find((page) => page.path === routePath)?.help;
  const settingsOpen = useShellStore((s) => s.settingsOpen);
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);
  const helpRoute = useShellStore((s) => s.helpRoute);
  const setHelpRoute = useShellStore((s) => s.setHelpRoute);
  const openHelp = useOpenHelp();
  const closeHelp = useCallback(() => setHelpRoute(null), [setHelpRoute]);
  // The drawer belongs to one route, so it renders only on that route. A
  // navigation therefore closes it by arithmetic rather than by an effect that
  // could fire after the next open and swallow it (#1305).
  const helpOpen = helpRoute !== null && helpRoute === routePath;
  const [settingsLoaded, setSettingsLoaded] = useState(settingsOpen);
  const [helpLoaded, setHelpLoaded] = useState(helpOpen);

  // Discard a drawer whose page the user has left. Judged against the address
  // bar, not the committed route: a drawer opened for a route still arriving
  // has not been left, it has not got there yet.
  useEffect(() => {
    if (helpRoute === null || helpRoute === routePath) {
      return;
    }
    if (
      helpRoute !==
      routeForPathname(
        pages.map((page) => page.path),
        window.location.pathname,
      )
    ) {
      setHelpRoute(null);
    }
  }, [helpRoute, routePath, pages, setHelpRoute]);

  // Mount both drawers just after the shell paints and keep them mounted. A
  // dialog records the click that opens it so Safari, which does not focus a
  // button on click, can return focus to the opener; that listener lives in the
  // dialog, so a dialog mounted by the click itself would miss the click.
  useEffect(() => {
    setSettingsLoaded(true);
    setHelpLoaded(true);
  }, []);

  const {
    selectedTests,
    setSelectedTests,
    rfc2544Config,
    setRFC2544Config,
    rfc2889Config,
    setRFC2889Config,
    rfc6349Config,
    setRFC6349Config,
    y1564Config,
    setY1564Config,
    y1731Config,
    setY1731Config,
    tsnConfig,
    setTSNConfig,
    trafficGenConfig,
    setTrafficGenConfig,
  } = useTestStore();

  return (
    <>
      <SidebarLayout
        groups={navGroups}
        version={version}
        status={status}
        productSlot={roleSwitch.productSlot}
        onOpenHelp={openHelp}
        onOpenSettings={() => setSettingsOpen(true)}
        onToggleTheme={onToggleTheme}
        isDark={isDark}
        onLogout={onLogout}
      >
        {roleSwitch.overlay}
        <DaemonUnreachableBanner status={connection} />
        <Suspense fallback={<PageLoader />}>
          <Routes>
            <Route path="/" element={<Navigate to="/reflector" replace={true} />} />
            {pages.map((page) => (
              <Route
                key={page.path}
                path={page.path}
                element={
                  <PageWithHeader page={page} testStatus={testStatus} testResult={testResult}>
                    <page.component />
                  </PageWithHeader>
                }
              />
            ))}
            <Route path="*" element={<NotFoundPanel />} />
          </Routes>
        </Suspense>
      </SidebarLayout>

      {settingsLoaded ? (
        <Suspense fallback={null}>
          <SettingsDrawer
            isOpen={settingsOpen}
            onClose={() => setSettingsOpen(false)}
            selectedTests={selectedTests}
            setSelectedTests={setSelectedTests}
            rfc2544Config={rfc2544Config}
            setRFC2544Config={setRFC2544Config}
            rfc2889Config={rfc2889Config}
            setRFC2889Config={setRFC2889Config}
            rfc6349Config={rfc6349Config}
            setRFC6349Config={setRFC6349Config}
            y1564Config={y1564Config}
            setY1564Config={setY1564Config}
            y1731Config={y1731Config}
            setY1731Config={setY1731Config}
            tsnConfig={tsnConfig}
            setTSNConfig={setTSNConfig}
            trafficGenConfig={trafficGenConfig}
            setTrafficGenConfig={setTrafficGenConfig}
          />
        </Suspense>
      ) : null}

      {helpLoaded ? (
        <Suspense fallback={null}>
          <HelpDrawer
            key={routePath ?? 'none'}
            isOpen={helpOpen}
            {...helpTopic}
            onClose={closeHelp}
          />
        </Suspense>
      ) : null}
    </>
  );
}

/**
 * PageWithHeader renders the section frame every routed page shares — the
 * page header — from the registry entry rather than from the page body. Pages
 * render only their own content. A page that runs a test carries the run
 * controls under its header and the outcome below its body; the others have
 * no test outcome to show, so they carry no results region.
 *
 * It also owns `document.title`, so tabs, bookmarks and history entries
 * name the route rather than all reading the product name from index.html
 * (#1260). The title is the rail label, the same string the user clicked.
 */
function PageWithHeader({
  page,
  testStatus,
  testResult,
  children,
}: {
  page: PageConfig;
  testStatus: Stats['testStatus'];
  testResult: TestResult | null;
  children: ReactNode;
}) {
  const openHelp = useOpenHelp();
  const { t } = useTranslation('common');
  const productName = t('app.title');
  const [actionsSlot, setActionsSlot] = useState<HTMLElement | null>(null);

  useEffect(() => {
    document.title = documentTitle(page.label, productName);
  }, [page.label, productName]);

  return (
    <PageActionsSlotContext.Provider value={actionsSlot}>
      <section className="stack-xl">
        <RoutedPageHeader
          kicker={page.eyebrow}
          title={page.label}
          section={page.section}
          subtitle={page.description}
          actions={<span ref={setActionsSlot} className="contents" />}
          onHelp={openHelp}
        />
        {page.runControls ? <TestRunControls /> : null}
        {children}
        {page.runControls ? <TestResults testStatus={testStatus} result={testResult} /> : null}
      </section>
    </PageActionsSlotContext.Provider>
  );
}
