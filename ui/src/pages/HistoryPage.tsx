import { useAppContext } from '../contexts/AppContext';

/**
 * History page — shows the latest test result and points the user at
 * the History drawer for the full archive. The drawer remains mounted
 * at AppShell level (it's a long-lived component that already manages
 * its own data fetching) and is opened via the sidebar History button.
 */
export function HistoryPage() {
  const { testResult } = useAppContext();
  return (
    <>
      {testResult ? (
        <div className="rounded-lg border border-surface-border bg-surface-raised pad-lg stack-lg">
          <div className="flex flex-wrap items-start justify-between gap-comfortable">
            <div>
              <h2 className="heading-3">{testResult.testType}</h2>
              <p className="caption text-text-muted">Module: {testResult.module}</p>
            </div>
            <span
              className={`status-badge ${testResult.success ? 'success' : 'error'}`}
              role="status"
            >
              {testResult.success ? 'PASSED' : 'FAILED'}
            </span>
          </div>
          {testResult.error ? <div className="alert error">{testResult.error}</div> : null}
          {testResult.metrics && Object.keys(testResult.metrics).length > 0 ? (
            <div>
              <div className="section-title mb-2">Metrics</div>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-default">
                {Object.entries(testResult.metrics).map(([k, v]) => (
                  <div
                    key={k}
                    className="rounded-lg bg-surface-base border border-surface-border pad-sm"
                  >
                    <div className="caption capitalize">{k.replace(/_/g, ' ')}</div>
                    <div className="heading-3 text-text-primary">{String(v)}</div>
                  </div>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      ) : (
        <div className="rounded-lg border border-surface-border bg-surface-raised pad-lg text-sm text-text-muted">
          No test has completed yet in this session. Run a test from the Tests pages, then return
          here for the result snapshot.
        </div>
      )}
    </>
  );
}
