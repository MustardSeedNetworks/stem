import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RFC2544ConfigForm } from '../components/RFC2544ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';

const groups: TestGroup[] = ['rfc2544'];

export function BenchmarkPage() {
  const { rfc2544Config, setRFC2544Config, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="Benchmark">
      <ModuleGate path="/tests/benchmark" preview={configured}>
        {configured ? (
          <RFC2544ConfigForm
            config={rfc2544Config}
            setConfig={setRFC2544Config}
            selectedTests={selectedTests}
          />
        ) : (
          <ModuleEmptyState moduleName="Benchmark" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
