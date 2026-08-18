import { RFC2889ConfigForm } from '../components/RFC2889ConfigForm';
import { RFC6349ConfigForm } from '../components/RFC6349ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { TSNConfigForm } from '../components/TSNConfigForm';
import { useAppContext } from '../contexts/AppContext';

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
  );
}
