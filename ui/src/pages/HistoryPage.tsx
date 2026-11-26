/**
 * HistoryPage — a table of completed runs and the selected run's detail.
 *
 * Every completed run is a record: the table carries test type, module, time
 * and verdict; the detail card carries the verdict, the error when there is
 * one, and the run's metrics.
 *
 * The archive used to live in a drawer while this page showed only the last
 * result. One thing, two surfaces — so the drawer is gone and the page owns
 * it, which is also what the shell convention says (History is a nav item).
 */

import { History } from 'lucide-react';
import { type JSX, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useTableLabels } from '../hooks/useTableLabels';
import { Alert } from '../shared/Alert';
import { Button } from '../shared/Button';
import { Card } from '../shared/Card';
import { ConfirmDialog } from '../shared/ConfirmDialog';
import { DataTable, type DataTableColumn } from '../shared/DataTable';
import { EmptyState } from '../shared/EmptyState';
import { SegmentedControl } from '../shared/SegmentedControl';
import { StatusPill, type StatusTone } from '../shared/StatusPill';
import {
  downloadFile,
  exportFileName,
  historyToCsv,
  historyToJson,
} from '../stores/history-export';
import { HISTORY_MAX_ITEMS, type HistoricalResult, useHistoryStore } from '../stores/history-store';
import { RefreshAction } from '../ui/PageActions';

type Facet = 'all' | 'failed';

/**
 * A run that recorded no verdict is unknown, not a pass. `success` is optional
 * on the wire, and treating a missing verdict as success is how a broken run
 * comes to look like a good one.
 */
function runTone(result: HistoricalResult): StatusTone {
  if (result.success === undefined) {
    return 'unknown';
  }
  return result.success ? 'success' : 'error';
}

