import { Tooltip } from './ui/Tooltip';
/**
 * @fileoverview The Stem - Module Selector Component
 * @description A component that displays tests organized by module (Reflector, Benchmark,
 *              ServiceTest, TrafficGen, Measure, Certify) with module-specific colors.
 *              Allows users to select tests using the module-oriented architecture.
 */

import { type ReactElement, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelpIcon } from './HelpIcon';

// Module definitions matching the Go backend
interface Module {
  name: string;
  displayName: string;
  description: string;
  color: string;
  standard: string;
  tests: string[];
}

const testTranslationKeys = {
  custom_stream: 'tests.trafficgen.stream',
  trafficgen_burst: 'tests.trafficgen.burst',
  trafficgen_multistream: 'tests.trafficgen.multistream',
  tsn_timing: 'tests.tsn.timing',
  tsn_isolation: 'tests.tsn.isolation',
  tsn_latency: 'tests.tsn.latency',
  tsn_full: 'tests.tsn.full',
  mef_config: 'tests.mef.config',
  mef_perf: 'tests.mef.performance',
  mef_full: 'tests.mef.full',
  y1731_delay: 'tests.y1731.delay',
  y1731_loss: 'tests.y1731.loss',
  y1731_slm: 'tests.y1731.slm',
  y1731_loopback: 'tests.y1731.loopback',
  rfc2889_forwarding: 'tests.rfc2889.forwarding',
  rfc2889_caching: 'tests.rfc2889.caching',
  rfc2889_learning: 'tests.rfc2889.learning',
  rfc2889_broadcast: 'tests.rfc2889.broadcast',
  rfc2889_congestion: 'tests.rfc2889.congestion',
  y1564_config: 'tests.y1564.config',
  y1564_perf: 'tests.y1564.performance',
  y1564_full: 'tests.y1564.full',
  rfc6349_throughput: 'tests.rfc6349.capacity',
  rfc6349_path: 'tests.rfc6349.path',
  rfc2544_throughput: 'tests.rfc2544.throughput',
  rfc2544_latency: 'tests.rfc2544.latency',
  rfc2544_frame_loss: 'tests.rfc2544.frameLoss',
  rfc2544_back_to_back: 'tests.rfc2544.backToBack',
  rfc2544_system_recovery: 'tests.rfc2544.systemRecovery',
  rfc2544_reset: 'tests.rfc2544.reset',
  y1564: 'tests.y1564.full',
  mef: 'tests.mef.full',
  tsn: 'tests.tsn.full',
} as const;

interface ModuleSelectorProps {
  selectedTests: string[];
  setSelectedTests: (tests: string[]) => void;
}

