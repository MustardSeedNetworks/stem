/**
 * @fileoverview TestResults — the outcome of the run, under a page that runs tests.
 * @description Renders the latest test result, or one sentence about a run that
 *              has not produced one yet. With nothing to say (idle, no result)
 *              it renders nothing: a strip reading "no tests running" under
 *              every route was a dead region. Extracted from App.tsx during the
 *              W5.5 providers+routing decomposition.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { Alert } from '../shared/Alert';
import { Card } from '../shared/Card';
import { StatusPill, type StatusTone } from '../shared/StatusPill';
import type { Stats, TestResult } from '../types/api';

export interface TestResultsProps {
  testStatus: Stats['testStatus'];
  result: TestResult | null;
}

const STEP_TONES: Record<string, StatusTone> = {
  completed: 'success',
  passed: 'success',
  failed: 'error',
  error: 'error',
  running: 'info',
  stopped: 'warning',
  cancelled: 'warning',
};

export function TestResults({ testStatus, result }: TestResultsProps): ReactElement | null {
  const { t } = useTranslation('common');
  if (!result) {
    // A run that has begun but has no result yet: what is happening, in one
    // sentence. Idle has nothing to say, so it renders nothing.
    switch (testStatus) {
      case 'starting':
        return <Alert tone="info">{t('results.placeholder.starting')}</Alert>;
      case 'running':
        return <Alert tone="info">{t('results.placeholder.running')}</Alert>;
      case 'cancelled':
        return <Alert tone="warning">{t('results.placeholder.cancelled')}</Alert>;
      case 'stopped':
        return <Alert tone="warning">{t('results.placeholder.stopped')}</Alert>;
      case 'error':
        return <Alert tone="error">{t('results.placeholder.error')}</Alert>;
      default:
        return null;
    }
  }

  // A run the operator stopped has no verdict: result.success is the start
  // acknowledgement for a reflector, so reading it would print PASSED over a
  // run that never finished (#1248).
  const stopped = result.status === 'stopped';
  const verdict: { tone: StatusTone; label: string } = stopped
    ? { tone: 'warning', label: t('status.stopped') }
    : result.success
      ? { tone: 'success', label: t('status.passed') }
      : { tone: 'error', label: t('status.failed') };

  return (
    <Card
      title={t('labels.testResults')}
      headerAction={<StatusPill tone={verdict.tone} label={verdict.label} />}
      data-testid="test-results"
    >
      <p className="text-base font-semibold text-text-primary">{result.testType}</p>
      <p className="text-sm text-text-muted">
        {t('results.moduleName', { module: result.module })}
      </p>

      {result.error ? (
        <Alert tone="error" className="mt-3">
          {result.error}
        </Alert>
      ) : null}

      {result.steps && result.steps.length > 0 ? (
        <ul
          className="mt-4 divide-y divide-hairline border-t border-hairline"
          aria-label={t('accessibility.runPlanResults')}
        >
          {result.steps.map((step, index) => (
            <li
              key={`${step.testType}-${index}`}
              data-testid={`run-plan-result-${index}`}
              data-status={step.status}
              className="py-3"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="font-medium text-text-primary">{step.testType}</span>
                <StatusPill
                  tone={STEP_TONES[step.status] ?? 'unknown'}
                  label={t(`status.${step.status}`, { defaultValue: step.status })}
                />
              </div>
              {step.error ? (
                <p className="mt-1 text-sm text-status-error-strong">{step.error}</p>
              ) : null}
              {step.result?.data ? (
                <pre className="mt-2 max-w-full overflow-auto text-xs text-text-muted">
                  {JSON.stringify(step.result.data, null, 2)}
                </pre>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
    </Card>
  );
}
