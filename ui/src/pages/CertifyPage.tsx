import { Award } from 'lucide-react';
import { RFC2889ConfigForm } from '../components/RFC2889ConfigForm';
import { RFC6349ConfigForm } from '../components/RFC6349ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { TSNConfigForm } from '../components/TSNConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

export function CertifyPage() {
  const {
    rfc2889Config,
    setRFC2889Config,
    rfc6349Config,
    setRFC6349Config,
    tsnConfig,
    setTSNConfig,
    selectedTests,
  } = useAppContext();

  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={Award}
        title="Certify"
        description="RFC 2889 forwarding, RFC 6349 TCP throughput, and TSN time-sensitive networking certification."
        iconColorClass="text-module-certify"
      />
      <RoleGuard requires="test_master" moduleName="Certify">
        <RFC2889ConfigForm
          config={rfc2889Config}
          setConfig={setRFC2889Config}
          selectedTests={selectedTests}
        />
        <RFC6349ConfigForm
          config={rfc6349Config}
          setConfig={setRFC6349Config}
          selectedTests={selectedTests}
        />
        <TSNConfigForm config={tsnConfig} setConfig={setTSNConfig} selectedTests={selectedTests} />
      </RoleGuard>
    </section>
  );
}
