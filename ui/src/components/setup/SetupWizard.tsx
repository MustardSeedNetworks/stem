import { Tooltip } from '../../shared/Tooltip';
/**
 * @fileoverview Initial Setup Wizard Component
 * @description Guides users through the first-time setup process for Stem application.
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import { Activity, Copy, Eye, EyeOff, Lock, Repeat, Target, Zap } from 'lucide-react';
import type { ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { type StemRole, useRole } from '../../contexts/RoleContext';
import { SetupWizardSchema } from '../../schemas/auth';
import { Button } from '../../shared/Button';
import { FormField } from '../../shared/FormField';
import { Input } from '../../shared/Input';
import { RadioGroup } from '../../shared/RadioGroup';
import { deadlineExpired, requestDeadline } from '../../utils/http';

/** Minimum password length (matches backend validation) */
const MIN_PASSWORD_LENGTH = 12;

interface SetupFormFields {
  setupToken: string;
  password: string;
  confirmPassword: string;
}

/** Props for SetupWizard component */
interface SetupWizardProps {
  /** Callback invoked when setup is complete and user is logged in */
  onComplete: () => void;
  /** Function to attempt login after password is set */
  onLogin: (username: string, password: string) => Promise<boolean>;
  /** Optional pre-generated password suggestion */
  suggestedPassword?: string;
  /** Username from setup status */
  username?: string;
  /** Host file the daemon wrote the one-time setup token to */
  setupTokenFile?: string;
}

/**
 * SetupWizard Component
 *
 * Modal-like component that requires user to set admin password before
 * accessing the main application.
 */
