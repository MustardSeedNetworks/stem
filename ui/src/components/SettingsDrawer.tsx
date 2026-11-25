/**
 * SettingsDrawer Component
 *
 * Configuration panel for Stem: the shared RightDrawer with SettingsTabs.
 *
 * The drawer holds app-global concerns, one tab each:
 *   - General: theme, language, automatic updates (ADR-0012)
 *   - License: license management
 *   - Tests: test-suite configuration (RFC 2544, Y.1564, etc.), shown only to
 *     a Test Master. A Reflector stem picks its profile on the Reflector page.
 *
 * Role selection lives in the header RoleChip, interface selection in the
 * HeaderInterfaceSelector.
 */

import { Cog, FlaskConical, KeyRound } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useRole } from '../contexts/RoleContext';
import { RightDrawer } from '../shared/RightDrawer';
import { SegmentedControl } from '../shared/SegmentedControl';
import { SettingsTabs } from '../shared/SettingsTabs';
import { LicenseSection } from './LicenseSection';
import { ModuleSelector } from './ModuleSelector';
import type { RFC2544Config } from './RFC2544ConfigForm';
import type { RFC2889Config } from './RFC2889ConfigForm';
import type { RFC6349Config } from './RFC6349ConfigForm';
import {
  AppearanceSection,
  LanguageSection,
  MEFSection,
  RFC2544Section,
  RFC2889Section,
  RFC6349Section,
  TrafficGenSection,
  TSNSection,
  UpdateSection,
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

type TabId = 'general' | 'license' | 'tests';
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
}: SettingsDrawerProps): React.ReactElement {
  const { t } = useTranslation(['common', 'settings']);
  const { role } = useRole();
  const [tab, setTab] = useState<TabId>('general');
  const [viewMode, setViewMode] = useState<ViewMode>('standard');

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

  const isTestMaster = role === 'test_master';
  const tabs = [
    { id: 'general', label: t('settings:tabs.general'), icon: <Cog className="h-4 w-4" /> },
    { id: 'license', label: t('settings:tabs.license'), icon: <KeyRound className="h-4 w-4" /> },
    ...(isTestMaster
      ? [
          {
            id: 'tests',
            label: t('settings:tabs.tests'),
            icon: <FlaskConical className="h-4 w-4" />,
          },
        ]
      : []),
  ];

  return (
    <RightDrawer
      open={isOpen}
      onClose={onClose}
      title={t('settings:title')}
      closeLabel={t('accessibility.closeSettings')}
      data-testid="settings-drawer"
    >
      <SettingsTabs
        tabs={tabs}
        value={tab}
        onChange={(id): void => setTab(id as TabId)}
        label={t('settings:tabs.label')}
      >
        <div className="stack-lg">
          {tab === 'general' ? (
            <>
              <AppearanceSection />
              <LanguageSection />
              <UpdateSection />
            </>
          ) : null}

          {tab === 'license' ? <LicenseSection /> : null}

          {tab === 'tests' && isTestMaster ? (
            <>
              <SegmentedControl
                label={t('settings:viewBy')}
                value={viewMode}
                onChange={(value): void => setViewMode(value as ViewMode)}
                options={[
                  { value: 'standard', label: t('settings:viewStandard') },
                  { value: 'module', label: t('settings:viewModule') },
                ]}
              />

              {viewMode === 'module' ? (
                <ModuleSelector selectedTests={selectedTests} setSelectedTests={setSelectedTests} />
              ) : (
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
          ) : null}
        </div>
      </SettingsTabs>
    </RightDrawer>
  );
}

export default SettingsDrawer;
