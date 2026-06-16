/**
 * @fileoverview TopBar — the shell's top strip + test-master control row.
 * @description Connection badge, role chip, theme/refresh/logout controls, the
 *              test-master interface picker + Start/Stop button, live run
 *              status, and the progress bar. Extracted from App.tsx during the
 *              W5.5 providers+routing decomposition — behavior is unchanged.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  AlertTriangle,
  LogOut,
  Moon,
  Play,
  RefreshCw,
  Square,
  Sun,
  Wifi,
  WifiOff,
} from 'lucide-react';
import type { ReactElement } from 'react';
import type { StemRole } from '../contexts/RoleContext';
import type { InterfaceInfo, Stats } from '../types/api';
import { RoleChip } from './RoleChip';
import type { useTestProgress } from './TestProgressBar';
import { TestProgressBar } from './TestProgressBar';

export interface TopBarProps {
  connected: boolean;
  isDark: boolean;
  onToggleTheme: () => void;
  onRefresh: () => void;
  onLogout: () => void;
  mode: StemRole;
  selectedInterface: string;
  setSelectedInterface: (name: string) => void;
  interfaces: InterfaceInfo[];
  stats: Stats;
  isStartingTest: boolean;
  isStoppingTest: boolean;
  testStartError: string | null;
  onStartTest: () => void;
  onStopTest: () => void;
  testProgress: ReturnType<typeof useTestProgress>;
}

export function TopBar({
  connected,
  isDark,
  onToggleTheme,
  onRefresh,
  onLogout,
  mode,
  selectedInterface,
  setSelectedInterface,
  interfaces,
  stats,
  isStartingTest,
  isStoppingTest,
  testStartError,
  onStartTest,
  onStopTest,
  testProgress,
}: TopBarProps): ReactElement {
  return (
    <div className="px-4 sm:px-6 lg:px-8 pt-6 pb-inline stack-lg">
      {/* Top strip: connection status + role chip + theme/refresh/logout */}
      <div className="flex flex-wrap items-center justify-between gap-default">
        <div className="flex items-center gap-default">
          <div className={`status-badge ${connected ? 'success' : 'error'}`}>
            {connected ? (
              <>
                <Wifi className="h-3 w-3" /> Connected
              </>
            ) : (
              <>
                <WifiOff className="h-3 w-3" /> Disconnected
              </>
            )}
          </div>
        </div>
        <div className="flex items-center gap-compact">
          <RoleChip />
          <button
            type="button"
            data-testid="header-theme-toggle"
            onClick={onToggleTheme}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {isDark ? (
              <Sun className="h-5 w-5" aria-hidden="true" />
            ) : (
              <Moon className="h-5 w-5" aria-hidden="true" />
            )}
          </button>
          <button
            type="button"
            onClick={onRefresh}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title="Refresh interfaces"
            aria-label="Refresh interfaces"
          >
            <RefreshCw className="h-5 w-5" aria-hidden="true" />
          </button>
          <button
            type="button"
            onClick={onLogout}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title="Logout"
            aria-label="Logout"
            data-testid="logout-button"
          >
            <LogOut className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>
      </div>

      {/* Test-Master control row — only when role is test_master. The
          Reflector role drives Start/Stop from the Reflector page. The
          per-test-page Start/Stop is coming in Phase A.1 (#64). */}
      {mode === 'test_master' ? (
        <div className="flex flex-wrap items-center gap-default">
          <select
            value={selectedInterface}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
              setSelectedInterface(e.target.value)
            }
            className="w-48"
            aria-label="Select network interface"
          >
            <option value="">Select Interface</option>
            {interfaces.map((iface) => (
              <option key={iface.name} value={iface.name}>
                {iface.name} ({iface.speed}Mbps)
              </option>
            ))}
          </select>

          {stats.testStatus === 'running' || stats.testStatus === 'starting' ? (
            <button
              type="button"
              onClick={onStopTest}
              className="btn btn-secondary"
              disabled={isStoppingTest}
              aria-busy={isStoppingTest}
            >
              {isStoppingTest ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  Stopping...
                </>
              ) : (
                <>
                  <Square className="w-4 h-4" aria-hidden="true" />
                  Stop Test
                </>
              )}
            </button>
          ) : (
            <button
              type="button"
              onClick={onStartTest}
              className="btn btn-primary"
              disabled={!selectedInterface || isStartingTest}
              aria-busy={isStartingTest}
            >
              {isStartingTest ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  Starting...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" aria-hidden="true" />
                  Start Test
                </>
              )}
            </button>
          )}

          {testStartError ? (
            <div
              className="text-sm text-status-error flex items-center gap-compact"
              role="alert"
              aria-live="assertive"
            >
              <AlertTriangle className="w-4 h-4" aria-hidden="true" />
              {testStartError}
            </div>
          ) : null}

          <div
            className="flex items-center gap-default ml-auto"
            aria-live="polite"
            aria-atomic="true"
          >
            {stats.testStatus === 'running' || stats.testStatus === 'starting' ? (
              <output className="status-badge success flex items-center gap-compact">
                <span
                  className="w-2 h-2 rounded-full bg-status-success animate-pulse"
                  aria-hidden="true"
                />
                {stats.testStatus === 'starting' ? 'Starting' : 'Running'}:{' '}
                {stats.currentTest || mode}
              </output>
            ) : null}
            {stats.testStatus === 'completed' ? (
              <output className="status-badge info">Completed: {stats.currentTest}</output>
            ) : null}
            {stats.testStatus === 'error' ? (
              <output className="status-badge error" role="alert">
                Error: {stats.currentTest || 'Test failed'}
              </output>
            ) : null}
            {stats.testStatus === 'cancelled' ? (
              <output className="status-badge warning">Stopped: {stats.currentTest}</output>
            ) : null}
          </div>
        </div>
      ) : null}

      <TestProgressBar progress={testProgress} />
    </div>
  );
}
