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

import { type ReactElement, type ReactNode, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { HelpDrawer } from './components/HelpDrawer';
import { ResultHistory } from './components/ResultHistory';
import { SettingsDrawer } from './components/SettingsDrawer';
import { TestResults } from './components/TestResults';
import { navGroups } from './navGroups';
import { pages } from './pageRegistry';
import { useShellStore } from './stores/shell-store';
import { useTestStore } from './stores/test-store';
import type { Stats, TestResult } from './types/api';
import { PageLoader } from './ui/PageLoader';
import { SidebarLayout } from './ui/Sidebar';

export interface AppShellProps {
  version?: string;
  topBar: ReactNode;
  testResult: TestResult | null;
  testStatus: Stats['testStatus'];
}

export function AppShell({ version, topBar, testResult, testStatus }: AppShellProps): ReactElement {
  const settingsOpen = useShellStore((s) => s.settingsOpen);
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);
  const helpOpen = useShellStore((s) => s.helpOpen);
  const setHelpOpen = useShellStore((s) => s.setHelpOpen);
  const historyOpen = useShellStore((s) => s.historyOpen);
  const setHistoryOpen = useShellStore((s) => s.setHistoryOpen);

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
        onOpenHistory={() => setHistoryOpen(true)}
        topBar={topBar}
      >
        <Suspense fallback={<PageLoader />}>
          <Routes>
            <Route path="/" element={<Navigate to="/reflector" replace={true} />} />
            {pages.map((page) => (
              <Route key={page.path} path={page.path} element={<page.component />} />
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

      <HelpDrawer isOpen={helpOpen} onClose={() => setHelpOpen(false)} />

      <ResultHistory
        isOpen={historyOpen}
        onClose={() => setHistoryOpen(false)}
        currentResult={testResult}
      />
    </>
  );
}
