/**
 * @fileoverview The Stem - Module Card Component
 * @description Card component for each test module (Benchmark, ServiceTest, etc.)
 *              with enable/disable toggles, autostart options, and test execution.
 */

import { ChevronDown, ChevronUp, Play, Power, RefreshCw, Settings2, Square } from 'lucide-react';
import type { ReactElement } from 'react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, icon as iconTokens, spacing, status as statusColor } from '../styles/theme';
import type { ModuleTestResults } from './ModuleResultsTables';
import {
  FrameSizeResultsTable,
  OamResultsTable,
  ServiceFlowResultsTable,
} from './ModuleResultsTables';

// Re-exported so ModuleSettingsContext and the story keep importing the result
// types from ModuleCard, which is where they have always come from.
export type {
  FrameSizeResult,
  ModuleTestResults,
  OamMeasurementResult,
  ServiceFlowResult,
} from './ModuleResultsTables';

export interface ModuleTest {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
}

export interface ModuleConfig {
  name: string;
  displayName: string;
  description: string;
  color: string;
  standard: string;
  enabled: boolean;
  autoStart: boolean;
  tests: ModuleTest[];
}

export interface ModuleStatus {
  status: 'idle' | 'starting' | 'running' | 'completed' | 'error' | 'cancelled';
  currentTest: string | null;
  progress?: number;
  message?: string;
}

interface ModuleCardProps {
  config: ModuleConfig;
  status: ModuleStatus;
  results?: ModuleTestResults | null;
  onToggleModule: (enabled: boolean) => void;
  onToggleAutoStart: (enabled: boolean) => void;
  onToggleTest: (testId: string, enabled: boolean) => void;
  onStart: () => void;
  onStop: () => void;
  onConfigure: () => void;
}

/** Status indicator badge for module card */
function ModuleStatusIndicator({
  status,
  isRunning,
}: {
  status: ModuleStatus;
  isRunning: boolean;
}): ReactElement | null {
  if (isRunning) {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.successSoft,
        )}
      >
        <span className={cn('w-2 h-2 rounded-full animate-pulse', statusColor.bg.success)} />
        <span className={cn('text-xs font-medium', statusColor.text.success)}>
          {status.status === 'starting' ? 'Starting...' : status.currentTest || 'Running'}
        </span>
      </div>
    );
  }
  if (status.status === 'completed') {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.infoSoft,
        )}
      >
        <span className={cn('text-xs font-medium', statusColor.text.info)}>Completed</span>
      </div>
    );
  }
  if (status.status === 'error') {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.errorSoft,
        )}
      >
        <span className={cn('text-xs font-medium', statusColor.text.error)}>
          Error{status.message ? `: ${status.message}` : ''}
        </span>
      </div>
    );
  }
  return null;
}

/** Start/Stop button for module card */
function ModuleActionButton({
  config,
  isRunning,
  enabledTestCount,
  onStart,
  onStop,
}: {
  config: ModuleConfig;
  isRunning: boolean;
  enabledTestCount: number;
  onStart: () => void;
  onStop: () => void;
}): ReactElement | null {
  const { t } = useTranslation('modules');
  if (!config.enabled) {
    return null;
  }
  if (isRunning) {
    return (
      <button
        type="button"
        onClick={onStop}
        title={t('card.stop.title', { name: config.displayName })}
        aria-label={t('card.stop.ariaLabel', { name: config.displayName })}
        className={cn(
          'px-4 py-row rounded-lg flex items-center gap-compact transition-colors',
          statusColor.badge.error,
          statusColor.hover.errorStrong,
        )}
      >
        <Square className="w-4 h-4" />
        <span className="text-sm font-medium">{t('card.action.stopLabel')}</span>
      </button>
    );
  }
  return (
    <button
      type="button"
      onClick={onStart}
      disabled={enabledTestCount === 0}
      title={
        enabledTestCount === 0
          ? t('card.start.titleEmpty', { name: config.displayName })
          : t('card.start.titleEnabled', {
              count: enabledTestCount,
              name: config.displayName,
            })
      }
      aria-label={t('card.start.ariaLabel', { name: config.displayName })}
      className={cn(
        'px-4 py-row rounded-lg flex items-center gap-compact transition-colors',
        enabledTestCount > 0
          ? 'bg-brand-primary text-on-brand hover:bg-brand-primary'
          : 'bg-surface-base text-text-muted cursor-not-allowed',
      )}
    >
      <Play className="w-4 h-4" />
      <span className="text-sm font-medium">{t('card.action.startLabel')}</span>
    </button>
  );
}

