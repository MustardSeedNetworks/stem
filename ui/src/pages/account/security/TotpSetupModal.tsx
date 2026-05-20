/**
 * TOTP enrolment modal.
 *
 * Displays the QR PNG returned by /api/v1/auth/totp/setup along with
 * the base32 secret (for users who can't scan) and a code-input that
 * POSTs to /api/v1/auth/totp/verify on submit.
 */

import { type FormEvent, type ReactElement, useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { fetchCsrfToken, mfaApi, type TotpSetupResponse } from './mfaApi';

interface Props {
  setup: TotpSetupResponse;
  onComplete: () => Promise<void> | void;
  onCancel: () => void;
}

export function TotpSetupModal({ setup, onComplete, onCancel }: Props): ReactElement {
  const { t } = useTranslation('security');
  const [code, setCode] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = useCallback(
    async (event: FormEvent<HTMLFormElement>): Promise<void> => {
      event.preventDefault();
      setBusy(true);
      setError(null);
      try {
        const csrf = await fetchCsrfToken();
        await mfaApi.totpVerify(code, csrf);
        await onComplete();
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Verification failed';
        setError(msg);
      } finally {
        setBusy(false);
      }
    },
    [code, onComplete],
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="totp-setup-title"
        className="w-full max-w-md rounded-3xl border border-surface-border bg-surface-raised p-6 shadow-2xl"
      >
        <h2 id="totp-setup-title" className="text-lg font-semibold text-text-primary">
          {t('mfa.setup.title')}
        </h2>
        <p className="text-sm text-text-muted">{t('mfa.setup.instructions')}</p>

        <div className="mt-4 flex justify-center">
          <img
            src={`data:image/png;base64,${setup.qrCodePngBase64}`}
            alt="TOTP QR code"
            width={256}
            height={256}
            className="rounded-lg border border-surface-border bg-white p-2"
          />
        </div>

        <div className="mt-4">
          <p className="text-xs text-text-muted">{t('mfa.setup.secretLabel')}</p>
          <code className="block mt-1 break-all rounded-lg bg-surface-base p-2 text-xs font-mono text-text-primary">
            {setup.secret}
          </code>
        </div>

        <form className="mt-4 space-y-4" onSubmit={handleSubmit}>
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
              value={code}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCode(e.target.value)}
              className="mt-1 w-full rounded-xl border border-surface-border bg-surface-base px-3 py-2 text-sm font-mono tracking-widest"
            />
          </div>
          {error ? <p className="text-xs text-[var(--color-status-error)]">{error}</p> : null}
          <div className="flex gap-2 justify-end">
            <button type="button" className="btn btn-secondary" onClick={onCancel}>
              {t('mfa.setup.cancelButton')}
            </button>
            <button type="submit" className="btn btn-primary" disabled={busy}>
              {busy ? '...' : t('mfa.setup.confirmButton')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
