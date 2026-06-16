/**
 * @fileoverview The Stem - Main Application Component
 * @description Composition root: wires the providers (Role, ModuleSettings,
 *              Router, AppContext), assembles the AppContext surface the routed
 *              pages read, and switches between the authenticated AppShell and
 *              the unauthenticated AuthGate overlays. The orchestration logic
 *              lives in useTestExecution; the UI lives in TopBar / AppShell /
 *              AuthGate.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { type ReactElement, useCallback } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { AppShell } from './AppShell';
import { AuthGate } from './components/auth/AuthGate';
import { TopBar } from './components/TopBar';
import { CommandPalette } from './components/ui/CommandPalette';
import { AppContext, type AppContextValue } from './contexts/AppContext';
import { ModuleSettingsProvider } from './contexts/ModuleSettingsContext';
import { RoleProvider, useRole } from './contexts/RoleContext';
import { useBuildVersion } from './hooks/useBuildVersion';
import { useTestExecution } from './hooks/useTestExecution';
import { useTheme } from './hooks/useTheme';
import { navGroups } from './navGroups';
import { useAuthStore } from './stores/auth-store';
import { useShellStore } from './stores/shell-store';
import { useTestStore } from './stores/test-store';

function AppContent(): ReactElement {
  const { isDark, toggleTheme } = useTheme();
  const buildVersion = useBuildVersion();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const { role: mode } = useRole();

  // Command-palette open state + the drawer openers it shares with the shell.
  const paletteOpen = useShellStore((s) => s.paletteOpen);
  const setPaletteOpen = useShellStore((s) => s.setPaletteOpen);
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);
  const setHelpOpen = useShellStore((s) => s.setHelpOpen);

  // Test config lives in the test-store; the routed pages read it via
  // AppContext. The shell drawers read the same store directly.
  const {
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
    selectedTests,
    reflectorProfile,
    setReflectorProfile,
  } = useTestStore();

  // Test/interface/stats orchestration (queries, connection, start/stop).
  const exec = useTestExecution();

  // Logout: the store calls the server, clears the auth flag, and
  // cancel-then-clears the React Query cache (connected follows isAuthenticated).
  const handleLogout = useCallback((): void => {
    void useAuthStore.getState().logout();
  }, []);

  const appContextValue: AppContextValue = {
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
    selectedTests,
    testResult: exec.testResult,
    interfaces: exec.interfaces,
    selectedInterface: exec.selectedInterface,
    setSelectedInterface: exec.setSelectedInterface,
    stats: exec.stats,
    reflectorProfile,
    setReflectorProfile,
    onStartReflector: () => {
      exec.handleStartTest().catch(() => {
        // Errors surface via testStartError state.
      });
    },
    onStopReflector: () => {
      exec.handleStopTest().catch(() => {
        // Errors are already logged inside handleStopTest.
      });
    },
    isStartingReflector: exec.isStartingTest,
    isStoppingReflector: exec.isStoppingTest,
    reflectorStartError: exec.testStartError,
  };

  return (
    <BrowserRouter>
      <AppContext.Provider value={appContextValue}>
        {/* Only mount the authenticated shell once signed in. Rendering the
            full SidebarLayout + lazy routes + live TestResults *behind* the
            login modal was the dominant CLS source (Suspense fallback→page
            swap and WebSocket-driven TestResults height changes shifting the
            background). Unauthenticated users get a stable gradient backdrop
            under the auth overlays — also avoids briefly exposing the app
            shell and loading routes they can't access. */}
        {isAuthenticated ? (
          <AppShell
            version={buildVersion.version}
            testResult={exec.testResult}
            testStatus={exec.stats.testStatus}
            topBar={
              <TopBar
                connected={exec.connected}
                isDark={isDark}
                onToggleTheme={toggleTheme}
                onRefresh={exec.refetchInterfaces}
                onLogout={handleLogout}
                mode={mode}
                selectedInterface={exec.selectedInterface}
                setSelectedInterface={exec.setSelectedInterface}
                interfaces={exec.interfaces}
                stats={exec.stats}
                isStartingTest={exec.isStartingTest}
                isStoppingTest={exec.isStoppingTest}
                testStartError={exec.testStartError}
                onStartTest={() => {
                  exec.handleStartTest().catch(() => {
                    // Errors surface via testStartError state.
                  });
                }}
                onStopTest={() => {
                  exec.handleStopTest().catch(() => {
                    // Errors are already logged inside handleStopTest.
                  });
                }}
                testProgress={exec.testProgress}
              />
            }
          />
        ) : (
          <div className="min-h-screen bg-gradient-to-br from-surface-base via-surface-raised to-surface-deep" />
        )}

        {/* Setup / recovery / login overlays — self-contained on the auth-store. */}
        <AuthGate />

        {/* Command palette (⌘K / Ctrl+K) — authenticated feature only */}
        {isAuthenticated ? (
          <CommandPalette
            groups={navGroups}
            open={paletteOpen}
            onOpenChange={setPaletteOpen}
            onOpenSettings={() => setSettingsOpen(true)}
            onOpenHelp={() => setHelpOpen(true)}
            onToggleTheme={toggleTheme}
            isDark={isDark}
          />
        ) : null}
      </AppContext.Provider>
    </BrowserRouter>
  );
}

// Wrapper component that provides context
function App(): ReactElement {
  return (
    <RoleProvider>
      <ModuleSettingsProvider>
        <AppContent />
      </ModuleSettingsProvider>
    </RoleProvider>
  );
}

export default App;
