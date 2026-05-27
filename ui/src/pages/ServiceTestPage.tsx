import { Settings2 } from 'lucide-react';
import { RoleGuard } from '../components/RoleGuard';
import { Y1564ConfigForm } from '../components/Y1564ConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

export function ServiceTestPage() {
  const { y1564Config, setY1564Config, selectedTests } = useAppContext();
  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={Settings2}
        title="ServiceTest"
        description="Y.1564 / MEF service activation and performance verification."
        iconColorClass="text-module-servicetest"
      />
      <RoleGuard requires="test_master" moduleName="ServiceTest">
        <Y1564ConfigForm
          config={y1564Config}
          setConfig={setY1564Config}
          selectedTests={selectedTests}
        />
      </RoleGuard>
    </section>
  );
}
