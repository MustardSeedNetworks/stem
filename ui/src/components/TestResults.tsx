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
  // Show placeholder messages when no result data
  if (!result) {
    let message: string;
    switch (testStatus) {
      case 'idle':
        message = 'No tests running. Configure tests in Settings and click Start.';
        break;
      case 'starting':
        message = 'Test is starting. Results will stream in shortly.';
        break;
      case 'running':
        message = 'Test in progress... Results will appear here when complete.';
        break;
      case 'cancelled':
        message = 'Test cancelled. Adjust settings or restart when ready.';
        break;
      case 'error':
        message = 'An error occurred during the test.';
        break;
      default:
        message = 'Waiting for the backend to report a status.';
    }

    return (
      <div className="card">
        <div className="card-header">
          <AlertTriangle className="w-4 h-4" />
          Test Results
        </div>
        <div className="text-center py-centered text-text-muted">
          <p>{message}</p>
        </div>
      </div>
    );
  }

  // Show actual test results
  const statusColor = result.success ? 'text-status-success' : 'text-status-error';

  return (
    <div className="card">
      <div className="card-header">
        <Activity className="w-4 h-4" />
        Test Results
      </div>

      {/* Test Header */}
      <div className="flex-between mb-content pb-4 border-b border-surface-border">
        <div>
          <div className="heading-3 text-text-primary">{result.testType}</div>
          <div className="text-sm text-text-muted">Module: {result.module}</div>
        </div>
        <div className="text-right">
          <div className={`heading-3 ${statusColor}`}>{result.success ? 'PASSED' : 'FAILED'}</div>
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
