import { useTranslation } from 'react-i18next';
import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RoleGuard } from '../components/RoleGuard';
import { TrafficGenConfigForm } from '../components/TrafficGenConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';
import { Card } from '../shared/Card';

const groups: TestGroup[] = ['trafficgen'];

export function TrafficGenPage() {
  const { t } = useTranslation('settings');
  const { trafficGenConfig, setTrafficGenConfig, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="TrafficGen">
      <ModuleGate path="/tests/trafficgen" preview={configured}>
        {configured ? (
          <Card title={t('testConfig.trafficgen.title')}>
            <TrafficGenConfigForm
              config={trafficGenConfig}
              setConfig={setTrafficGenConfig}
              selectedTests={selectedTests}
            />
          </Card>
        ) : (
          <ModuleEmptyState moduleName="TrafficGen" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
