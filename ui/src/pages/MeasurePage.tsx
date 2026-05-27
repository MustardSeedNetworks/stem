import { Waves } from 'lucide-react';
import { RoleGuard } from '../components/RoleGuard';
import { Y1731ConfigForm } from '../components/Y1731ConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

export function MeasurePage() {
  const { y1731Config, setY1731Config, selectedTests } = useAppContext();
  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={Waves}
        title="Measure"
        description="Y.1731 OAM delay / loss / synthetic loss measurement."
        iconColorClass="text-module-measure"
      />
      <RoleGuard requires="test_master" moduleName="Measure">
        <Y1731ConfigForm
          config={y1731Config}
          setConfig={setY1731Config}
          selectedTests={selectedTests}
        />
      </RoleGuard>
    </section>
  );
}
