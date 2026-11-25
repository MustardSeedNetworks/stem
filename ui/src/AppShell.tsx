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
import { TestResults } from './components/TestResults';
import { TestRunControls } from './components/TestRunControls';
import { useConnectionStatus } from './hooks/useConnectionStatus';
import { useNavGroups } from './navGroups';
import { type PageConfig, usePages } from './pageRegistry';
import { useRecordTestResult } from './stores/history-store';
import { useShellStore } from './stores/shell-store';
import { useTestStore } from './stores/test-store';
import type { Stats, TestResult } from './types/api';
import { DaemonUnreachableBanner } from './ui/DaemonUnreachableBanner';
import { NotFoundPanel } from './ui/NotFoundPanel';
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
  onRefresh: () => void;
  onLogout: () => void;
  roleControl: ReactNode;
}

export function AppShell({
  version,
  testResult,
  testStatus,
  status,
  isDark,
  onToggleTheme,
  onRefresh,
  onLogout,
  roleControl,
}: AppShellProps): ReactElement {
  // A run is recorded because it ended, not because a view is open.
  useRecordTestResult(testResult, testStatus);
  const navGroups = useNavGroups();
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

  useEffect(() => {
    if (settingsOpen) {
      setSettingsLoaded(true);
    }
  }, [settingsOpen]);

  useEffect(() => {
    if (helpOpen) {
      setHelpLoaded(true);
    }
  }, [helpOpen]);

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
        onOpenHelp={openHelp}
        onOpenSettings={() => setSettingsOpen(true)}
        onToggleTheme={onToggleTheme}
        isDark={isDark}
        onRefresh={onRefresh}
        onLogout={onLogout}
        roleControl={roleControl}
      >
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
          {helpOpen ? <HelpDrawer isOpen={true} {...helpTopic} onClose={closeHelp} /> : null}
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

  useEffect(() => {
    document.title = `${page.label} | ${productName}`;
  }, [page.label, productName]);

  return (
    <section className="stack-xl">
      <RoutedPageHeader
        kicker={page.eyebrow}
        title={page.label}
        section={page.section}
        subtitle={page.description}
        onHelp={openHelp}
      />
      {page.runControls ? <TestRunControls /> : null}
      {children}
      {page.runControls ? <TestResults testStatus={testStatus} result={testResult} /> : null}
    </section>
  );
}
