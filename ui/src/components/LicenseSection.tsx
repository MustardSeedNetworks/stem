/**
 * @fileoverview The Stem - License Section Component
 * @description Displays license status and provides activation functionality.
 *              Supports full license activation and 14-day trial mode.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
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
  return (
    <div className="flex items-center gap-2">
      {licenseInfo.activated ? (
        <span className="status-badge success">
          <CheckCircle className="w-3 h-3" />
          {licenseInfo.isTrialMode ? 'Trial Active' : 'Licensed'}
        </span>
      ) : (
        <span className="status-badge warning">
          <AlertTriangle className="w-3 h-3" />
          Not Activated
        </span>
      )}
      {licenseInfo.activated ? (
        <span className="text-sm text-[var(--color-text-muted)]">
          {tierNames[licenseInfo.tier] || 'Unknown'}
        </span>
      ) : null}
    </div>
  );
}

function LicenseDetails({ licenseInfo }: LicenseStatusProps): ReactElement | null {
  if (!licenseInfo.activated) {
    return null;
  }

  return (
    <div className="bg-[var(--color-surface-hover)] rounded-md p-3 text-sm space-y-1">
      <div className="flex justify-between">
        <span className="text-[var(--color-text-muted)]">Tier</span>
        <span className="font-medium">{tierNames[licenseInfo.tier]}</span>
      </div>
      {licenseInfo.isTrialMode ? (
        <div className="flex justify-between">
          <span className="text-[var(--color-text-muted)]">Days Remaining</span>
          <span className="font-medium text-[var(--color-status-warning)]">
            {licenseInfo.daysRemaining} days
          </span>
        </div>
      ) : null}
      {!licenseInfo.isTrialMode && licenseInfo.expiresAt && (
        <div className="flex justify-between">
          <span className="text-[var(--color-text-muted)]">Expires</span>
          <span className="font-medium">{formatDate(licenseInfo.expiresAt)}</span>
        </div>
      )}
      <div className="flex justify-between">
        <span className="text-[var(--color-text-muted)]">Device ID</span>
        <span className="font-mono text-xs">{licenseInfo.deviceHash.slice(0, 8)}...</span>
      </div>
    </div>
  );
}

function LicenseFeatures({ licenseInfo }: LicenseStatusProps): ReactElement | null {
  if (!licenseInfo.features || licenseInfo.features.length === 0) {
    return null;
  }

  return (
    <div>
      <div className="text-sm text-[var(--color-text-muted)] mb-2">Enabled Features</div>
      <div className="flex flex-wrap gap-1">
        {licenseInfo.features.map((feature) => (
          <span
            key={feature}
            className="px-2 py-0.5 text-xs bg-brand-primary/10 text-brand-primary rounded-full"
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
  return (
    <div className="border-t border-[var(--color-surface-border)] pt-4 space-y-3">
      <div className="text-sm font-medium">Activate License</div>

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
        className="btn btn-primary w-full"
      >
        {loading ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" /> Activating...
          </>
        ) : (
          <>
            <Shield className="w-4 h-4" /> Activate License
          </>
        )}
      </button>

      {showTrial ? (
        <button
          type="button"
          onClick={onStartTrial}
          disabled={loading}
          className="btn btn-secondary w-full"
        >
          <Clock className="w-4 h-4" />
          {t('common:startTrial')}
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
        <div className="text-sm text-[var(--color-status-error)] bg-red-500 bg-opacity-10 p-2 rounded">
          {error}
        </div>
      ) : null}
      {success ? (
        <div className="text-sm text-[var(--color-status-success)] bg-green-500 bg-opacity-10 p-2 rounded">
          {success}
        </div>
      ) : null}
    </>
  );
}

export function LicenseSection(): ReactElement {
  const { t } = useTranslation();
  const [licenseInfo, setLicenseInfo] = useState<LicenseInfo | null>(null);
  const [licenseKey, setLicenseKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const fetchLicenseStatus = useCallback(async () => {
    try {
      const response = await fetch('/api/license');
      if (response.ok) {
        const data = response.json();
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
      setError('Please enter a license key');
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

      const data = response.json();

      if (data.success) {
        setSuccess(data.message);
        setLicenseKey('');
        fetchLicenseStatus().catch(() => {
          // Handle error silently
        });
      } else {
        setError(data.message || 'Activation failed');
      }
    } catch {
      setError('Failed to connect to server');
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

      const data = response.json();

      if (data.success) {
        setSuccess(data.message);
        fetchLicenseStatus().catch(() => {
          // Handle error silently
        });
      } else {
        setError(data.message || 'Failed to start trial');
      }
    } catch {
      setError('Failed to connect to server');
    } finally {
      setLoading(false);
    }
  };

  const showActivationForm = !licenseInfo?.activated || licenseInfo?.isTrialMode;
  const showTrialButton = !licenseInfo?.activated;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Key className="w-4 h-4" />
          <span>License</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {licenseInfo ? (
          <div className="space-y-3">
            <LicenseStatusBadge licenseInfo={licenseInfo} />
            <LicenseDetails licenseInfo={licenseInfo} />
            <LicenseFeatures licenseInfo={licenseInfo} />
          </div>
        ) : (
          <div className="text-sm text-[var(--color-text-muted)]">Loading license status...</div>
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
