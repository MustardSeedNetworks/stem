/**
 * SettingsDrawer Component
 *
 * Slimmed configuration panel for The Stem.
 *
 * After the #66 redesign the drawer holds only app-global concerns:
 *   - License management
 *   - Test-suite configuration (RFC 2544, Y.1564, etc.)
 *
 * Role selection moved to the header RoleChip. Interface selection
 * moved to the top-bar HeaderInterfaceSelector. Reflector profile
 * moved to the Reflector page. Uses theme tokens for consistent
 * styling.
 */

import { Grid, List, X } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRole } from '../contexts/RoleContext';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { cn, radius, spacing } from '../styles/theme';
import { LicenseSection } from './LicenseSection';
import { ModuleSelector } from './ModuleSelector';
import type { RFC2544Config } from './RFC2544ConfigForm';
import type { RFC2889Config } from './RFC2889ConfigForm';
import type { RFC6349Config } from './RFC6349ConfigForm';
import {
  MEFSection,
  RFC2544Section,
  RFC2889Section,
  RFC6349Section,
  TrafficGenSection,
  TSNSection,
  Y1564Section,
  Y1731Section,
} from './settings';
import type { TrafficGenConfig } from './TrafficGenConfigForm';
import type { TSNConfig } from './TSNConfigForm';
import type { Y1564Config } from './Y1564ConfigForm';
import type { Y1731Config } from './Y1731ConfigForm';

interface SettingsDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  selectedTests: string[];
  setSelectedTests: (tests: string[]) => void;
  rfc2544Config: RFC2544Config;
  setRFC2544Config: (config: RFC2544Config) => void;
  rfc2889Config: RFC2889Config;
  setRFC2889Config: (config: RFC2889Config) => void;
  rfc6349Config: RFC6349Config;
  setRFC6349Config: (config: RFC6349Config) => void;
  y1564Config: Y1564Config;
  setY1564Config: (config: Y1564Config) => void;
  y1731Config: Y1731Config;
  setY1731Config: (config: Y1731Config) => void;
  tsnConfig: TSNConfig;
  setTSNConfig: (config: TSNConfig) => void;
  trafficGenConfig: TrafficGenConfig;
  setTrafficGenConfig: (config: TrafficGenConfig) => void;
}

type ViewMode = 'standard' | 'module';

