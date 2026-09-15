import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { RoleGuard } from '../components/RoleGuard';
import { Y1564ConfigForm } from '../components/Y1564ConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';

const groups: TestGroup[] = ['y1564'];

export function ServiceTestPage() {
  const { y1564Config, setY1564Config, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="ServiceTest">
      {configured ? (
        <Y1564ConfigForm
          config={y1564Config}
          setConfig={setY1564Config}
          selectedTests={selectedTests}
        />
      ) : (
        <ModuleEmptyState moduleName="ServiceTest" />
      )}
    </RoleGuard>
  );
}
