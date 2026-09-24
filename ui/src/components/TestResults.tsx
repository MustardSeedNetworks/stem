/**
 * @fileoverview TestResults — pinned test-outcome card.
 * @description Renders the latest test result (or a status-appropriate
 *              placeholder) beneath the routed page. Extracted from App.tsx
 *              during the W5.5 providers+routing decomposition.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Activity, AlertTriangle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { Stats, TestResult } from '../types/api';

export interface TestResultsProps {
  testStatus: Stats['testStatus'];
  result: TestResult | null;
}

export function TestResults({ testStatus, result }: TestResultsProps): ReactElement {
  const { t } = useTranslation('common');
  // Show placeholder messages when no result data
  if (!result) {
    const placeholders: Record<string, string> = {
      idle: t('results.placeholder.idle', { runTest: t('buttons.runTest') }),
      starting: t('results.placeholder.starting'),
      running: t('results.placeholder.running'),
      cancelled: t('results.placeholder.cancelled'),
      stopped: t('results.placeholder.stopped'),
      error: t('results.placeholder.error'),
    };
    const message = placeholders[testStatus] ?? t('results.placeholder.unknown');

    // One strip, not a panel. This card renders under every route, including
    // the ones where no test can be started, and a header over 96px of empty
    // card read as a region that had failed to load (UI-STEM-18). With no
    // result there is one sentence to show, so it sits on the header's line.
    return (
      <div className="card flex flex-wrap items-center gap-default">
        <div className="card-header mb-0">
          <AlertTriangle className="w-4 h-4" />
          {t('labels.testResults')}
        </div>
        <p className="text-sm text-text-muted">{message}</p>
      </div>
    );
  }

  // Show actual test results. A run the operator stopped has no verdict:
  // result.success is the start acknowledgement for a reflector, so reading it
  // would print PASSED over a run that never finished (#1248).
  const stopped = result.status === 'stopped';
  const statusColor = stopped
    ? 'text-status-warning'
    : result.success
      ? 'text-status-success'
      : 'text-status-error';
  const verdict = stopped ? 'STOPPED' : result.success ? 'PASSED' : 'FAILED';

  return (
    <div className="card">
      <div className="card-header">
        <Activity className="w-4 h-4" />
        {t('labels.testResults')}
      </div>

      {result.steps && result.steps.length > 0 ? (
        <section
          className="mb-content grid gap-default"
          aria-label={t('accessibility.runPlanResults')}
        >
          {result.steps.map((step, index) => (
            <div
              key={`${step.testType}-${index}`}
              className="pad-sm rounded-lg bg-surface-base border border-surface-border"
            >
              <div className="flex-between">
                <span className="font-medium text-text-primary">{step.testType}</span>
                <span>{step.status.toUpperCase()}</span>
              </div>
              {step.error ? <div className="text-status-error">{step.error}</div> : null}
              {step.result?.data ? (
                <pre className="mt-inline overflow-auto text-xs text-text-muted">
                  {JSON.stringify(step.result.data, null, 2)}
                </pre>
              ) : null}
            </div>
          ))}
        </section>
      ) : null}

      {/* Test Header */}
      <div className="flex-between mb-content pb-4 border-b border-surface-border">
        <div>
          <div className="heading-3 text-text-primary">{result.testType}</div>
          <div className="text-sm text-text-muted">Module: {result.module}</div>
        </div>
        <div className="text-right">
          <div className={`heading-3 ${statusColor}`}>{verdict}</div>
        </div>
      </div>

      {/* Error Message */}
      {result.error ? (
        <div className="mb-content pad-sm rounded-lg bg-status-error/10 border border-status-error/20">
          <div className="text-sm font-medium text-status-error-strong">Error</div>
          <div className="text-sm text-text-primary">{result.error}</div>
        </div>
      ) : null}
    </div>
  );
}
