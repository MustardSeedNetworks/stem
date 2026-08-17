import { RoleGuard } from '../components/RoleGuard';
import { TrafficGenConfigForm } from '../components/TrafficGenConfigForm';
import { useAppContext } from '../contexts/AppContext';

export function TrafficGenPage() {
  const { trafficGenConfig, setTrafficGenConfig, selectedTests } = useAppContext();
  return (
    <RoleGuard requires="test_master" moduleName="TrafficGen">
      <TrafficGenConfigForm
        config={trafficGenConfig}
        setConfig={setTrafficGenConfig}
        selectedTests={selectedTests}
      />
    </RoleGuard>
  );
}