export function SetupWizard({
  onComplete,
  onLogin,
  suggestedPassword,
  username = 'admin',
  setupTokenFile,
}: SetupWizardProps): ReactElement {
  const { t } = useTranslation(['setup', 'help']);
  const { role: currentRole, setRole } = useRole();
  const [selectedRole, setSelectedRole] = useState<StemRole>(currentRole);
  const [passwordMode, setPasswordMode] = useState<'generated' | 'custom'>('custom');
  const [showPassword, setShowPassword] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<SetupFormFields>({
    resolver: valibotResolver(SetupWizardSchema),
    defaultValues: { setupToken: '', password: '', confirmPassword: '' },
    mode: 'onBlur',
  });

  // Update password fields when switching to generated mode
  useEffect(() => {
    if (passwordMode === 'generated' && suggestedPassword) {
      setValue('password', suggestedPassword, { shouldValidate: true, shouldDirty: true });
      setValue('confirmPassword', suggestedPassword, { shouldValidate: true, shouldDirty: true });
    }
  }, [passwordMode, suggestedPassword, setValue]);

  const handlePasswordModeChange = (mode: 'generated' | 'custom'): void => {
    setPasswordMode(mode);
    if (mode === 'generated' && suggestedPassword) {
      setValue('password', suggestedPassword, { shouldValidate: true, shouldDirty: true });
      setValue('confirmPassword', suggestedPassword, { shouldValidate: true, shouldDirty: true });
      setShowPassword(true);
    } else {
      setValue('password', '', { shouldValidate: false });
      setValue('confirmPassword', '', { shouldValidate: false });
      setShowPassword(false);
    }
    setSubmitError(null);
  };

  const handleCopyPassword = (): void => {
    if (suggestedPassword) {
      navigator.clipboard.writeText(suggestedPassword).catch(() => {
        // Clipboard API failed silently
      });
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const onSubmit: SubmitHandler<SetupFormFields> = async ({ setupToken, password }) => {
    setSubmitError(null);
    // Declared outside the try so the catch can read it.
    const deadline = requestDeadline();
    try {
      const response = await fetch('/api/v1/setup/complete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password, setupToken }),
        signal: deadline,
      });
      if (!response.ok) {
        const data = await (response.json() as Promise<{
          error?: string;
          message?: string;
        }>);
        setSubmitError(data.error ?? data.message ?? t('errors.setupFailed'));
        return;
      }
      // Persist the chosen role to RoleContext + localStorage.
      setRole(selectedRole);
      // Automatically log in with the new password.
      const loginSuccess = await onLogin(username, password);
      if (!loginSuccess) {
        setSubmitError(t('errors.loginFailed'));
        onComplete();
        return;
      }
      onComplete();
    } catch {
      setSubmitError(deadlineExpired(deadline) ? t('errors.timeout') : t('errors.networkError'));
    }
  };

  // Cross-field error (passwords don't match) from valibot v.check().
  const rootErrors = errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <div className="fixed inset-0 z-50 flex-center bg-scrim/60 backdrop-blur-sm pad">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-section">
          <div className="w-16 h-16 mx-auto mb-content flex-center rounded-2xl bg-brand-primary text-on-brand">
            <Activity className="w-8 h-8" />
          </div>
          <h1 className="heading-1 text-text-primary">{t('welcome.title')}</h1>
          <p className="text-sm text-text-muted mt-tight">{t('welcome.subtitle')}</p>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit(onSubmit)}
          className="rounded-3xl border border-surface-border bg-surface-raised pad-lg shadow-2xl"
        >
          {/* Username display */}
          <div className="mb-section pad-sm rounded-xl bg-surface-base border border-surface-border">
            <p className="text-sm text-text-muted">
              {t('username.label')} <strong className="text-text-primary">{username}</strong>
            </p>
            <p className="text-xs text-text-muted mt-tight">{t('username.description')}</p>
          </div>

          {/* Setup token: copied from the host, never sent by the server */}
          <FormField
            id="setup-token"
            label={t('setupToken.label')}
            error={errors.setupToken?.message}
            hint={setupTokenFile ? t('setupToken.help', { path: setupTokenFile }) : undefined}
            className="mb-section"
          >
            {(control) => (
              <Input
                {...control}
                data-testid="setup-token"
                type="password"
                autoComplete="off"
                placeholder={t('setupToken.placeholder')}
                {...register('setupToken')}
              />
            )}
          </FormField>

          {/* Role selection */}
          <div className="mb-section">
            <RadioGroup
              label={t('role.title')}
              hint={t('role.subtitle')}
              name="stemRole"
              value={selectedRole}
              onValueChange={(value) =>
                setSelectedRole(value === 'test_master' ? 'test_master' : 'reflector')
              }
              options={[
                {
                  value: 'reflector',
                  label: (
                    <span className="inline-flex items-center gap-2">
                      <Repeat className="h-4 w-4" aria-hidden="true" />
                      {t('role.reflector.title')}
                    </span>
                  ),
                  description: t('role.reflector.description'),
                },
                {
                  value: 'test_master',
                  label: (
                    <span className="inline-flex items-center gap-2">
                      <Target className="h-4 w-4" aria-hidden="true" />
                      {t('role.testMaster.title')}
                    </span>
                  ),
                  description: t('role.testMaster.description'),
                },
              ]}
            />
          </div>

          {/* Password mode selection */}
          <div className="mb-section stack">
            <RadioGroup
              label={t('password.chooseMethod')}
              name="passwordMode"
              value={passwordMode}
              onValueChange={(value) =>
                handlePasswordModeChange(value === 'generated' ? 'generated' : 'custom')
              }
              options={[
                {
                  value: 'custom',
                  label: (
                    <span className="inline-flex items-center gap-2">
                      <Lock className="h-4 w-4" aria-hidden="true" />
                      {t('password.custom.title')}
                    </span>
                  ),
                  description: t('password.custom.description'),
                },
                ...(suggestedPassword
                  ? [
                      {
                        value: 'generated',
                        label: (
                          <span className="inline-flex items-center gap-2">
                            <Zap className="h-4 w-4" aria-hidden="true" />
                            {t('password.generated.title')}
                          </span>
                        ),
                        description: t('password.generated.description'),
                      },
                    ]
                  : []),
              ]}
            />

            {suggestedPassword && passwordMode === 'generated' ? (
              <div className="pad-xs rounded-lg bg-surface-sunken border border-surface-border">
                <div className="flex items-center gap-compact">
                  <code className="flex-1 font-mono text-xs text-brand-primary select-all break-all">
                    {suggestedPassword}
                  </code>
                  <Tooltip text={t('buttons.copyTooltip')}>
                    <button
                      type="button"
                      onClick={handleCopyPassword}
                      className="shrink-0 p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-surface-base border border-surface-border"
                      aria-label={t('buttons.copyPassword')}
                    >
                      <Copy className="w-3.5 h-3.5" />
                    </button>
                  </Tooltip>
                </div>
                {copied ? (
                  <p className="text-xs text-status-success-strong mt-tight">
                    {t('buttons.copied')}
                  </p>
                ) : null}
                <p className="text-xs text-status-warning-strong mt-inline">
                  {t('password.generated.saveWarning')}
                </p>
              </div>
            ) : null}
          </div>

          {/* Custom password fields */}
          {passwordMode === 'custom' && (
            <>
              <FormField
                id="setup-password"
                label={t('password.label')}
                error={errors.password?.message}
                hint={
                  errors.password
                    ? undefined
                    : t('password.minLength', { count: MIN_PASSWORD_LENGTH })
                }
                className="mb-content"
              >
                {(control) => (
                  <div className="relative">
                    <Input
                      {...control}
                      type={showPassword ? 'text' : 'password'}
                      className="pr-10"
                      placeholder={t('password.placeholder')}
                      {...register('password')}
                    />
                    <Tooltip
                      text={
                        showPassword
                          ? t('help:tooltips.hidePassword')
                          : t('help:tooltips.showPassword')
                      }
                    >
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
                        aria-label={showPassword ? 'Hide password' : 'Show password'}
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </Tooltip>
                  </div>
                )}
              </FormField>

              <FormField
                id="setup-confirm-password"
                label={t('password.confirm.label')}
                error={errors.confirmPassword?.message}
                className="mb-section"
              >
                {(control) => (
                  <Input
                    {...control}
                    type={showPassword ? 'text' : 'password'}
                    placeholder={t('password.confirm.placeholder')}
                    {...register('confirmPassword')}
                  />
                )}
              </FormField>
            </>
          )}

          {/* Cross-field error (passwords don't match) */}
          {crossFieldError ? (
            <div
              role="alert"
              className="mb-content pad-sm rounded-xl bg-status-error/10 border border-status-error/20 text-sm text-status-error-strong"
            >
              {crossFieldError.message}
            </div>
          ) : null}

          {/* Submit error (network / server) */}
          {submitError !== null ? (
            <div
              role="alert"
              className="mb-content pad-sm rounded-xl bg-status-error/10 border border-status-error/20 text-sm text-status-error-strong"
            >
              {submitError}
            </div>
          ) : null}

          {/* Submit button */}
          <Button type="submit" disabled={isSubmitting} className="w-full">
            {isSubmitting ? t('buttons.settingUp') : t('buttons.completeSetup')}
          </Button>
        </form>

        {/* Footer */}
        <p className="text-xs text-text-muted text-center mt-content">
          {t('footer.copyright', { year: new Date().getFullYear() })}
        </p>
      </div>
    </div>
  );
}
