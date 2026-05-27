import { BarChart3 } from 'lucide-react';
import { RFC2544ConfigForm } from '../components/RFC2544ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { useAppContext } from '../contexts/AppContext';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

export function BenchmarkPage() {
  const { rfc2544Config, setRFC2544Config, selectedTests } = useAppContext();
  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={BarChart3}
        title="Benchmark"
        description="RFC 2544 throughput, latency, frame-loss, and back-to-back measurements."
        iconColorClass="text-module-benchmark"
      />
      <RoleGuard requires="test_master" moduleName="Benchmark">
        <RFC2544ConfigForm
          config={rfc2544Config}
          setConfig={setRFC2544Config}
          selectedTests={selectedTests}
        />
      </RoleGuard>
    </section>
  );
}
