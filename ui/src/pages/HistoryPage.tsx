/**
 * HistoryPage — List + detail.
 *
 * Every completed run is a record: the list carries test type, module and
 * verdict; the detail pane carries the verdict, the error when there is one,
 * and the run's metrics.
 *
 * The archive used to live in a drawer while this page showed only the last
 * result. One thing, two surfaces — so the drawer is gone and the page owns
 * it, which is also what the shell convention says (History is a nav item).
 */

import { type JSX, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { type HistoricalResult, useHistoryStore } from '../stores/history-store';
import {
  DetailEmpty,
  DetailFacts,
  DetailPane,
  FilterChip,
  ListDetail,
  RecordPane,
  RecordRow,
  type RecordState,
} from '../ui/ListDetail';

type Facet = 'all' | 'failed';

/**
 * A run that recorded no verdict is unknown, not a pass. `success` is optional
 * on the wire, and treating a missing verdict as success is how a broken run
 * comes to look like a good one.
 */
function runState(result: HistoricalResult): RecordState {
  if (result.success === undefined) {
    return 'unknown';
  }
  return result.success ? 'ok' : 'crit';
}

function verdict(result: HistoricalResult): string {
  if (result.success === undefined) {
    return '—';
  }
  return result.success ? 'PASS' : 'FAIL';
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
  const results = useHistoryStore((s) => s.results);
  const remove = useHistoryStore((s) => s.remove);
  const clear = useHistoryStore((s) => s.clear);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [facet, setFacet] = useState<Facet>('all');

  const shown = facet === 'failed' ? results.filter((r) => r.success === false) : results;
  const selected = shown.find((r) => r.id === selectedId) ?? shown[0] ?? null;

  return (
    <>
      <div className="flex-between">
        <p className="body-small">{t('history.runsRecorded', { count: results.length })}</p>
        {results.length > 0 ? (
          <button
            type="button"
            onClick={(): void => {
              clear();
              setSelectedId(null);
            }}
            className="rounded-md px-3 py-2 text-sm text-text-muted hover:text-text-primary"
          >
            {t('history.clear')}
          </button>
        ) : null}
      </div>

      <ListDetail>
        <RecordPane
          chips={
            <>
              <FilterChip
                label={t('history.filterAll')}
                count={results.length}
                active={facet === 'all'}
                onClick={(): void => setFacet('all')}
              />
              <FilterChip
                label={t('history.filterFailed')}
                count={results.filter((r) => r.success === false).length}
                active={facet === 'failed'}
                onClick={(): void => setFacet('failed')}
              />
            </>
          }
          empty={results.length === 0 ? t('history.emptyNoRuns') : t('history.emptyNoMatch')}
        >
          {shown.map((result) => (
            <RecordRow
              key={result.id}
              data-testid={`history-row-${result.id}`}
              name={result.testType}
              nameKind="prose"
              meta={`${result.module} · ${formatWhen(result.completedAt)}`}
              value={verdict(result)}
              state={runState(result)}
              selected={selected?.id === result.id}
              onSelect={(): void => setSelectedId(result.id)}
            />
          ))}
        </RecordPane>

        {selected ? (
          <DetailPane
            eyebrow={t('history.selectedRun')}
            title={selected.testType}
            meta={`${selected.module} · ${formatWhen(selected.completedAt)}`}
            status={<RunVerdict result={selected} />}
            actions={
              <button
                type="button"
                onClick={(): void => {
                  remove(selected.id);
                  setSelectedId(null);
                }}
                className="rounded-md px-3 py-2 text-sm text-status-error hover:bg-status-error/10"
              >
                {t('history.deleteRun')}
              </button>
            }
          >
            {selected.error ? (
              <div className="rounded-lg border border-status-error/40 bg-status-error/10 pad-sm">
                <p className="caption text-status-error">{t('history.error')}</p>
                <p className="text-sm text-text-primary">{selected.error}</p>
              </div>
            ) : null}

            <DetailFacts
              items={[
                { label: t('history.status'), value: selected.status || '—' },
                { label: t('history.duration'), value: formatDuration(selected.duration) },
                { label: t('history.started'), value: formatWhen(selected.startedAt) },
                { label: t('history.completed'), value: formatWhen(selected.completedAt) },
              ]}
            />

            <RunMetrics metrics={selected.metrics} />
          </DetailPane>
        ) : (
          <DetailEmpty>{t('history.emptyDetail')}</DetailEmpty>
        )}
      </ListDetail>
    </>
  );
}

/** The verdict, said in words — the row's colour edge is not the only carrier. */
function RunVerdict({ result }: { result: HistoricalResult }): JSX.Element {
  const { t } = useTranslation('pages');
  if (result.success === undefined) {
    return (
      <span className="rounded-lg border border-surface-border bg-surface-sunken px-3 py-1.5 text-xs font-semibold text-text-muted">
        {t('history.noVerdict')}
      </span>
    );
  }
  if (result.success) {
    return (
      <span className="rounded-lg border border-surface-border px-3 py-1.5 text-xs font-semibold text-text-secondary">
        {t('history.passed')}
      </span>
    );
  }
  return (
    <span className="rounded-lg border border-status-error/40 bg-status-error/10 px-3 py-1.5 text-xs font-semibold text-status-error">
      {t('history.failed')}
    </span>
  );
}

/** The run's measurements — the sub-table the archetype calls for. */
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
    <div className="stack-xs">
      <p className="caption">{t('history.metrics')}</p>
      <dl className="grid gap-default sm:grid-cols-2 lg:grid-cols-3">
        {entries.map(([key, value]) => (
          <div
            key={key}
            className="rounded-lg border border-surface-border bg-surface-base pad-sm stack-xs"
          >
            <dt className="caption capitalize">{key.replace(/_/g, ' ')}</dt>
            <dd className="figure text-sm text-text-primary">{String(value)}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
