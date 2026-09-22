import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RFC2889ConfigForm } from '../components/RFC2889ConfigForm';
import { RFC6349ConfigForm } from '../components/RFC6349ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { TSNConfigForm } from '../components/TSNConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';

// Certify covers three standards; any one of them selected means the page has
// a form to show, and the forms themselves drop out individually.
const groups: TestGroup[] = ['rfc2889', 'rfc6349', 'tsn'];

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
  const configured = hasAnyGroupTests(groups, selectedTests);

  return (
    <RoleGuard requires="test_master" moduleName="Certify">
      <ModuleGate path="/tests/certify" preview={configured}>
        {configured ? (
          <>
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
            <TSNConfigForm
              config={tsnConfig}
              setConfig={setTSNConfig}
              selectedTests={selectedTests}
            />
          </>
        ) : (
          <ModuleEmptyState moduleName="Certify" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
