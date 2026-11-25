import { useTranslation } from 'react-i18next';
import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RoleGuard } from '../components/RoleGuard';
import { Y1731ConfigForm } from '../components/Y1731ConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';
import { Card } from '../shared/Card';

const groups: TestGroup[] = ['y1731'];

export function MeasurePage() {
  const { t } = useTranslation('settings');
  const { y1731Config, setY1731Config, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="Measure">
      <ModuleGate path="/tests/measure" preview={configured}>
        {configured ? (
          <Card title={t('testConfig.y1731.title')}>
            <Y1731ConfigForm
              config={y1731Config}
              setConfig={setY1731Config}
              selectedTests={selectedTests}
            />
          </Card>
        ) : (
          <ModuleEmptyState moduleName="Measure" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