/** Results section for module card */
function ModuleResultsSection({
  results,
  config,
}: {
  results: ModuleTestResults | null | undefined;
  config: ModuleConfig;
}): ReactElement {
  const { t } = useTranslation('modules');
  return (
    <div className="border-t border-surface-border bg-surface-base/50">
      <div className={spacing.pad.sm}>
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-2">
          {results?.testType
            ? `${results.testType} ${t('card.section.results')}`
            : t('card.section.results')}
        </div>

        {/* Frame Size Results (RFC 2544 style) */}
        {results?.frameSizeResults && results.frameSizeResults.length > 0 && (
          <FrameSizeResultsTable results={results.frameSizeResults} color={config.color} />
        )}

        {/* Service Flow Results (Y.1564 style) */}
        {results?.serviceFlowResults && results.serviceFlowResults.length > 0 && (
          <ServiceFlowResultsTable results={results.serviceFlowResults} />
        )}

        {/* OAM Results (Y.1731 style) */}
        {results?.oamResults && results.oamResults.length > 0 && (
          <OamResultsTable results={results.oamResults} />
        )}

        {/* Error message */}
        {results?.error ? (
          <div
            className={cn(
              'mt-inline pad-xs rounded-lg border',
              statusColor.bg.errorSoft,
              statusColor.border.errorSoft,
            )}
          >
            <span className={cn('text-xs', statusColor.text.error)}>{results.error}</span>
          </div>
        ) : null}

        {/* Duration */}
        {results?.duration !== undefined ? (
          <div className="mt-inline text-xs text-text-muted">
            Duration: {(results.duration / 1000).toFixed(1)}s
          </div>
        ) : null}
      </div>
    </div>
  );
}

/** Expanded test list section for module card */
function ModuleExpandedContent({
  config,
  status,
  isRunning,
  onToggleAutoStart,
  onToggleTest,
}: {
  config: ModuleConfig;
  status: ModuleStatus;
  isRunning: boolean;
  onToggleAutoStart: (enabled: boolean) => void;
  onToggleTest: (testId: string, enabled: boolean) => void;
}): ReactElement {
  const { t } = useTranslation('modules');
  return (
    <div className="border-t border-surface-border">
      {/* Auto-start Toggle */}
      <div className={cn(spacing.pad.sm, 'flex-between bg-surface-base')}>
        <div className="flex items-center gap-compact">
          <RefreshCw className={cn(iconTokens.size.sm, 'text-text-muted')} />
          <span className="text-sm text-text-secondary">{t('card.section.autoStartLabel')}</span>
        </div>
        <button
          type="button"
          onClick={(): void => onToggleAutoStart(!config.autoStart)}
          title={
            config.autoStart ? t('card.autoStart.titleEnabled') : t('card.autoStart.titleDisabled')
          }
          aria-label={
            config.autoStart
              ? t('card.autoStart.ariaLabelEnabled')
              : t('card.autoStart.ariaLabelDisabled')
          }
          className={cn(
            'w-10 h-6 rounded-full relative transition-colors',
            config.autoStart ? 'bg-brand-primary' : 'bg-surface-border',
          )}
        >
          <span
            className={cn(
              'absolute top-1 w-4 h-4 rounded-full bg-knob transition-transform',
              config.autoStart ? 'translate-x-5' : 'translate-x-1',
            )}
          />
        </button>
      </div>

      {/* Test List */}
      <div className={spacing.pad.sm}>
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-2">
          {t('card.section.tests')}
        </div>
        <div className="stack-xs">
          {config.tests.map((test) => (
            <label
              key={test.id}
              title={t('card.test.title', { description: test.description, name: test.name })}
              className={cn(
                'flex items-center gap-default pad-xs rounded-lg cursor-pointer transition-colors',
                'hover:bg-surface-hover',
                // The unchecked box already says "off"; dimming cost contrast.
                test.enabled ? '' : 'text-text-muted',
              )}
            >
              <input
                type="checkbox"
                checked={test.enabled}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  onToggleTest(test.id, e.target.checked)
                }
                aria-label={t('card.test.ariaLabel', { name: test.name })}
                className="w-4 h-4"
                style={{ accentColor: config.color }}
              />
              <div className="flex-1 min-w-0">
                <div className="label">{test.name}</div>
                <div className="text-xs text-text-muted truncate">{test.description}</div>
              </div>
              {isRunning && status.currentTest === test.id && (
                <span className={cn(statusColor.dot, statusColor.bg.success, 'animate-pulse')} />
              )}
            </label>
          ))}
        </div>
      </div>
    </div>
  );
}

