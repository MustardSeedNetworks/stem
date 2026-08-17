import { RoleGuard } from '../components/RoleGuard';
import { Y1564ConfigForm } from '../components/Y1564ConfigForm';
import { useAppContext } from '../contexts/AppContext';

export function ServiceTestPage() {
  const { y1564Config, setY1564Config, selectedTests } = useAppContext();
  return (
    <RoleGuard requires="test_master" moduleName="ServiceTest">
      <Y1564ConfigForm
        config={y1564Config}
        setConfig={setY1564Config}
        selectedTests={selectedTests}
      />
    </RoleGuard>
  );
}
