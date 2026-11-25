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

import { Play, Square } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { useAppContext } from '../contexts/AppContext';
import { useRole } from '../contexts/RoleContext';
import { Alert } from '../shared/Alert';
import { Button } from '../shared/Button';
import { FormField } from '../shared/FormField';
import { Input } from '../shared/Input';
import { Select } from '../shared/Select';
import { StatusPill } from '../shared/StatusPill';
import { StopOutcomeMessage } from './StopOutcomeMessage';
import { TestProgressBar } from './TestProgressBar';

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

  const startBlocked =
    !selectedInterface || !peer.trim() || peerPort < 1 || peerPort > 65535 || isStartingTest;

  return (
    <div className="stack-sm mb-content" data-testid="test-run-controls">
      <div className="flex flex-wrap items-end gap-3">
        <div className="w-52 max-w-full">
          <FormField label={t('accessibility.selectInterface')}>
            {(control) => (
              <Select
                {...control}
                data-testid="interface-select"
                value={selectedInterface}
                onValueChange={setSelectedInterface}
                placeholder={t('accessibility.selectInterface')}
                options={interfaces.map((iface) => ({
                  value: iface.name,
                  label: `${iface.name} (${iface.speed}Mbps)`,
                }))}
              />
            )}
          </FormField>
        </div>

        <div className="w-52 max-w-full">
          <FormField label={t('accessibility.peerHost')}>
            {(control) => (
              <Input
                {...control}
                data-testid="peer-input"
                value={peer}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void => setPeer(e.target.value)}
              />
            )}
          </FormField>
        </div>
        <div className="w-28 max-w-full">
          <FormField label={t('accessibility.peerPort')}>
            {(control) => (
              <Input
                {...control}
                data-testid="peer-port-input"
                type="number"
                min={1}
                max={65535}
                value={peerPort}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  setPeerPort(Number(e.target.value))
                }
              />
            )}
          </FormField>
        </div>

        {isRunning ? (
          <Button
            data-testid="stop-test-button"
            onClick={onStopTest}
            variant="secondary"
            loading={isStopping}
            icon={<Square className="h-4 w-4" aria-hidden="true" />}
          >
            {isStopping ? t('status.stopping') : t('buttons.stopTest')}
          </Button>
        ) : (
          <Button
            data-testid="start-test-button"
            variant="primary"
            onClick={onStartTest}
            disabled={startBlocked}
            loading={isStartingTest}
            icon={<Play className="h-4 w-4" aria-hidden="true" />}
          >
            {isStartingTest ? t('status.starting') : t('buttons.runTest')}
          </Button>
        )}

        <StopOutcomeMessage outcome={stopOutcome} />

        {testStartError ? (
          <div data-testid="test-start-error">
            <Alert tone="error">{testStartError}</Alert>
          </div>
        ) : null}

        <div className="ml-auto flex items-center gap-3" aria-live="polite" aria-atomic="true">
          {isRunning ? (
            <StatusPill
              tone="success"
              label={
                stats.testStatus === 'starting'
                  ? t('status.testStarting', { test: stats.currentTest || role })
                  : t('status.testRunning', { test: stats.currentTest || role })
              }
            />
          ) : null}
          {stats.testStatus === 'completed' ? (
            <StatusPill
              tone="info"
              label={t('status.testCompleted', { test: stats.currentTest })}
            />
          ) : null}
          {stats.testStatus === 'error' ? (
            <span role="alert">
              <StatusPill
                tone="error"
                label={t('status.testError', { test: stats.currentTest || t('status.failed') })}
              />
            </span>
          ) : null}
          {stats.testStatus === 'cancelled' ? (
            <StatusPill
              tone="warning"
              label={t('status.testStopped', { test: stats.currentTest })}
            />
          ) : null}
        </div>
      </div>

      <TestProgressBar progress={testProgress} />
    </div>
  );
}
