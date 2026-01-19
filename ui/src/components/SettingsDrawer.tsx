// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * SettingsDrawer Component
 *
 * Main configuration panel for The Stem test suite.
 * Refactored to use modular section components for maintainability.
 *
 * Features:
 * - Operating mode selection (Reflector / Test Master)
 * - Network interface selection
 * - Test suite configuration (RFC 2544, Y.1564, etc.)
 * - Reflector profile selection
 * - License management
 *
 * Uses theme tokens for consistent styling.
 */

import { Grid, List, X } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { cn, radius, spacing } from '../styles/theme';
import { LicenseSection } from './LicenseSection';
import { ModuleSelector } from './ModuleSelector';
import type { RFC2544Config } from './RFC2544ConfigForm';
import type { RFC2889Config } from './Rfc2889ConfigForm';
import type { RFC6349Config } from './Rfc6349ConfigForm';
import {
  InterfaceSection,
  MEFSection,
  ModeSection,
  ReflectorSection,
  RFC2544Section,
  RFC2889Section,
  RFC6349Section,
  TrafficGenSection,
  TSNSection,
  Y1564Section,
  Y1731Section,
} from './settings';
import type { InterfaceInfo, OperatingMode, ReflectorProfile } from './settings/types';
import type { TrafficGenConfig } from './TrafficGenConfigForm';
import type { TSNConfig } from './TsnConfigForm';
import type { Y1564Config } from './Y1564ConfigForm';
import type { Y1731Config } from './Y1731ConfigForm';

interface SettingsDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  mode: OperatingMode;
  setMode: (mode: OperatingMode) => void;
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  setSelectedInterface: (iface: string) => void;
  selectedTests: string[];
  setSelectedTests: (tests: string[]) => void;
  reflectorProfile: ReflectorProfile;
  setReflectorProfile: (profile: ReflectorProfile) => void;
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
  mode,
  setMode,
  interfaces,
  selectedInterface,
  setSelectedInterface,
  selectedTests,
  setSelectedTests,
  reflectorProfile,
  setReflectorProfile,
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
}: SettingsDrawerProps) {
  const { t } = useTranslation();
  const [viewMode, setViewMode] = useState<ViewMode>('standard');

  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: isOpen,
    onEscape: onClose,
  });

  const toggleTest = useCallback(
    (testId: string) => {
      if (selectedTests.includes(testId)) {
        setSelectedTests(selectedTests.filter((t) => t !== testId));
      } else {
        setSelectedTests([...selectedTests, testId]);
      }
    },
    [selectedTests, setSelectedTests],
  );

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <button
        type="button"
        className="fixed inset-0 bg-black/50 z-40 cursor-default"
        onClick={onClose}
        aria-label={t('accessibility.closeSettings', 'Close settings drawer')}
      />

      {/* Drawer */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label={t('settings.title', 'Settings')}
        className="fixed right-0 top-0 h-full w-96 max-w-full bg-surface-raised border-l border-surface-border z-50 overflow-y-auto"
      >
        {/* Header */}
        <div className="sticky top-0 bg-surface-raised border-b border-surface-border px-4 py-3 flex items-center justify-between">
          <h2 className="heading-3 text-text-primary">{t('settings.title', 'Settings')}</h2>
          <button
            type="button"
            onClick={onClose}
            className={cn(spacing.pad.sm, 'hover:bg-surface-hover', radius.lg, 'transition-colors')}
            aria-label={t('accessibility.closeSettings', 'Close settings')}
          >
            <X className="w-5 h-5 text-text-muted" aria-hidden="true" />
          </button>
        </div>

        {/* Content */}
        <div className={cn(spacing.pad.default, 'space-y-4')}>
          {/* License Section */}
          <LicenseSection />

          {/* Mode Selection */}
          <ModeSection mode={mode} onModeChange={setMode} />

          {/* Interface Selection */}
          <InterfaceSection
            interfaces={interfaces}
            selectedInterface={selectedInterface}
            onInterfaceChange={setSelectedInterface}
          />

          {/* Test Master Mode */}
          {mode === 'test_master' && (
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

          {/* Reflector Mode */}
          {mode === 'reflector' && (
            <ReflectorSection profile={reflectorProfile} onProfileChange={setReflectorProfile} />
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

function ViewToggle({ viewMode, onViewModeChange }: ViewToggleProps) {
  const { t } = useTranslation();

  return (
    <div
      className={cn(
        'flex items-center justify-between',
        spacing.pad.sm,
        'bg-surface-base',
        radius.lg,
        'mb-2',
      )}
    >
      <span className="body-small text-text-muted">{t('settings.viewBy', 'View by')}:</span>
      <div className={cn('flex', radius.lg, 'overflow-hidden border border-surface-border')}>
        <button
          type="button"
          onClick={() => onViewModeChange('standard')}
          className={cn(
            'flex items-center gap-1 px-3 py-1.5 caption',
            viewMode === 'standard'
              ? 'bg-brand-primary text-text-inverse'
              : 'bg-surface-raised text-text-muted hover:bg-surface-hover',
          )}
        >
          <List className="w-3 h-3" aria-hidden="true" />
          {t('settings.viewStandard', 'Standard')}
        </button>
        <button
          type="button"
          onClick={() => onViewModeChange('module')}
          className={cn(
            'flex items-center gap-1 px-3 py-1.5 caption',
            viewMode === 'module'
              ? 'bg-brand-primary text-text-inverse'
              : 'bg-surface-raised text-text-muted hover:bg-surface-hover',
          )}
        >
          <Grid className="w-3 h-3" aria-hidden="true" />
          {t('settings.viewModule', 'Module')}
        </button>
      </div>
    </div>
  );
}

export default SettingsDrawer;
