import { Zap } from 'lucide-react';
import { RoleGuard } from '../components/RoleGuard';
import { TrafficGenConfigForm } from '../components/TrafficGenConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

export function TrafficGenPage() {
  const { trafficGenConfig, setTrafficGenConfig, selectedTests } = useAppContext();
  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={Zap}
        title="TrafficGen"
        description="Custom traffic generation — shape streams, drive load, and validate paths."
        iconColorClass="text-module-trafficgen"
      />
      <RoleGuard requires="test_master" moduleName="TrafficGen">
        <TrafficGenConfigForm
          config={trafficGenConfig}
          setConfig={setTrafficGenConfig}
          selectedTests={selectedTests}
        />
      </RoleGuard>
    </section>
  );
}
