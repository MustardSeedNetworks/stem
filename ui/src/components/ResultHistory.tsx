/**
 * @fileoverview The Stem - Test Result History Component
 * @description Displays and manages historical test results with localStorage persistence.
 */

import { Activity, ChevronDown, ChevronRight, Clock, History, Trash2, X } from 'lucide-react';
import type { ReactElement } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { useFocusTrap } from '../hooks/useFocusTrap';

/** Test result record stored in history */
export interface HistoricalResult {
  id: string;
  testType: string;
  module: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
  success?: boolean;
  error?: string;
  metrics?: Record<string, number | string>;
  data?: Record<string, unknown>;
}

const HISTORY_STORAGE_KEY = 'stem-result-history';
const MAX_HISTORY_ITEMS = 50;

/** Load history from localStorage */
function loadHistory(): HistoricalResult[] {
  if (typeof window === 'undefined') {
    return [];
  }
  try {
    const stored = window.localStorage.getItem(HISTORY_STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as HistoricalResult[];
      return Array.isArray(parsed) ? parsed : [];
    }
  } catch {
    // Ignore parse errors
  }
  return [];
}

/** Save history to localStorage */
function saveHistory(history: HistoricalResult[]): void {
  if (typeof window === 'undefined') {
    return;
  }
  try {
    // Keep only the most recent items
    const trimmed = history.slice(0, MAX_HISTORY_ITEMS);
    window.localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(trimmed));
  } catch {
    // Ignore storage errors
  }
}

/** Generate a unique ID for a result */
function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

