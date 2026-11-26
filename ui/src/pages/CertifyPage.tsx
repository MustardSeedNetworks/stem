import { useTranslation } from 'react-i18next';
import { ModuleEmptyState } from '../components/ModuleEmptyState';
import { ModuleGate } from '../components/ModuleGate';
import { RFC2889ConfigForm } from '../components/RFC2889ConfigForm';
import { RFC6349ConfigForm } from '../components/RFC6349ConfigForm';
import { RoleGuard } from '../components/RoleGuard';
import { TSNConfigForm } from '../components/TSNConfigForm';
import { useAppContext } from '../contexts/AppContext';
import { hasAnyGroupTests, type TestGroup } from '../lib/testGroups';
import { Card } from '../shared/Card';

// Certify covers three standards; any one of them selected means the page has
// a form to show, and the forms themselves drop out individually.
const groups: TestGroup[] = ['rfc2889', 'rfc6349', 'tsn'];

export function CertifyPage() {
  const { t } = useTranslation('settings');
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
            {hasAnyGroupTests(['rfc2889'], selectedTests) ? (
              <Card title={t('testConfig.rfc2889.title')}>
                <RFC2889ConfigForm
                  config={rfc2889Config}
                  setConfig={setRFC2889Config}
                  selectedTests={selectedTests}
                />
              </Card>
            ) : null}
            {hasAnyGroupTests(['rfc6349'], selectedTests) ? (
              <Card title={t('testConfig.rfc6349.title')}>
                <RFC6349ConfigForm
                  config={rfc6349Config}
                  setConfig={setRFC6349Config}
                  selectedTests={selectedTests}
                />
              </Card>
            ) : null}
            {hasAnyGroupTests(['tsn'], selectedTests) ? (
              <Card title={t('testConfig.tsn.title')}>
                <TSNConfigForm
                  config={tsnConfig}
                  setConfig={setTSNConfig}
                  selectedTests={selectedTests}
                />
              </Card>
            ) : null}
          </>
        ) : (
          <ModuleEmptyState moduleName="Certify" />
        )}
      </ModuleGate>
    </RoleGuard>
  );
}
