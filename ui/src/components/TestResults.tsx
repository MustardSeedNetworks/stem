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

function formatNumber(num: number): string {
  if (num >= 1e9) {
    return `${(num / 1e9).toFixed(2)}B`;
  }
  if (num >= 1e6) {
    return `${(num / 1e6).toFixed(2)}M`;
  }
  if (num >= 1e3) {
    return `${(num / 1e3).toFixed(2)}K`;
  }
  return num.toString();
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  const minutes = Math.floor(ms / 60000);
  const seconds = ((ms % 60000) / 1000).toFixed(0);
  return `${minutes}m ${seconds}s`;
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
          {result.duration !== undefined && (
            <div className="text-sm text-text-muted">
              Duration: {formatDuration(result.duration)}
            </div>
          )}
        </div>
      </div>

      {/* Error Message */}
      {result.error ? (
        <div className="mb-content pad-sm rounded-lg bg-status-error/10 border border-status-error/20">
          <div className="text-sm font-medium text-status-error">Error</div>
          <div className="text-sm text-text-primary">{result.error}</div>
        </div>
      ) : null}

      {/* Metrics Grid */}
      {result.metrics && Object.keys(result.metrics).length > 0 && (
        <div className="mb-content">
          <div className="text-sm font-semibold text-text-muted mb-2">Metrics</div>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-default">
            {Object.entries(result.metrics).map(([key, value]) => (
              <div
                key={key}
                className="pad-sm rounded-lg bg-surface-base border border-surface-border"
              >
                <div className="text-xs text-text-muted capitalize">{key.replace(/_/g, ' ')}</div>
                <div className="heading-3 text-text-primary">
                  {typeof value === 'number' ? formatNumber(value) : String(value)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timestamps */}
      <div className="text-xs text-text-muted flex gap-comfortable">
        {result.startedAt ? (
          <span>Started: {new Date(result.startedAt).toLocaleString()}</span>
        ) : null}
        {result.completedAt ? (
          <span>Completed: {new Date(result.completedAt).toLocaleString()}</span>
        ) : null}
      </div>
    </div>
  );
}