interface ResultHistoryProps {
  isOpen: boolean;
  onClose: () => void;
  currentResult?: {
    testType: string;
    module: string;
    status: string;
    startedAt?: string;
    completedAt?: string;
    duration?: number;
    success?: boolean;
    error?: string;
    metrics?: Record<string, number | string>;
    data?: Record<string, unknown>;
  } | null;
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

interface ResultCardProps {
  result: HistoricalResult;
  isExpanded: boolean;
  onToggle: () => void;
  onDelete: () => void;
}

function ResultCard({ result, isExpanded, onToggle, onDelete }: ResultCardProps): ReactElement {
  const statusColor = result.success ? 'text-status-success' : 'text-status-error';

  return (
    <div className="rounded-lg border border-surface-border bg-surface-base overflow-hidden">
      {/* Summary Row */}
      <button
        type="button"
        onClick={onToggle}
        title={
          isExpanded
            ? `Collapse details for ${result.testType}`
            : `Expand to view metrics, duration, and any errors from this ${result.testType} run`
        }
        aria-label={isExpanded ? 'Collapse result details' : 'Expand result details'}
        aria-expanded={isExpanded}
        className="w-full flex-between pad-sm hover:bg-surface-hover transition-colors text-left"
      >
        <div className="flex items-center gap-default">
          {isExpanded ? (
            <ChevronDown className="w-4 h-4 text-text-muted" />
          ) : (
            <ChevronRight className="w-4 h-4 text-text-muted" />
          )}
          <Activity className="w-4 h-4 text-brand-primary" />
          <div>
            <div className="font-medium text-text-primary">{result.testType}</div>
            <div className="text-xs text-text-muted">{result.module}</div>
          </div>
        </div>
        <div className="flex items-center gap-default">
          <span className={`text-sm font-medium ${statusColor}`}>
            {result.success ? 'PASS' : 'FAIL'}
          </span>
          <div className="text-xs text-text-muted flex items-center gap-tight">
            <Clock className="w-3 h-3" />
            {result.completedAt ? new Date(result.completedAt).toLocaleString() : 'N/A'}
          </div>
        </div>
      </button>

      {/* Expanded Details */}
      {isExpanded ? (
        <div className="border-t border-surface-border pad stack-lg">
          {result.error ? (
            <div className="pad-sm rounded-lg bg-status-error/10 border border-status-error/20">
              <div className="text-sm font-medium text-status-error">Error</div>
              <div className="text-sm text-text-primary">{result.error}</div>
            </div>
          ) : null}

          {result.duration !== undefined && (
            <div className="text-sm">
              <span className="text-text-muted">Duration: </span>
              <span className="font-medium">{formatDuration(result.duration)}</span>
            </div>
          )}

          {result.metrics && Object.keys(result.metrics).length > 0 && (
            <div>
              <div className="text-sm font-semibold text-text-muted mb-2">Metrics</div>
              <div className="grid grid-cols-2 gap-compact">
                {Object.entries(result.metrics).map(([key, value]) => (
                  <div
                    key={key}
                    className="pad-xs rounded bg-surface-base border border-surface-border"
                  >
                    <div className="text-xs text-text-muted capitalize">
                      {key.replace(/_/g, ' ')}
                    </div>
                    <div className="font-medium text-text-primary">
                      {typeof value === 'number' ? formatNumber(value) : String(value)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="text-xs text-text-muted stack-xs">
            {result.startedAt ? (
              <div>Started: {new Date(result.startedAt).toLocaleString()}</div>
            ) : null}
            {result.completedAt ? (
              <div>Completed: {new Date(result.completedAt).toLocaleString()}</div>
            ) : null}
          </div>

          <button
            type="button"
            onClick={onDelete}
            title="Permanently remove this result from history; cannot be undone"
            aria-label="Delete this test result from history"
            className="btn btn-ghost text-status-error text-sm"
          >
            <Trash2 className="w-3 h-3" />
            Delete
          </button>
        </div>
      ) : null}
    </div>
  );
}

export function ResultHistory({
  isOpen,
  onClose,
  currentResult,
}: ResultHistoryProps): ReactElement | null {
  const [history, setHistory] = useState<HistoricalResult[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [lastSavedResult, setLastSavedResult] = useState<string | null>(null);
  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: isOpen,
    onEscape: onClose,
  });

  // Load history on mount
  useEffect(() => {
    setHistory(loadHistory());
  }, []);

  // Auto-save new results when currentResult changes
  useEffect(() => {
    if (currentResult?.completedAt && currentResult.completedAt !== lastSavedResult) {
      const newResult: HistoricalResult = {
        id: generateId(),
        ...currentResult,
      };
      setHistory((prev) => {
        const updated = [newResult, ...prev];
        saveHistory(updated);
        return updated;
      });
      setLastSavedResult(currentResult.completedAt);
    }
  }, [currentResult, lastSavedResult]);

  const handleDelete = useCallback((id: string) => {
    setHistory((prev) => {
      const updated = prev.filter((r) => r.id !== id);
      saveHistory(updated);
      return updated;
    });
  }, []);

  const handleClearAll = useCallback(() => {
    setHistory([]);
    saveHistory([]);
    setExpandedId(null);
  }, []);

  const toggleExpand = useCallback((id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  }, []);

  const handleBackdropKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      if (event.key === 'Escape' || event.key === 'Enter') {
        onClose();
      }
    },
    [onClose],
  );

  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-end">
      {/* Backdrop */}
      <button
        type="button"
        className="absolute inset-0 bg-scrim/50 cursor-default"
        onClick={onClose}
        onKeyDown={handleBackdropKeyDown}
        title="Click outside to close the history drawer"
        aria-label="Close history drawer"
      />

      {/* Drawer */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Test History"
        className="relative h-full w-full max-w-lg bg-surface-raised shadow-xl overflow-hidden flex flex-col"
      >
        {/* Header */}
        <div className="flex-between px-6 py-4 border-b border-surface-border">
          <div className="flex items-center gap-compact">
            <History className="w-5 h-5 text-brand-primary" />
            <h2 className="heading-3 text-text-primary">Test History</h2>
            <span className="text-sm text-text-muted">({history.length} results)</span>
          </div>
          <div className="flex items-center gap-compact">
            {history.length > 0 && (
              <button
                type="button"
                onClick={handleClearAll}
                className="btn btn-ghost text-status-error"
                title={`Permanently delete all ${history.length} saved test result${history.length === 1 ? '' : 's'} from local storage; cannot be undone`}
                aria-label="Clear all test history"
              >
                <Trash2 className="w-4 h-4" />
                Clear All
              </button>
            )}
            <button
              type="button"
              onClick={onClose}
              className="btn btn-ghost"
              title="Close the test history drawer and return to the main view"
              aria-label="Close test history"
            >
              <X className="w-5 h-5" aria-hidden="true" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto pad stack">
          {history.length === 0 ? (
            <div className="text-center py-centered text-text-muted">
              <History className="w-12 h-12 mx-auto mb-content opacity-50" />
              <p>No test history yet.</p>
              <p className="text-sm">Completed tests will appear here.</p>
            </div>
          ) : (
            history.map((result) => (
              <ResultCard
                key={result.id}
                result={result}
                isExpanded={expandedId === result.id}
                onToggle={() => toggleExpand(result.id)}
                onDelete={() => handleDelete(result.id)}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}

export default ResultHistory;
