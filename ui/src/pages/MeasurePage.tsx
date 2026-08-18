import { RoleGuard } from '../components/RoleGuard';
import { Y1731ConfigForm } from '../components/Y1731ConfigForm';
import { useAppContext } from '../contexts/AppContext';

export function MeasurePage() {
  const { y1731Config, setY1731Config, selectedTests } = useAppContext();
  return (
    <RoleGuard requires="test_master" moduleName="Measure">
      <Y1731ConfigForm
        config={y1731Config}
        setConfig={setY1731Config}
        selectedTests={selectedTests}
      />
    </RoleGuard>
  );
}