export function ModuleSelector({
  selectedTests,
  setSelectedTests,
}: ModuleSelectorProps): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'modules', 'help']);
  const [modules, setModules] = useState<Module[]>([]);
  const [expandedModule, setExpandedModule] = useState<string | null>('benchmark');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch the module catalogue from the daemon.
  //
  // This used to request '/api/modules' — no /v1, unlike every other call in
  // the app. That path does not 404: it falls through to the SPA handler and
  // returns index.html with HTTP 200, so `response.ok` was true, `.json()`
  // threw on HTML, and the catch silently substituted a hardcoded list. The
  // catalogue therefore never reflected the daemon, and nothing surfaced it.
  useEffect(() => {
    const fetchModules = async (): Promise<void> => {
      try {
        const response = await fetch('/api/v1/modules');
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        // An HTML body means the request was routed to the SPA rather than the
        // API. Checking explicitly keeps that failure loud instead of letting
        // it read as a parse error.
        const contentType = response.headers.get('content-type') ?? '';
        if (!contentType.includes('application/json')) {
          throw new Error(`expected JSON, got ${contentType || 'no content-type'}`);
        }
        const data = (await response.json()) as { modules?: Module[] };
        setModules(data.modules ?? []);
        setError(null);
      } catch (cause) {
        // No static fallback. One that cannot be told apart from success is
        // how this went unnoticed; an operator choosing tests the daemon does
        // not offer would fail at start time anyway.
        setModules([]);
        setError(cause instanceof Error ? cause.message : 'Failed to load modules');
      } finally {
        setLoading(false);
      }
    };
    void fetchModules();
  }, []);

  const toggleTest = (test: string): void => {
    if (selectedTests.includes(test)) {
      setSelectedTests(selectedTests.filter((t): boolean => t !== test));
    } else {
      setSelectedTests([...selectedTests, test]);
    }
  };

  const toggleModule = (moduleName: string): void => {
    setExpandedModule(expandedModule === moduleName ? null : moduleName);
  };

  const getSelectedCount = (mod: Module): number =>
    mod.tests.filter((t): boolean => selectedTests.includes(t)).length;

  const selectAllInModule = (mod: Module): void => {
    const newTests = [...selectedTests];
    for (const test of mod.tests) {
      if (!newTests.includes(test)) {
        newTests.push(test);
      }
    }
    setSelectedTests(newTests);
  };

  const deselectAllInModule = (mod: Module): void => {
    setSelectedTests(selectedTests.filter((t): boolean => !mod.tests.includes(t)));
  };

  if (loading) {
    return <div className="text-center py-8 text-text-muted">{t('status.loadingModules')}</div>;
  }

  if (error) {
    return (
      <div role="alert" className="text-center py-8 text-status-error">
        Could not load the test modules from this stem ({error}).
      </div>
    );
  }

  return (
    // Named so a spec can assert on which modules are listed without matching
    // loose text across the whole settings drawer (#941).
    <div className="stack-sm" data-testid="module-selector">
      {modules.map((mod) => (
        <div key={mod.name} className="border border-surface-border rounded-lg overflow-hidden">
          {/* Module Header */}
          <Tooltip
            text={t(
              expandedModule === mod.name
                ? 'modules:card.expand.titleExpanded'
                : 'modules:card.expand.titleCollapsed',
              { name: mod.displayName },
            )}
          >
            <button
              type="button"
              data-testid={`module-toggle-${mod.name}`}
              onClick={(): void => toggleModule(mod.name)}
              aria-label={t(
                expandedModule === mod.name
                  ? 'modules:card.expand.titleExpanded'
                  : 'modules:card.expand.titleCollapsed',
                { name: mod.displayName },
              )}
              aria-expanded={expandedModule === mod.name}
              className="w-full flex-between pad-sm hover:bg-surface-hover transition-colors"
              style={{ borderLeft: `4px solid ${mod.color}` }}
            >
              <div className="flex items-center gap-default">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: mod.color }} />
                <div className="text-left">
                  <div className="font-medium text-sm">{mod.displayName}</div>
                  <div className="text-xs text-text-muted">{mod.standard}</div>
                </div>
              </div>
              <div className="flex items-center gap-compact">
                <span className="text-xs text-text-muted">
                  {getSelectedCount(mod)}/{mod.tests.length}
                </span>
                <svg
                  className={`w-4 h-4 transition-transform ${expandedModule === mod.name ? 'rotate-180' : ''}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </div>
            </button>
          </Tooltip>

          {/* Module Tests */}
          {expandedModule === mod.name && (
            <div className="border-t border-surface-border bg-surface-base">
              {/* Select All / Deselect All */}
              <div className="flex justify-end gap-compact px-3 py-row border-b border-surface-border">
                <Tooltip text={t('help:tooltips.selectAll', { name: mod.displayName })}>
                  <button
                    type="button"
                    onClick={() => selectAllInModule(mod)}
                    className="text-xs text-status-info hover:underline"
                  >
                    {t('buttons.selectAll')}
                  </button>
                </Tooltip>
                <span className="text-text-muted">|</span>
                <Tooltip text={t('help:tooltips.clearAll', { name: mod.displayName })}>
                  <button
                    type="button"
                    onClick={() => deselectAllInModule(mod)}
                    className="text-xs text-text-muted hover:underline"
                  >
                    {t('buttons.deselectAll')}
                  </button>
                </Tooltip>
              </div>

              {/* Test List */}
              <div className="pad-xs stack-xs">
                {mod.tests.map((testId) => {
                  const key = testTranslationKeys[testId as keyof typeof testTranslationKeys];
                  const testInfo = key
                    ? {
                        name: t(`settings:${key}.name`),
                        desc: t(`settings:${key}.desc`),
                        tooltip: t(`settings:${key}.tooltip`),
                      }
                    : {
                        name: testId,
                        desc: '',
                        tooltip: t('help:tooltips.runTest', { test: testId }),
                      };
                  return (
                    <label
                      key={testId}
                      className="flex items-start gap-default pad-xs rounded-lg cursor-pointer hover:bg-surface-hover"
                    >
                      <input
                        type="checkbox"
                        data-testid={`test-checkbox-${testId}`}
                        checked={selectedTests.includes(testId)}
                        onChange={() => toggleTest(testId)}
                        aria-label={t('modules:card.test.ariaLabel', { name: testInfo.name })}
                        className="mt-0.5 w-4 h-4"
                        style={{ accentColor: mod.color }}
                      />
                      <div className="flex-1">
                        <div className="font-medium text-sm flex items-center gap-tight">
                          {testInfo.name}
                          <HelpIcon tooltip={testInfo.tooltip} />
                        </div>
                        {testInfo.desc ? (
                          <div className="text-xs text-text-muted">{testInfo.desc}</div>
                        ) : null}
                      </div>
                    </label>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

// Fallback static module data when API is unavailable

export default ModuleSelector;
