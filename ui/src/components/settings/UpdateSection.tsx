/**
 * UpdateSection Component
 *
 * The Settings toggle for the daily automatic update (ADR-0012). It is off
 * until an operator turns it on; the daemon only saves the choice, and the
 * root-owned update timer reads it.
 */

import type { ReactElement } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card } from '../../shared/Card';
import { Toggle } from '../../shared/Toggle';
import { authFetch } from '../../stores/auth-store';
import type { UpdateSettings } from '../../types/generated/update-settings';
import { useToast } from '../ui/useToast';

export const UPDATE_SETTINGS_ENDPOINT = '/api/v1/update/settings';

export function UpdateSection(): ReactElement {
  const { t } = useTranslation(['common', 'settings', 'errors']);
  const [settings, setSettings] = useState<UpdateSettings | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { addToast } = useToast();

  const load = useCallback(async (): Promise<void> => {
    try {
      const response = await authFetch(UPDATE_SETTINGS_ENDPOINT);
      if (!response.ok) {
        setError(t('settings:update.loadFailed'));
        return;
      }
      setSettings((await response.json()) as UpdateSettings);
    } catch {
      setError(t('errors:network.connectionFailed'));
    }
  }, [t]);

  useEffect(() => {
    load().catch(() => {
      // load reports its own failures.
    });
  }, [load]);

  const save = async (automatic: boolean): Promise<void> => {
    setSaving(true);
    setError(null);
    try {
      const response = await authFetch(UPDATE_SETTINGS_ENDPOINT, {
        method: 'POST',
        body: JSON.stringify({ automatic }),
      });
      if (response.ok) {
        setSettings((await response.json()) as UpdateSettings);
      } else {
        setError(t('settings:update.saveFailed'));
        addToast(t('settings:update.saveFailed'), 'error');
      }
    } catch {
      setError(t('errors:network.connectionFailed'));
      addToast(t('errors:network.connectionFailed'), 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card title={t('settings:update.sectionTitle')}>
      <div className="stack" data-testid="settings-update-section">
        {settings ? (
          <>
            <Toggle
              label={t('settings:update.automatic')}
              description={t('settings:update.automaticDesc')}
              checked={settings.automatic}
              disabled={saving}
              onChange={(checked): void => {
                save(checked).catch(() => {
                  // save reports its own failures.
                });
              }}
            />
            <div className="flex justify-between text-sm">
              <span className="text-text-muted">{t('settings:update.source')}</span>
              <span className="font-mono text-xs" data-testid="settings-update-source">
                {settings.source}
              </span>
            </div>
          </>
        ) : (
          !error && <div className="text-sm text-text-muted">{t('settings:update.loading')}</div>
        )}
        {error ? (
          <div
            role="alert"
            className="text-sm text-status-error-strong bg-status-error/10 pad-xs rounded"
          >
            {error}
          </div>
        ) : null}
      </div>
    </Card>
  );
}
