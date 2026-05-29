/**
 * Account Security page.
 *
 * Wave 3 (#85) introduces MFA + WebAuthn. This page is the user-
 * facing surface for enrolling / disabling the second factor and for
 * adding passkeys. Kept intentionally compact — the dominant UX is the
 * login flow, not this configuration page.
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import { Lock, ShieldCheck, ShieldOff } from 'lucide-react';
import { type ReactElement, useCallback, useEffect, useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { TotpDisableSchema } from '../../../schemas/auth';
import {
  isMFARequired as _isMFARequired,
  fetchCsrfToken,
  MFAError,
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
  const { t } = useTranslation('security');
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
      const csrf = await fetchCsrfToken();
      const data = await mfaApi.totpSetup(csrf);
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
      const csrf = await fetchCsrfToken();
      const options = await mfaApi.webauthnRegisterBegin(csrf);
      // The server returns the WebAuthn options in JSON; the browser
      // needs ArrayBuffer values for challenge/user.id. Cast and let
      // the browser API handle the rest — we rely on the operator's
      // browser supporting the @simplewebauthn-style JSON shape.
      const credential = (await navigator.credentials.create({
        publicKey: options as PublicKeyCredentialCreationOptions,
      })) as PublicKeyCredential | null;
      if (!credential) {
        throw new MFAError(0, 'No credential returned by browser');
      }
      await mfaApi.webauthnRegisterFinish(credential, csrf);
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
    return <div className="text-text-muted">Loading...</div>;
  }

  return (
    <section className="stack-xl">
      <header>
        <h1 className="heading-1">{t('title')}</h1>
        <p className="text-sm text-text-muted">{t('subtitle')}</p>
      </header>

      {error ? (
        <div
          role="alert"
          className="rounded-lg border border-status-error/40 bg-status-error/10 pad-sm text-sm text-status-error"
        >
          {error}
        </div>
      ) : null}

      {/* TOTP card */}
      <section className="rounded-2xl border border-surface-border bg-surface-raised pad-lg stack-lg">
        <div className="flex items-start justify-between gap-comfortable">
          <div className="flex items-start gap-default">
            {status?.totpEnabled ? (
              <ShieldCheck className="w-5 h-5 text-status-success mt-0.5" />
            ) : (
              <ShieldOff className="w-5 h-5 text-text-muted mt-0.5" />
            )}
            <div>
              <h2 className="heading-3 text-text-primary">{t('mfa.heading')}</h2>
              <p className="text-sm text-text-muted">
                {status?.totpEnabled ? t('mfa.statusEnabled') : t('mfa.statusDisabled')}
              </p>
              <p className="text-sm text-text-muted mt-inline">{t('mfa.description')}</p>
            </div>
          </div>
          <div>
            {status?.totpEnabled ? (
              <DisableTotpButton onDisabled={refresh} />
            ) : (
              <button type="button" className="btn btn-primary" onClick={handleEnableTOTP}>
                {t('mfa.enableButton')}
              </button>
            )}
          </div>
        </div>
      </section>

      {/* Passkeys card */}
      <section className="rounded-2xl border border-surface-border bg-surface-raised pad-lg stack-lg">
        <div className="flex items-start justify-between gap-comfortable">
          <div className="flex items-start gap-default">
            <Lock className="w-5 h-5 text-text-muted mt-0.5" />
            <div>
              <h2 className="heading-3 text-text-primary">{t('passkeys.heading')}</h2>
              <p className="text-sm text-text-muted">{t('passkeys.description')}</p>
              <p className="text-xs text-text-muted mt-inline">
                {t('passkeys.countLabel', { count: status?.webauthnCredentialCount ?? 0 })}
              </p>
              {passkeyMsg ? (
                <p className="text-xs text-status-success mt-inline">{passkeyMsg}</p>
              ) : null}
            </div>
          </div>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleAddPasskey}
            disabled={passkeyBusy}
          >
            {passkeyBusy ? t('passkeys.addingButton') : t('passkeys.addButton')}
          </button>
        </div>
      </section>

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
      const csrf = await fetchCsrfToken();
      await mfaApi.totpDisable(password, code, csrf);
      setOpen(false);
      reset();
      await onDisabled();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to disable TOTP');
    }
  };

  if (!open) {
    return (
      <button type="button" className="btn btn-secondary" onClick={() => setOpen(true)}>
        {t('mfa.disableButton')}
      </button>
    );
  }

  return (
    <div className="fixed inset-0 z-50 flex-center bg-scrim/60 backdrop-blur-sm">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="totp-disable-title"
        className="w-full max-w-md rounded-3xl border border-surface-border bg-surface-raised pad-lg shadow-2xl"
      >
        <h2 id="totp-disable-title" className="heading-3 text-text-primary">
          {t('mfa.disable.title')}
        </h2>
        <p className="text-sm text-text-muted">{t('mfa.disable.instructions')}</p>
        <form className="mt-content stack-lg" onSubmit={handleSubmit(onSubmit)}>
          <div>
            <label
              htmlFor="totp-disable-password"
              className="text-xs font-semibold text-text-muted"
            >
              {t('mfa.disable.passwordLabel')}
            </label>
            <input
              id="totp-disable-password"
              type="password"
              autoComplete="current-password"
              {...register('password')}
              className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm"
            />
            {errors.password ? (
              <p className="mt-tight text-xs text-status-error">{errors.password.message}</p>
            ) : null}
          </div>
          <div>
            <label htmlFor="totp-disable-code" className="text-xs font-semibold text-text-muted">
              {t('mfa.disable.codeLabel')}
            </label>
            <input
              id="totp-disable-code"
              type="text"
              inputMode="numeric"
              pattern="[0-9]{6}"
              {...register('code')}
              className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm font-mono tracking-widest"
            />
            {errors.code ? (
              <p className="mt-tight text-xs text-status-error">{errors.code.message}</p>
            ) : null}
          </div>
          {submitError ? <p className="text-xs text-status-error">{submitError}</p> : null}
          <div className="flex gap-compact justify-end">
            <button type="button" className="btn btn-secondary" onClick={() => setOpen(false)}>
              {t('mfa.disable.cancelButton')}
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
              {isSubmitting ? '...' : t('mfa.disable.confirmButton')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
