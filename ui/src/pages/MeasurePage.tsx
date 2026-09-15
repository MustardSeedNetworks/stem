import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { RoleGuard } from '../components/RoleGuard';
import { Y1731ConfigForm } from '../components/Y1731ConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';

const groups: TestGroup[] = ['y1731'];

export function MeasurePage() {
  const { y1731Config, setY1731Config, selectedTests } = useAppContext();
  const configured = hasAnyGroupTests(groups, selectedTests);
  return (
    <RoleGuard requires="test_master" moduleName="Measure">
      {configured ? (
        <Y1731ConfigForm
          config={y1731Config}
          setConfig={setY1731Config}
          selectedTests={selectedTests}
        />
      ) : (
        <ModuleEmptyState moduleName="Measure" />
      )}
    </RoleGuard>
  );
}
