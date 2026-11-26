/**
 * Account Security page.
 *
 * Wave 3 (#85) introduces MFA + WebAuthn. This page is the user-
 * facing surface for enrolling / disabling the second factor and for
 * adding passkeys. Kept intentionally compact — the dominant UX is the
 * login flow, not this configuration page.
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import { type ReactElement, useCallback, useEffect, useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { registerPasskey } from '../../../lib/webauthn-register';
import { TotpDisableSchema } from '../../../schemas/auth';
import { Alert } from '../../../shared/Alert';
import { Button } from '../../../shared/Button';
import { Card } from '../../../shared/Card';
import { FormField } from '../../../shared/FormField';
import { Input } from '../../../shared/Input';
import { Modal } from '../../../shared/Modal';
import { StatusPill } from '../../../shared/StatusPill';
import { RefreshAction } from '../../../ui/PageActions';
import {
  isMFARequired as _isMFARequired,
  type MFAStatusResponse,
  mfaApi,
  type TotpSetupResponse,
} from './mfaApi';
import { TotpSetupModal } from './TotpSetupModal';

// Keep the unused export sentry happy so tree-shakers don't drop the
// helper. The login form (App.tsx) is the actual consumer; importing
// it from a typed location keeps the surface discoverable.
export { _isMFARequired as isMFARequired };

export function SecurityPage(): ReactElement {
  const { t } = useTranslation(['security', 'common']);
  const [status, setStatus] = useState<MFAStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [setupOpen, setSetupOpen] = useState(false);
  const [setupData, setSetupData] = useState<TotpSetupResponse | null>(null);
  const [passkeyBusy, setPasskeyBusy] = useState(false);
  const [passkeyMsg, setPasskeyMsg] = useState<string | null>(null);

  const refresh = useCallback(async (): Promise<void> => {
    try {
      const s = await mfaApi.status();
      setStatus(s);
      setError(null);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to load status';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh().catch(() => {
      // refresh sets its own error state.
    });
  }, [refresh]);

  const handleEnableTOTP = useCallback(async (): Promise<void> => {
    setError(null);
    try {
      const data = await mfaApi.totpSetup();
      setSetupData(data);
      setSetupOpen(true);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to begin TOTP setup';
      setError(msg);
    }
  }, []);

  const handleSetupComplete = useCallback(async (): Promise<void> => {
    setSetupOpen(false);
    setSetupData(null);
    await refresh();
  }, [refresh]);

  const handleAddPasskey = useCallback(async (): Promise<void> => {
    setPasskeyBusy(true);
    setPasskeyMsg(null);
    setError(null);
    try {
      await registerPasskey();
      setPasskeyMsg(t('passkeys.successMessage'));
      await refresh();
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Passkey registration failed';
      setError(msg);
    } finally {
      setPasskeyBusy(false);
    }
  }, [refresh, t]);

  if (loading) {
    return <div className="text-text-muted">{t('common:status.loading')}</div>;
  }

  return (
    <section className="stack-xl">
      <RefreshAction onRefresh={() => void refresh()} />
      {error ? <Alert tone="error">{error}</Alert> : null}

      <Card
        title={t('mfa.heading')}
        headerAction={
          status?.totpEnabled ? (
            <DisableTotpButton onDisabled={refresh} />
          ) : (
            <Button variant="primary" onClick={handleEnableTOTP}>
              {t('mfa.enableButton')}
            </Button>
          )
        }
      >
        <div className="mb-2">
          <StatusPill
            tone={status?.totpEnabled ? 'success' : 'unknown'}
            label={status?.totpEnabled ? t('mfa.statusEnabled') : t('mfa.statusDisabled')}
          />
        </div>
        <p className="text-sm text-text-muted">{t('mfa.description')}</p>
      </Card>

      <Card
        title={t('passkeys.heading')}
        headerAction={
          <Button variant="secondary" onClick={handleAddPasskey} loading={passkeyBusy}>
            {passkeyBusy ? t('passkeys.addingButton') : t('passkeys.addButton')}
          </Button>
        }
      >
        <p className="text-sm text-text-muted">{t('passkeys.description')}</p>
        <p data-testid="passkey-count" className="mt-1 text-xs text-text-muted">
          {t('passkeys.countLabel', { count: status?.webauthnCredentialCount ?? 0 })}
        </p>
        {passkeyMsg ? (
          <p className="mt-1 text-xs text-status-success-strong">{passkeyMsg}</p>
        ) : null}
      </Card>

      {setupOpen && setupData ? (
        <TotpSetupModal
          setup={setupData}
          onComplete={handleSetupComplete}
          onCancel={() => setSetupOpen(false)}
        />
      ) : null}
    </section>
  );
}

interface DisableTotpButtonProps {
  onDisabled: () => Promise<void>;
}

interface TotpDisableForm {
  password: string;
  code: string;
}

function DisableTotpButton({ onDisabled }: DisableTotpButtonProps): ReactElement {
  const { t } = useTranslation('security');
  const { t: tCommon } = useTranslation('common');
  const [open, setOpen] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<TotpDisableForm>({
    resolver: valibotResolver(TotpDisableSchema),
    defaultValues: { password: '', code: '' },
    mode: 'onBlur',
  });

  const onSubmit: SubmitHandler<TotpDisableForm> = async ({ password, code }) => {
    setSubmitError(null);
    try {
      await mfaApi.totpDisable(password, code);
      setOpen(false);
      reset();
      await onDisabled();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to disable TOTP');
    }
  };

  return (
    <>
      <Button variant="destructive" onClick={() => setOpen(true)}>
        {t('mfa.disableButton')}
      </Button>
      <Modal
        open={open}
        onClose={() => setOpen(false)}
        title={t('mfa.disable.title')}
        closeLabel={tCommon('buttons.close')}
        busy={isSubmitting}
        footer={
          <>
            <Button variant="secondary" onClick={() => setOpen(false)} disabled={isSubmitting}>
              {t('mfa.disable.cancelButton')}
            </Button>
            <Button
              type="submit"
              form="totp-disable-form"
              variant="destructive"
              loading={isSubmitting}
            >
              {t('mfa.disable.confirmButton')}
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-muted">{t('mfa.disable.instructions')}</p>
        <form id="totp-disable-form" className="mt-4 stack-lg" onSubmit={handleSubmit(onSubmit)}>
          <FormField label={t('mfa.disable.passwordLabel')} error={errors.password?.message}>
            {(control) => (
              <Input
                {...control}
                type="password"
                autoComplete="current-password"
                {...register('password')}
              />
            )}
          </FormField>
          <FormField label={t('mfa.disable.codeLabel')} error={errors.code?.message}>
            {(control) => (
              <Input
                {...control}
                type="text"
                inputMode="numeric"
                pattern="[0-9]{6}"
                className="font-mono tracking-widest"
                {...register('code')}
              />
            )}
          </FormField>
          {submitError ? <Alert tone="error">{submitError}</Alert> : null}
        </form>
      </Modal>
    </>
  );
}
