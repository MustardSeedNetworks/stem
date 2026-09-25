/**
 * TestRunControls — the run controls, on the page that runs the test.
 *
 * Owner decision 2026-09-15 (fleet): the shell is the rail plus the page
 * header, and nothing else sits above the page. Stem carried a second bar
 * under the header on every route — interface, peer host, peer port, Start /
 * Stop, the live status and the progress bar — so a Test Master configuring
 * one module was looking at the run controls of all of them, and the History
 * and Security pages carried a Start button that had nothing to do with them.
 *
 * This renders directly under the page header of the five test pages, from
 * `runControls` in the page registry. The Reflector page keeps its own Start /
 * Stop: it runs `reflect`, not the Tests page's selection.
 *
 * The control ids are the ones the E2E suite has always used — the controls
 * moved, they were not replaced.
 */

import { AlertTriangle, Play, RefreshCw, Square } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { useAppContext } from '../contexts/AppContext';
import { useRole } from '../contexts/RoleContext';
import { StopOutcomeMessage } from './StopOutcomeMessage';
import { TestProgressBar } from './TestProgressBar';
import { Button } from './ui/Button';

export function TestRunControls(): ReactElement | null {
  const { t } = useTranslation('common');
  const { role } = useRole();
  const {
    interfaces,
    selectedInterface,
    setSelectedInterface,
    peer,
    setPeer,
    peerPort,
    setPeerPort,
    stats,
    isStartingTest,
    stopOutcome,
    testStartError,
    onStartTest,
    onStopTest,
    testProgress,
  } = useAppContext();

  // A Reflector-role stem runs nothing from a test page — RoleGuard already
  // tells it so, and a disabled run row underneath would say it twice.
  if (role !== 'test_master') {
    return null;
  }

  const isStopping = stopOutcome.kind === 'stopping';
  const isRunning = stats.testStatus === 'running' || stats.testStatus === 'starting';

  return (
    <div className="stack-sm mb-content" data-testid="test-run-controls">
      <div className="flex flex-wrap items-center gap-default">
        <select
          data-testid="interface-select"
          value={selectedInterface}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
            setSelectedInterface(e.target.value)
          }
          className="w-48"
          aria-label={t('accessibility.selectInterface')}
        >
          <option value="">{t('accessibility.selectInterface')}</option>
          {interfaces.map((iface) => (
            <option key={iface.name} value={iface.name}>
              {iface.name} ({iface.speed}Mbps)
            </option>
          ))}
        </select>

        <input
          data-testid="peer-input"
          value={peer}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void => setPeer(e.target.value)}
          className="w-48"
          placeholder={t('accessibility.peerHost')}
          aria-label={t('accessibility.peerHost')}
        />
        <input
          data-testid="peer-port-input"
          type="number"
          min={1}
          max={65535}
          value={peerPort}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
            setPeerPort(Number(e.target.value))
          }
          className="w-28"
          aria-label={t('accessibility.peerPort')}
        />

        {isRunning ? (
          <Button
            data-testid="stop-test-button"
            onClick={onStopTest}
            variant="outline"
            disabled={isStopping}
            aria-busy={isStopping}
          >
            {isStopping ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                {t('status.stopping')}
              </>
            ) : (
              <>
                <Square className="w-4 h-4" aria-hidden="true" />
                {t('buttons.stopTest')}
              </>
            )}
          </Button>
        ) : (
          <Button
            data-testid="start-test-button"
            onClick={onStartTest}
            disabled={
              !selectedInterface ||
              !peer.trim() ||
              peerPort < 1 ||
              peerPort > 65535 ||
              isStartingTest
            }
            aria-busy={isStartingTest}
          >
            {isStartingTest ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                {t('status.starting')}
              </>
            ) : (
              <>
                <Play className="w-4 h-4" aria-hidden="true" />
                {t('buttons.runTest')}
              </>
            )}
          </Button>
        )}

        <StopOutcomeMessage outcome={stopOutcome} />

        {testStartError ? (
          <div
            className="text-sm text-status-error flex items-center gap-compact"
            role="alert"
            aria-live="assertive"
            data-testid="test-start-error"
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
          {isRunning ? (
            <output className="status-badge success flex items-center gap-compact">
              <span
                className="w-2 h-2 rounded-full bg-status-success animate-pulse"
                aria-hidden="true"
              />
              {stats.testStatus === 'starting'
                ? t('status.testStarting', { test: stats.currentTest || role })
                : t('status.testRunning', { test: stats.currentTest || role })}
            </output>
          ) : null}
          {stats.testStatus === 'completed' ? (
            <output className="status-badge info">
              {t('status.testCompleted', { test: stats.currentTest })}
            </output>
          ) : null}
          {stats.testStatus === 'error' ? (
            <output className="status-badge error" role="alert">
              {t('status.testError', { test: stats.currentTest || t('status.failed') })}
            </output>
          ) : null}
          {stats.testStatus === 'cancelled' ? (
            <output className="status-badge warning">
              {t('status.testStopped', { test: stats.currentTest })}
            </output>
          ) : null}
        </div>
      </div>

      <TestProgressBar progress={testProgress} />
    </div>
  );
}
