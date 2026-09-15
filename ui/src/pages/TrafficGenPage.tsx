import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { RoleGuard } from '../components/RoleGuard';
import { TrafficGenConfigForm } from '../components/TrafficGenConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';

const groups: TestGroup[] = ['trafficgen'];

export function TrafficGenPage() {
  const { trafficGenConfig, setTrafficGenConfig, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="TrafficGen">
      {configured ? (
        <TrafficGenConfigForm
          config={trafficGenConfig}
          setConfig={setTrafficGenConfig}
          selectedTests={selectedTests}
        />
      ) : (
        <ModuleEmptyState moduleName="TrafficGen" />
      )}
    </RoleGuard>
  );
}
