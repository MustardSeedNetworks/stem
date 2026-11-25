import { Tooltip } from '../shared/Tooltip';

/**
 * @fileoverview Stem - License Section Component
 * @description Displays license status and provides activation functionality.
 *              Supports full license activation and 14-day trial mode.
 */

import type { TFunction } from 'i18next';
import { Clock, Shield } from 'lucide-react';
import type { ReactElement } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { LICENSE_ENDPOINT } from '../contexts/LicenseContext';
import { Alert } from '../shared/Alert';
import { Button } from '../shared/Button';
import { Card } from '../shared/Card';
import { CategoryChip } from '../shared/CategoryChip';
import { FormField } from '../shared/FormField';
import { Input } from '../shared/Input';
import { StatusPill } from '../shared/StatusPill';
import { authFetch } from '../stores/auth-store';
import type { LicenseStatus } from '../types/generated/license-status';

/**
 * The wire type, generated from the Go DTO rather than re-typed here: the
 * hand-written copy was missing `platform`, `licenseKey` and `message`, and
 * made `expiresAt` required where the DTO omits a zero time (UI-STEM-15).
 */
type LicenseInfo = LicenseStatus;

/** Length of the no-key trial, in days. Interpolated into the copy so the
    number lives in one place rather than in three locale strings. */
const TRIAL_DAYS = 14;

/**
 * Reads an activation/trial reply, whichever shape the daemon sent.
 *
 * A success carries foundation's ActivationResult (`success`, `message`); a
 * rejection carries HTTPErrorResponse (`error`, `code`, `message`) or
 * ErrorResponse (`success:false`, `message`). All three name the reason in
 * `message`, so the operator sees it instead of the blanket "connection
 * failed" every failure used to collapse into. A body that is not JSON at all
 * (the SPA's index.html, a proxy's error page) has no message to show and
 * falls back to the caller's copy.
 */
async function readActivation(response: Response): Promise<{ success: boolean; message: string }> {
  let body: { success?: boolean; message?: string } = {};
  try {
    body = (await response.json()) as typeof body;
  } catch {
    // Non-JSON body — leave the caller's fallback message in place.
  }
  return {
    success: response.ok && body.success === true,
    message: body.message ?? '',
  };
}

function tierName(tier: number, t: TFunction<['common', 'settings', 'errors']>): string {
  switch (tier) {
    case 0:
      return t('settings:license.tiers.invalid');
    case 1:
      return t('settings:license.tiers.reflector');
    case 2:
      return t('settings:license.tiers.professional');
    default:
      return t('settings:license.tiers.unknown');
  }
}

interface LicenseStatusProps {
  licenseInfo: LicenseInfo;
}

function LicenseStatusBadge({ licenseInfo }: LicenseStatusProps): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  return (
    <div className="flex items-center gap-2">
      {licenseInfo.activated ? (
        <StatusPill
          tone="success"
          label={
            licenseInfo.isTrialMode
              ? t('settings:license.trialActive')
              : t('settings:license.licensed')
          }
        />
      ) : (
        <StatusPill tone="warning" label={t('settings:license.notActivated')} />
      )}
      {licenseInfo.activated ? (
        <span className="text-sm text-text-muted">{tierName(licenseInfo.tier, t)}</span>
      ) : null}
    </div>
  );
}

function LicenseDetails({ licenseInfo }: LicenseStatusProps): ReactElement | null {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  if (!licenseInfo.activated) {
    return null;
  }

  return (
    <dl className="stack-xs border-t border-hairline pt-3 text-sm">
      <div className="flex justify-between">
        <span className="text-text-muted">{t('settings:license.tier')}</span>
        <span className="font-medium">{tierName(licenseInfo.tier, t)}</span>
      </div>
      {licenseInfo.isTrialMode ? (
        <div className="flex justify-between">
          <span className="text-text-muted">{t('settings:license.daysRemaining')}</span>
          <span className="font-medium text-status-warning-strong">
            {t('settings:license.daysCount', { count: licenseInfo.daysRemaining })}
          </span>
        </div>
      ) : null}
      {!licenseInfo.isTrialMode && licenseInfo.expiresAt && (
        <div className="flex justify-between">
          <span className="text-text-muted">{t('settings:license.expiresAt')}</span>
          <span className="font-medium">
            {new Date(licenseInfo.expiresAt).toLocaleDateString()}
          </span>
        </div>
      )}
      <div className="flex justify-between">
        <span className="text-text-muted">{t('settings:license.deviceId')}</span>
        <span className="font-mono text-xs">{licenseInfo.deviceHash.slice(0, 8)}...</span>
      </div>
    </dl>
  );
}

function LicenseFeatures({ licenseInfo }: LicenseStatusProps): ReactElement | null {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  if (!licenseInfo.features || licenseInfo.features.length === 0) {
    return null;
  }

  return (
    <div>
      <div className="text-sm text-text-muted mb-2">{t('settings:license.enabledFeatures')}</div>
      <div className="flex flex-wrap gap-tight">
        {licenseInfo.features.map((feature) => (
          <CategoryChip key={feature} label={feature} />
        ))}
      </div>
    </div>
  );
}

