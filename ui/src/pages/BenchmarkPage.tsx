import { RFC2544ConfigForm } from '../components/RFC2544ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { useAppContext } from '../contexts/AppContext';

export function BenchmarkPage() {
  const { rfc2544Config, setRFC2544Config, selectedTests } = useAppContext();
  return (
    <RoleGuard requires="test_master" moduleName="Benchmark">
      <RFC2544ConfigForm
        config={rfc2544Config}
        setConfig={setRFC2544Config}
        selectedTests={selectedTests}
      />
    </RoleGuard>
  );
}
