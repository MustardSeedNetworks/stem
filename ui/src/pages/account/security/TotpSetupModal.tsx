/**
 * TOTP enrolment modal.
 *
 * Displays the QR PNG returned by /api/v1/auth/totp/setup along with
 * the base32 secret (for users who can't scan) and a code-input that
 * POSTs to /api/v1/auth/totp/verify on submit.
 *
 * Migrated to react-hook-form + valibot per #325. The schema lives at
 * src/schemas/auth.ts (TotpSetupVerifySchema).
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import type { ReactElement } from 'react';
import { useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { TotpSetupVerifySchema } from '../../../schemas/auth';
import { fetchCsrfToken, mfaApi, type TotpSetupResponse } from './mfaApi';

interface Props {
  setup: TotpSetupResponse;
  onComplete: () => Promise<void> | void;
  onCancel: () => void;
}

interface TotpVerifyForm {
  code: string;
}

export function TotpSetupModal({ setup, onComplete, onCancel }: Props): ReactElement {
  const { t } = useTranslation('security');
  const [submitError, setSubmitError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<TotpVerifyForm>({
    resolver: valibotResolver(TotpSetupVerifySchema),
    defaultValues: { code: '' },
    mode: 'onBlur',
  });

  const onSubmit: SubmitHandler<TotpVerifyForm> = async ({ code }) => {
    setSubmitError(null);
    try {
      const csrf = await fetchCsrfToken();
      await mfaApi.totpVerify(code, csrf);
      await onComplete();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Verification failed');
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex-center bg-scrim/60 backdrop-blur-sm">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="totp-setup-title"
        className="w-full max-w-md rounded-3xl border border-surface-border bg-surface-raised pad-lg shadow-2xl"
      >
        <h2 id="totp-setup-title" className="heading-3 text-text-primary">
          {t('mfa.setup.title')}
        </h2>
        <p className="text-sm text-text-muted">{t('mfa.setup.instructions')}</p>

        <div className="mt-content flex justify-center">
          <img
            src={`data:image/png;base64,${setup.qrCodePngBase64}`}
            alt="TOTP QR code"
            width={256}
            height={256}
            className="rounded-lg border border-surface-border bg-knob pad-xs"
          />
        </div>

        <div className="mt-content">
          <p className="text-xs text-text-muted">{t('mfa.setup.secretLabel')}</p>
          <code className="block mt-tight break-all rounded-lg bg-surface-base pad-xs text-xs font-mono text-text-primary">
            {setup.secret}
          </code>
        </div>

        <form className="mt-content stack-lg" onSubmit={handleSubmit(onSubmit)}>
          <div>
            <label htmlFor="totp-setup-code" className="text-xs font-semibold text-text-muted">
              {t('mfa.setup.codeLabel')}
            </label>
            <input
              id="totp-setup-code"
              type="text"
              inputMode="numeric"
              pattern="[0-9]{6}"
              placeholder={t('mfa.setup.codePlaceholder')}
              {...register('code')}
              className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm font-mono tracking-widest"
            />
            {errors.code ? (
              <p className="mt-tight text-xs text-status-error">{errors.code.message}</p>
            ) : null}
          </div>
          {submitError ? <p className="text-xs text-status-error">{submitError}</p> : null}
          <div className="flex gap-compact justify-end">
            <button type="button" className="btn btn-secondary" onClick={onCancel}>
              {t('mfa.setup.cancelButton')}
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
              {isSubmitting ? '...' : t('mfa.setup.confirmButton')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
