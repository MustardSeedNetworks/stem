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
import { Alert } from '../../../shared/Alert';
import { Button } from '../../../shared/Button';
import { FormField } from '../../../shared/FormField';
import { Input } from '../../../shared/Input';
import { Modal } from '../../../shared/Modal';
import { mfaApi, type TotpSetupResponse } from './mfaApi';

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
  const { t: tCommon } = useTranslation('common');
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
      await mfaApi.totpVerify(code);
      await onComplete();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Verification failed');
    }
  };

  return (
    <Modal
      open={true}
      onClose={onCancel}
      title={t('mfa.setup.title')}
      closeLabel={tCommon('buttons.close')}
      busy={isSubmitting}
      footer={
        <>
          <Button variant="secondary" onClick={onCancel} disabled={isSubmitting}>
            {t('mfa.setup.cancelButton')}
          </Button>
          <Button type="submit" form="totp-setup-form" loading={isSubmitting}>
            {t('mfa.setup.confirmButton')}
          </Button>
        </>
      }
    >
      <p className="text-sm text-text-muted">{t('mfa.setup.instructions')}</p>

      <div className="mt-4 flex justify-center">
        <img
          src={`data:image/png;base64,${setup.qrCodePngBase64}`}
          alt="TOTP QR code"
          width={256}
          height={256}
          className="rounded-lg border border-hairline bg-knob p-2"
        />
      </div>

      <div className="mt-4">
        <p className="text-xs text-text-muted">{t('mfa.setup.secretLabel')}</p>
        <code className="mt-1 block break-all rounded-lg bg-surface-sunken p-2 font-mono text-xs text-text-primary">
          {setup.secret}
        </code>
      </div>

      <form id="totp-setup-form" className="mt-4 stack-lg" onSubmit={handleSubmit(onSubmit)}>
        <FormField label={t('mfa.setup.codeLabel')} error={errors.code?.message}>
          {(control) => (
            <Input
              {...control}
              type="text"
              inputMode="numeric"
              pattern="[0-9]{6}"
              placeholder={t('mfa.setup.codePlaceholder')}
              className="font-mono tracking-widest"
              {...register('code')}
            />
          )}
        </FormField>
        {submitError ? <Alert tone="error">{submitError}</Alert> : null}
      </form>
    </Modal>
  );
}