export function SettingsDrawer({
  isOpen,
  onClose,
  selectedTests,
  setSelectedTests,
  rfc2544Config,
  setRFC2544Config,
  rfc2889Config,
  setRFC2889Config,
  rfc6349Config,
  setRFC6349Config,
  y1564Config,
  setY1564Config,
  y1731Config,
  setY1731Config,
  tsnConfig,
  setTSNConfig,
  trafficGenConfig,
  setTrafficGenConfig,
}: SettingsDrawerProps): React.ReactElement | null {
  const { t } = useTranslation();
  const { role } = useRole();
  const [viewMode, setViewMode] = useState<ViewMode>('standard');

  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: isOpen,
    onEscape: onClose,
  });

  const toggleTest = useCallback(
    (testId: string): void => {
      if (selectedTests.includes(testId)) {
        setSelectedTests(selectedTests.filter((currentTest) => currentTest !== testId));
      } else {
        setSelectedTests([...selectedTests, testId]);
      }
    },
    [selectedTests, setSelectedTests],
  );

  if (!isOpen) {
    return null;
  }

  return (
    <>
      {/* Backdrop */}
      <button
        type="button"
        className="fixed inset-0 bg-scrim/50 z-40 cursor-default"
        onClick={onClose}
        title={t('accessibility.closeSettings', 'Close settings drawer')}
        aria-label={t('accessibility.closeSettings', 'Close settings drawer')}
      />

      {/* Drawer */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-drawer-title"
        data-testid="settings-drawer"
        className="fixed right-0 top-0 h-full w-96 max-w-full bg-surface-raised border-l border-surface-border z-50 overflow-y-auto animate-slide-in"
      >
        {/* Header */}
        <div className="sticky top-0 bg-surface-raised border-b border-surface-border px-4 py-row-lg flex-between">
          <h2 id="settings-drawer-title" className="heading-3 text-text-primary">
            {t('settings:title', 'Settings')}
          </h2>
          <button
            type="button"
            onClick={onClose}
            data-testid="settings-drawer-close"
            className={cn(spacing.pad.sm, 'hover:bg-surface-hover', radius.lg, 'transition-colors')}
            title={t('accessibility.closeSettings', 'Close settings')}
            aria-label={t('accessibility.closeSettings', 'Close settings')}
          >
            <X className="w-5 h-5 text-text-muted" aria-hidden="true" />
          </button>
        </div>

        {/* Content */}
        <div className={cn(spacing.pad.default, 'stack-lg')}>
          {/* License Section */}
          <LicenseSection />

          {/* Test-Master test selection — only relevant when this stem
              is configured as a Test Master. Reflector role gets its
              picker on the Reflector page. */}
          {role === 'test_master' && (
            <>
              {/* View Toggle */}
              <ViewToggle viewMode={viewMode} onViewModeChange={setViewMode} />

              {/* Module View */}
              {viewMode === 'module' && (
                <ModuleSelector selectedTests={selectedTests} setSelectedTests={setSelectedTests} />
              )}

              {/* Standard View - Tests by Standard */}
              {viewMode === 'standard' && (
                <>
                  <RFC2544Section
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={rfc2544Config}
                    onConfigChange={setRFC2544Config}
                  />

                  <Y1564Section
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={y1564Config}
                    onConfigChange={setY1564Config}
                  />

                  <RFC2889Section
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={rfc2889Config}
                    onConfigChange={setRFC2889Config}
                  />

                  <RFC6349Section
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={rfc6349Config}
                    onConfigChange={setRFC6349Config}
                  />

                  <Y1731Section
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={y1731Config}
                    onConfigChange={setY1731Config}
                  />

                  <MEFSection selectedTests={selectedTests} onToggleTest={toggleTest} />

                  <TSNSection
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={tsnConfig}
                    onConfigChange={setTSNConfig}
                  />

                  <TrafficGenSection
                    selectedTests={selectedTests}
                    onToggleTest={toggleTest}
                    config={trafficGenConfig}
                    onConfigChange={setTrafficGenConfig}
                  />
                </>
              )}
            </>
          )}
        </div>
      </div>
    </>
  );
}

// =============================================================================
// View Toggle Sub-component
// =============================================================================

interface ViewToggleProps {
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
}

function ViewToggle({ viewMode, onViewModeChange }: ViewToggleProps): React.ReactElement {
  const { t } = useTranslation();

  return (
    <div className={cn('flex-between', spacing.pad.sm, 'bg-surface-base', radius.lg, 'mb-2')}>
      <span className="body-small text-text-muted">{t('settings:viewBy', 'View by')}:</span>
      <div className={cn('flex', radius.lg, 'overflow-hidden border border-surface-border')}>
        <button
          type="button"
          onClick={(): void => onViewModeChange('standard')}
          title={t(
            'settings:viewStandardHint',
            'Group settings by configuration section (mode, interface, thresholds)',
          )}
          className={cn(
            'flex items-center gap-tight px-3 py-compact-md caption',
            viewMode === 'standard'
              ? 'bg-brand-primary text-text-inverse'
              : 'bg-surface-raised text-text-muted hover:bg-surface-hover',
          )}
        >
          <List className="w-3 h-3" aria-hidden="true" />
          {t('settings:viewStandard', 'Standard')}
        </button>
        <button
          type="button"
          onClick={(): void => onViewModeChange('module')}
          title={t(
            'settings:viewModuleHint',
            'Group settings by test module (Benchmark, ServiceTest, TrafficGen, Measure, Certify)',
          )}
          className={cn(
            'flex items-center gap-tight px-3 py-compact-md caption',
            viewMode === 'module'
              ? 'bg-brand-primary text-text-inverse'
              : 'bg-surface-raised text-text-muted hover:bg-surface-hover',
          )}
        >
          <Grid className="w-3 h-3" aria-hidden="true" />
          {t('settings:viewModule', 'Module')}
        </button>
      </div>
    </div>
  );
}

export default SettingsDrawer;
