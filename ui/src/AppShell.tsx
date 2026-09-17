/**
 * @fileoverview AppShell — the authenticated application shell.
 * @description Sidebar layout + routed pages + the pinned TestResults card, plus
 *              the Settings / Help / History drawers. Reads drawer state from the
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
  useRef,
  useState,
} from 'react';
import { matchPath, Navigate, Route, Routes, useLocation } from 'react-router';
import { TestResults } from './components/TestResults';
import { useNavGroups } from './navGroups';
import { type PageConfig, usePages } from './pageRegistry';
import { useShellStore } from './stores/shell-store';
import { useTestStore } from './stores/test-store';
import type { Stats, TestResult } from './types/api';
import { Breadcrumbs } from './ui/Breadcrumbs';
import { PageHeader } from './ui/PageHeader';
import { PageLoader } from './ui/PageLoader';
import { SidebarLayout } from './ui/Sidebar';

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
  topBar: ReactNode;
  testResult: TestResult | null;
  testStatus: Stats['testStatus'];
}

export function AppShell({ version, topBar, testResult, testStatus }: AppShellProps): ReactElement {
  const navGroups = useNavGroups();
  const pages = usePages();
  const location = useLocation();
  const routePath = pages.find((page) => matchPath(page.path, location.pathname))?.path;
  const helpTopic = pages.find((page) => page.path === routePath)?.help;
  const settingsOpen = useShellStore((s) => s.settingsOpen);
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);
  const helpOpen = useShellStore((s) => s.helpOpen);
  const setHelpOpen = useShellStore((s) => s.setHelpOpen);
  const closeHelp = useCallback(() => setHelpOpen(false), [setHelpOpen]);
  const [settingsLoaded, setSettingsLoaded] = useState(settingsOpen);
  const [helpLoaded, setHelpLoaded] = useState(helpOpen);

  const previousRoute = useRef(routePath);
  useEffect(() => {
    if (previousRoute.current !== routePath) {
      previousRoute.current = routePath;
      setHelpOpen(false);
    }
  }, [routePath, setHelpOpen]);

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
        onOpenHelp={() => setHelpOpen(true)}
        onOpenSettings={() => setSettingsOpen(true)}
        topBar={topBar}
      >
        <Suspense fallback={<PageLoader />}>
          <Routes>
            <Route path="/" element={<Navigate to="/reflector" replace={true} />} />
            {pages.map((page) => (
              <Route
                key={page.path}
                path={page.path}
                element={
                  <PageWithHeader page={page}>
                    <page.component />
                  </PageWithHeader>
                }
              />
            ))}
            <Route path="*" element={<Navigate to="/reflector" replace={true} />} />
          </Routes>
        </Suspense>

        {/* Pinned below the routed page so test outcomes stay visible no
        matter which page is active. */}
        <div className="mt-6">
          <TestResults testStatus={testStatus} result={testResult} />
        </div>
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
 * PageWithHeader renders the section frame every routed page shares —
 * breadcrumbs plus the page header — from the registry entry rather
 * than from the page body. Pages render only their own content.
 */
function PageWithHeader({ page, children }: { page: PageConfig; children: ReactNode }) {
  const setHelpOpen = useShellStore((state) => state.setHelpOpen);
  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={page.icon}
        iconColorClass={page.iconColorClass}
        eyebrow={page.eyebrow}
        title={page.title}
        description={page.description}
        onHelp={() => setHelpOpen(true)}
      />
      {children}
    </section>
  );
}