export function ModuleCard({
  config,
  status,
  results,
  onToggleModule,
  onToggleAutoStart,
  onToggleTest,
  onStart,
  onStop,
  onConfigure,
}: ModuleCardProps): ReactElement {
  const { t } = useTranslation('modules');
  const [expanded, setExpanded] = useState(false);
  const enabledTestCount = config.tests.filter((t2) => t2.enabled).length;
  const isRunning = status.status === 'running' || status.status === 'starting';
  const hasResults = checkHasResults(results);
  const showResults = isRunning || status.status === 'completed' || status.status === 'error';

  return (
    <div
      className={cn(
        'border rounded-xl overflow-hidden transition-all',
        // Recessed surface rather than a layer-opacity dim, which dropped the
        // results-table headers to 2.44:1, under the 4.5:1 AA needs (#931).
        // Dim the colour token, not the layer.
        config.enabled ? 'border-surface-border' : 'border-transparent bg-surface-hover',
        'bg-surface-raised',
      )}
      style={{
        borderLeftWidth: '4px',
        borderLeftColor: config.enabled ? config.color : 'transparent',
      }}
    >
      {/* Module Header */}
      <div className={cn(spacing.pad.default, 'flex-between')}>
        <div className="flex items-center gap-default flex-1">
          {/* Enable Toggle */}
          <button
            type="button"
            onClick={(): void => onToggleModule(!config.enabled)}
            className={cn(
              'w-8 h-8 rounded-lg flex-center transition-colors',
              config.enabled ? statusColor.badge.successStrong : 'bg-surface-base text-text-muted',
            )}
            title={
              config.enabled
                ? t('card.module.titleEnabled', {
                    name: config.displayName,
                    standard: config.standard,
                  })
                : t('card.module.titleDisabled', {
                    name: config.displayName,
                    standard: config.standard,
                  })
            }
            aria-label={
              config.enabled
                ? t('card.module.ariaLabelEnabled', { name: config.displayName })
                : t('card.module.ariaLabelDisabled', { name: config.displayName })
            }
          >
            <Power className="w-4 h-4" />
          </button>

          {/* Module Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-compact">
              <span
                className="w-3 h-3 rounded-full flex-shrink-0"
                style={{ backgroundColor: config.color }}
              />
              <h3 className="font-semibold text-text-primary truncate">{config.displayName}</h3>
              <span className="text-xs px-cell py-0.5 rounded-full bg-surface-base text-text-muted">
                {config.standard}
              </span>
            </div>
            <p className="text-xs text-text-muted mt-0.5 truncate">
              {t('card.module.testsEnabledCount', {
                enabled: enabledTestCount,
                total: config.tests.length,
              })}
            </p>
          </div>

          {/* Status Indicator */}
          <ModuleStatusIndicator status={status} isRunning={isRunning} />
        </div>

        {/* Actions */}
        <div className="flex items-center gap-compact">
          {/* Configure Button */}
          <button
            type="button"
            onClick={onConfigure}
            className={cn(
              'pad-xs rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={t('card.configure.title', { name: config.displayName })}
            aria-label={t('card.configure.ariaLabel', { name: config.displayName })}
          >
            <Settings2 className="w-4 h-4" />
          </button>

          {/* Start/Stop Button */}
          <ModuleActionButton
            config={config}
            isRunning={isRunning}
            enabledTestCount={enabledTestCount}
            onStart={onStart}
            onStop={onStop}
          />

          {/* Expand Toggle */}
          <button
            type="button"
            onClick={(): void => setExpanded(!expanded)}
            className={cn(
              'pad-xs rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={
              expanded
                ? t('card.expand.titleExpanded', { name: config.displayName })
                : t('card.expand.titleCollapsed', { name: config.displayName })
            }
            aria-label={
              expanded ? t('card.expand.ariaLabelExpanded') : t('card.expand.ariaLabelCollapsed')
            }
          >
            {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>
      </div>

      {/* Test Results Section - Always visible when running or has results */}
      {config.enabled && showResults && hasResults ? (
        <ModuleResultsSection results={results} config={config} />
      ) : null}

      {/* Expanded Content - Settings and Test Selection */}
      {expanded && config.enabled ? (
        <ModuleExpandedContent
          config={config}
          status={status}
          isRunning={isRunning}
          onToggleAutoStart={onToggleAutoStart}
          onToggleTest={onToggleTest}
        />
      ) : null}
    </div>
  );
}

/** Helper to check if results have data */
function checkHasResults(results: ModuleTestResults | null | undefined): boolean {
  if (!results) {
    return false;
  }
  const hasFrameSizeResults =
    results.frameSizeResults !== null &&
    results.frameSizeResults !== undefined &&
    results.frameSizeResults.length > 0;
  const hasServiceFlowResults =
    results.serviceFlowResults !== null &&
    results.serviceFlowResults !== undefined &&
    results.serviceFlowResults.length > 0;
  const hasOamResults =
    results.oamResults !== null &&
    results.oamResults !== undefined &&
    results.oamResults.length > 0;
  return hasFrameSizeResults || hasServiceFlowResults || hasOamResults;
}

export default ModuleCard;
