/**
 * @fileoverview The Stem - License Section Component
 * @description Displays license status and provides activation functionality.
 *              Supports full license activation and 14-day trial mode.
 */

import { AlertTriangle, CheckCircle, Clock, Key, Loader2, Shield } from 'lucide-react';
import type { ReactElement } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from './CollapsibleSection';

interface LicenseInfo {
  activated: boolean;
  tier: number;
  tierName: string;
  isTrialMode: boolean;
  daysRemaining: number;
  features: string[];
  deviceHash: string;
  expiresAt: string;
}

/** Length of the no-key trial, in days. Interpolated into the copy so the
    number lives in one place rather than in three locale strings. */
const TRIAL_DAYS = 14;

const tierNames: Record<number, string> = {
  0: 'Invalid',
  1: 'Reflector',
  2: 'Test Suite',
  3: 'Enterprise',
};

function formatDate(dateStr: string): string {
  if (!dateStr) {
    return 'N/A';
  }
  const date = new Date(dateStr);
  return date.toLocaleDateString();
}

interface LicenseStatusProps {
  licenseInfo: LicenseInfo;
}

function LicenseStatusBadge({ licenseInfo }: LicenseStatusProps): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  return (
    <div className="flex items-center gap-compact">
      {licenseInfo.activated ? (
        <span className="status-badge success">
          <CheckCircle className="w-3 h-3" />
          {licenseInfo.isTrialMode ? 'Trial Active' : 'Licensed'}
        </span>
      ) : (
        <span className="status-badge warning">
          <AlertTriangle className="w-3 h-3" />
          {t('settings:license.notActivated')}
        </span>
      )}
      {licenseInfo.activated ? (
        <span className="text-sm text-text-muted">{tierNames[licenseInfo.tier] || 'Unknown'}</span>
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
    <div className="bg-surface-hover rounded-md pad-sm text-sm stack-xs">
      <div className="flex justify-between">
        <span className="text-text-muted">Tier</span>
        <span className="font-medium">{tierNames[licenseInfo.tier]}</span>
      </div>
      {licenseInfo.isTrialMode ? (
        <div className="flex justify-between">
          <span className="text-text-muted">{t('settings:license.daysRemaining')}</span>
          <span className="font-medium text-status-warning">{licenseInfo.daysRemaining} days</span>
        </div>
      ) : null}
      {!licenseInfo.isTrialMode && licenseInfo.expiresAt && (
        <div className="flex justify-between">
          <span className="text-text-muted">Expires</span>
          <span className="font-medium">{formatDate(licenseInfo.expiresAt)}</span>
        </div>
      )}
      <div className="flex justify-between">
        <span className="text-text-muted">{t('settings:license.deviceId')}</span>
        <span className="font-mono text-xs">{licenseInfo.deviceHash.slice(0, 8)}...</span>
      </div>
    </div>
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
          <span
            key={feature}
            className="px-cell py-0.5 text-xs bg-brand-primary/10 text-brand-primary rounded-full"
          >
            {feature}
          </span>
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
    <div className="border-t border-surface-border pt-section stack">
      <div className="text-sm font-medium">{t('settings:license.activate')}</div>

      <div>
        <input
          type="text"
          value={licenseKey}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
            onKeyChange(e.target.value.toUpperCase())
          }
          placeholder="XXXX-XXXX-XXXX-XXXX"
          className="font-mono text-center tracking-wider"
          maxLength={19}
        />
      </div>

      <button
        type="button"
        onClick={onActivate}
        disabled={loading || !licenseKey.trim()}
        title={t('settings:license.activateTooltip')}
        className="btn btn-primary w-full"
      >
        {loading ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" /> {t('settings:license.activating')}
          </>
        ) : (
          <>
            <Shield className="w-4 h-4" /> {t('settings:license.activate')}
          </>
        )}
      </button>

      {showTrial ? (
        <button
          type="button"
          onClick={onStartTrial}
          disabled={loading}
          title={t('settings:license.trialTooltip', { days: TRIAL_DAYS })}
          className="btn btn-secondary w-full"
        >
          <Clock className="w-4 h-4" />
          {t('settings:license.startTrialDays', { days: TRIAL_DAYS })}
        </button>
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
      {error ? (
        <div className="text-sm text-status-error bg-status-error/10 pad-xs rounded">{error}</div>
      ) : null}
      {success ? (
        <div className="text-sm text-status-success bg-status-success/10 pad-xs rounded">
          {success}
        </div>
      ) : null}
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
      const response = await fetch('/api/license');
      if (response.ok) {
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
      const response = await fetch('/api/license/activate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: licenseKey }),
      });

      const data = await (response.json() as Promise<{ success: boolean; message: string }>);

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
      const response = await fetch('/api/license/trial', {
        method: 'POST',
      });

      const data = await (response.json() as Promise<{ success: boolean; message: string }>);

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
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Key className="w-4 h-4" />
          <span>{t('settings:license.sectionTitle')}</span>
        </div>
      }
      defaultOpen={true}
    >
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
    </CollapsibleSection>
  );
}
