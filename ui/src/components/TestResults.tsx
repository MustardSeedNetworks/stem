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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

/**
 * The longest test-host stall in a Y.1564 result that could alone fail FD or
 * FDV, or null when no step and no performance half is host-limited (#1717).
 * The daemon returns a config result (`Steps`), a performance result, or both
 * under `config` and `performance`; the dataplane structs carry no JSON tags,
 * so the keys are the Go field names.
 */
export function hostLimitedStallMs(data: unknown): number | null {
  if (!isRecord(data)) return null;
  const halves =
    'config' in data || 'performance' in data ? [data.config, data.performance] : [data];
  const measurements = halves.flatMap((half) =>
    isRecord(half) && Array.isArray(half.Steps) ? half.Steps : [half],
  );
  let longest: number | null = null;
  for (const m of measurements) {
    if (isRecord(m) && m.HostLimited === true && typeof m.HostStallMaxMs === 'number') {
      longest = Math.max(longest ?? 0, m.HostStallMaxMs);
    }
  }
  return longest;
}

function HostLimitedWarning({ data }: { data: unknown }): ReactElement | null {
  const { t } = useTranslation('common');
  const stallMs = hostLimitedStallMs(data);
  if (stallMs === null) return null;
  return (
    <div
      data-testid="host-limited-warning"
      className="mb-content pad-sm rounded-lg bg-status-warning/10 border border-status-warning/20 text-sm text-status-warning-strong"
    >
      {t('results.hostLimited', { ms: stallMs.toFixed(1) })}
    </div>
  );
}

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
              data-testid={`run-plan-result-${index}`}
              data-status={step.status}
              className="pad-sm rounded-lg bg-surface-base border border-surface-border"
            >
              <div className="flex-between">
                <span className="font-medium text-text-primary">{step.testType}</span>
                <span>{step.status.toUpperCase()}</span>
              </div>
              {step.error ? <div className="text-status-error">{step.error}</div> : null}
              <HostLimitedWarning data={step.result?.data} />
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
          <div className="text-sm text-text-muted">
            {t('results.moduleName', { module: result.module })}
          </div>
        </div>
        <div className="text-right">
          <div className={`heading-3 ${statusColor}`}>{verdict}</div>
        </div>
      </div>

      <HostLimitedWarning data={result.data} />

      {/* Error Message */}
      {result.error ? (
        <div className="mb-content pad-sm rounded-lg bg-status-error/10 border border-status-error/20">
          <div className="text-sm font-medium text-status-error-strong">{t('status.error')}</div>
          <div className="text-sm text-text-primary">{result.error}</div>
        </div>
      ) : null}
    </div>
  );
}