interface ActivationFormProps {
  licenseKey: string;
  loading: boolean;
  showTrial: boolean;
  onKeyChange: (value: string) => void;
  onActivate: () => void;
  onStartTrial: () => void;
}

function ActivationForm({
  licenseKey,
  loading,
  showTrial,
  onKeyChange,
  onActivate,
  onStartTrial,
}: ActivationFormProps): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  return (
    <div className="stack border-t border-hairline pt-4">
      <FormField label={t('settings:license.enterKey')}>
        {(control) => (
          <Input
            {...control}
            type="text"
            value={licenseKey}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              onKeyChange(e.target.value.trim())
            }
            placeholder="MSN1.<payload>.<signature>"
            className="font-mono"
          />
        )}
      </FormField>

      <Tooltip text={t('settings:license.activateTooltip')}>
        <Button
          variant="primary"
          onClick={onActivate}
          loading={loading}
          disabled={!licenseKey.trim()}
          icon={<Shield className="h-4 w-4" />}
          className="w-full"
        >
          {loading ? t('settings:license.activating') : t('settings:license.activate')}
        </Button>
      </Tooltip>

      {showTrial ? (
        <Tooltip text={t('settings:license.trialTooltip', { days: TRIAL_DAYS })}>
          <Button
            onClick={onStartTrial}
            disabled={loading}
            variant="secondary"
            icon={<Clock className="h-4 w-4" />}
            className="w-full"
          >
            {t('settings:license.startTrialDays', { days: TRIAL_DAYS })}
          </Button>
        </Tooltip>
      ) : null}
    </div>
  );
}

interface MessageDisplayProps {
  error: string | null;
  success: string | null;
}

function MessageDisplay({ error, success }: MessageDisplayProps): ReactElement | null {
  if (!(error || success)) {
    return null;
  }

  return (
    <>
      {error ? <Alert tone="error">{error}</Alert> : null}
      {success ? <Alert tone="success">{success}</Alert> : null}
    </>
  );
}

export function LicenseSection(): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  const [licenseInfo, setLicenseInfo] = useState<LicenseInfo | null>(null);
  const [licenseKey, setLicenseKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const fetchLicenseStatus = useCallback(async (): Promise<void> => {
    try {
      const response = await authFetch(LICENSE_ENDPOINT);
      // An HTML body means the request reached the SPA handler rather than the
      // API, which is a 200 and would otherwise parse-error into the silent
      // catch below — the way the unversioned path hid for so long.
      const contentType = response.headers.get('content-type') ?? '';
      if (response.ok && contentType.includes('application/json')) {
        const data = await (response.json() as Promise<LicenseInfo>);
        setLicenseInfo(data);
      }
    } catch {
      // Network error - silently ignore on status check
    }
  }, []);

  useEffect(() => {
    fetchLicenseStatus().catch(() => {
      // Handle error silently
    });
  }, [fetchLicenseStatus]);

  const handleActivate = async (): Promise<void> => {
    if (!licenseKey.trim()) {
      setError(t('errors:license.keyRequired'));
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await authFetch(`${LICENSE_ENDPOINT}/activate`, {
        method: 'POST',
        body: JSON.stringify({ licenseKey }),
      });

      const data = await readActivation(response);

      if (data.success) {
        setSuccess(data.message);
        setLicenseKey('');
        fetchLicenseStatus().catch(() => {
          // Handle error silently
        });
      } else {
        setError(data.message || t('errors:license.activationFailed'));
      }
    } catch {
      setError(t('errors:network.connectionFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleStartTrial = async (): Promise<void> => {
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await authFetch(`${LICENSE_ENDPOINT}/trial`, {
        method: 'POST',
      });

      const data = await readActivation(response);

      if (data.success) {
        setSuccess(data.message);
        fetchLicenseStatus().catch(() => {
          // Handle error silently
        });
      } else {
        setError(data.message || t('errors:license.trialStartFailed'));
      }
    } catch {
      setError(t('errors:network.connectionFailed'));
    } finally {
      setLoading(false);
    }
  };

  const showActivationForm = !licenseInfo?.activated || licenseInfo?.isTrialMode;
  const showTrialButton = !licenseInfo?.activated;

  return (
    <Card title={t('settings:license.sectionTitle')} data-testid="license-section">
      <div className="stack-lg">
        {licenseInfo ? (
          <div className="stack">
            <LicenseStatusBadge licenseInfo={licenseInfo} />
            <LicenseDetails licenseInfo={licenseInfo} />
            <LicenseFeatures licenseInfo={licenseInfo} />
          </div>
        ) : (
          <div className="text-sm text-text-muted">{t('settings:license.loading')}</div>
        )}

        {showActivationForm ? (
          <ActivationForm
            licenseKey={licenseKey}
            loading={loading}
            showTrial={showTrialButton}
            onKeyChange={setLicenseKey}
            onActivate={handleActivate}
            onStartTrial={handleStartTrial}
          />
        ) : null}

        <MessageDisplay error={error} success={success} />
      </div>
    </Card>
  );
}
