import { useTranslation } from 'react-i18next';
import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RFC2544ConfigForm } from '../components/RFC2544ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';
import { Card } from '../shared/Card';

const groups: TestGroup[] = ['rfc2544'];

export function BenchmarkPage() {
  const { t } = useTranslation('settings');
  const { rfc2544Config, setRFC2544Config, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="Benchmark">
      <ModuleGate path="/tests/benchmark" preview={configured}>
        {configured ? (
          <Card title={t('testConfig.rfc2544.title')}>
            <RFC2544ConfigForm
              config={rfc2544Config}
              setConfig={setRFC2544Config}
              selectedTests={selectedTests}
            />
          </Card>
        ) : (
          <ModuleEmptyState moduleName="Benchmark" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