function formatDuration(ms?: number): string {
  if (ms === undefined) {
    return '—';
  }
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

function formatWhen(iso?: string): string {
  return iso ? new Date(iso).toLocaleString() : '—';
}

export function HistoryPage(): JSX.Element {
  const { t } = useTranslation('pages');
  const { t: tCommon } = useTranslation('common');
  const results = useHistoryStore((s) => s.results);
  const remove = useHistoryStore((s) => s.remove);
  const clear = useHistoryStore((s) => s.clear);
  const reload = useHistoryStore((s) => s.reload);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [facet, setFacet] = useState<Facet>('all');
  const [confirm, setConfirm] = useState<'delete' | 'clear' | null>(null);
  const tableLabels = useTableLabels<HistoricalResult>((row) => row.testType);

  const failedCount = results.filter((r) => r.success === false).length;
  const shown = facet === 'failed' ? results.filter((r) => r.success === false) : results;
  const selected = shown.find((r) => r.id === selectedId) ?? shown[0] ?? null;

  const verdictLabel = (result: HistoricalResult): string => {
    if (result.success === undefined) {
      return t('history.noVerdict');
    }
    return result.success ? t('history.passed') : t('history.failed');
  };

  const columns: DataTableColumn<HistoricalResult>[] = [
    {
      key: 'testType',
      header: t('history.colTest'),
      cell: (r) => <span className="font-medium text-text-primary">{r.testType}</span>,
      sortValue: (r) => r.testType,
    },
    {
      key: 'module',
      header: t('history.colModule'),
      cell: (r) => r.module,
      sortValue: (r) => r.module,
    },
    {
      key: 'completedAt',
      header: t('history.colCompleted'),
      cell: (r) => <span className="figure">{formatWhen(r.completedAt)}</span>,
      sortValue: (r) => r.completedAt ?? '',
    },
    {
      key: 'verdict',
      header: t('history.colVerdict'),
      cell: (r) => <StatusPill tone={runTone(r)} label={verdictLabel(r)} />,
      sortValue: (r) => verdictLabel(r),
    },
  ];

  if (results.length === 0) {
    return (
      <>
        <RefreshAction onRefresh={reload} />
        <EmptyState
          icon={<History className="h-6 w-6" />}
          title={t('history.emptyNoRunsTitle')}
          hint={t('history.emptyNoRuns')}
        />
      </>
    );
  }

  return (
    <>
      <RefreshAction onRefresh={reload} />
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-text-secondary" data-testid="history-scope">
          {t('history.scope', { count: results.length, max: HISTORY_MAX_ITEMS })}
        </p>
        <div className="flex flex-wrap items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            data-testid="history-export-csv"
            onClick={(): void =>
              downloadFile(
                exportFileName(new Date(), 'csv'),
                'text/csv;charset=utf-8',
                historyToCsv(results),
              )
            }
          >
            {t('history.exportCsv')}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            data-testid="history-export-json"
            onClick={(): void =>
              downloadFile(
                exportFileName(new Date(), 'json'),
                'application/json',
                historyToJson(results),
              )
            }
          >
            {t('history.exportJson')}
          </Button>
          <Button
            variant="destructive"
            size="sm"
            data-testid="history-clear"
            onClick={(): void => setConfirm('clear')}
          >
            {t('history.clear')}
          </Button>
        </div>
      </div>

      <SegmentedControl
        label={t('history.filterLabel')}
        value={facet}
        onChange={(value): void => setFacet(value as Facet)}
        options={[
          { value: 'all', label: `${t('history.filterAll')} (${results.length})` },
          { value: 'failed', label: `${t('history.filterFailed')} (${failedCount})` },
        ]}
      />

      <DataTable
        rows={shown}
        columns={columns}
        getRowKey={(r) => r.id}
        caption={t('history.caption')}
        empty={<EmptyState title={t('history.emptyNoMatch')} />}
        labels={tableLabels}
        defaultSort={{ key: 'completedAt', direction: 'desc' }}
        onRowClick={(r): void => setSelectedId(r.id)}
        {...(selected ? { selectedRowKey: selected.id } : {})}
        data-testid="history-table"
      />

      {selected ? (
        <Card
          title={selected.testType}
          headerAction={<StatusPill tone={runTone(selected)} label={verdictLabel(selected)} />}
          data-testid="history-detail"
        >
          <p className="mb-3 text-sm text-text-muted">
            {t('history.selectedRun')} · {selected.module} · {formatWhen(selected.completedAt)}
          </p>
          {selected.error ? (
            <Alert tone="error" title={t('history.error')} className="mb-3">
              {selected.error}
            </Alert>
          ) : null}

          <dl className="grid gap-3 sm:grid-cols-2">
            {[
              { label: t('history.status'), value: selected.status || '—' },
              { label: t('history.duration'), value: formatDuration(selected.duration) },
              { label: t('history.started'), value: formatWhen(selected.startedAt) },
              { label: t('history.completed'), value: formatWhen(selected.completedAt) },
            ].map((item) => (
              <div key={item.label}>
                <dt className="kicker">{item.label}</dt>
                <dd className="figure text-sm text-text-primary">{item.value}</dd>
              </div>
            ))}
          </dl>

          <RunMetrics metrics={selected.metrics} />

          <div className="mt-4 border-t border-hairline pt-3">
            <Button
              variant="destructive"
              size="sm"
              data-testid="history-delete-run"
              onClick={(): void => setConfirm('delete')}
            >
              {t('history.deleteRun')}
            </Button>
          </div>
        </Card>
      ) : null}

      <ConfirmDialog
        open={confirm !== null}
        title={
          confirm === 'clear' ? t('history.confirmClearTitle') : t('history.confirmDeleteTitle')
        }
        message={
          confirm === 'clear' ? t('history.confirmClearMessage') : t('history.confirmDeleteMessage')
        }
        confirmLabel={confirm === 'clear' ? t('history.clear') : t('history.deleteRun')}
        cancelLabel={tCommon('buttons.cancel')}
        closeLabel={tCommon('buttons.close')}
        onCancel={(): void => setConfirm(null)}
        onConfirm={(): void => {
          if (confirm === 'clear') {
            clear();
          } else if (selected) {
            remove(selected.id);
          }
          setSelectedId(null);
          setConfirm(null);
        }}
      />
    </>
  );
}

/** The run's measurements: label and value pairs. */
function RunMetrics({
  metrics,
}: {
  metrics?: Record<string, number | string>;
}): JSX.Element | null {
  const { t } = useTranslation('pages');
  const entries = Object.entries(metrics ?? {});
  if (entries.length === 0) {
    return null;
  }
  return (
    <div className="mt-4 border-t border-hairline pt-3">
      <p className="kicker mb-2">{t('history.metrics')}</p>
      <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {entries.map(([key, value]) => (
          <div key={key}>
            <dt className="text-xs capitalize text-text-muted">{key.replace(/_/g, ' ')}</dt>
            <dd className="figure text-sm text-text-primary">{String(value)}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
