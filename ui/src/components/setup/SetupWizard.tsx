/**
 * @fileoverview Initial Setup Wizard Component
 * @description Guides users through the first-time setup process for The Stem application.
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import { Activity, Copy, Eye, EyeOff, Lock, Repeat, Target, Zap } from 'lucide-react';
import type { ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { type StemRole, useRole } from '../../contexts/RoleContext';
import { SetupWizardSchema } from '../../schemas/auth';

/** Minimum password length (matches backend validation) */
const MIN_PASSWORD_LENGTH = 12;

interface SetupFormFields {
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
  /** One-time setup token required for security */
  setupToken?: string;
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
  setupToken,
}: SetupWizardProps): ReactElement {
  const { t } = useTranslation('setup');
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
    defaultValues: { password: '', confirmPassword: '' },
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

  const onSubmit: SubmitHandler<SetupFormFields> = async ({ password }) => {
    setSubmitError(null);
    try {
      const response = await fetch('/api/v1/setup/complete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password, setupToken }),
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
      setSubmitError(t('errors.networkError'));
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
    <div className="fixed inset-0 z-50 flex-center bg-black/60 backdrop-blur-sm pad">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-section">
          <div className="w-16 h-16 mx-auto mb-content flex-center rounded-2xl bg-brand-primary text-white">
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

          {/* Role selection */}
          <div className="mb-section stack">
            <div>
              <p className="text-xs font-semibold text-text-muted">
                {t('role.title', "Choose this stem's role")}
              </p>
              <p className="text-xs text-text-muted mt-tight">
                {t(
                  'role.subtitle',
                  'Each stem instance runs as either a passive Reflector or an active Test Master. You can switch roles later from the header.',
                )}
              </p>
            </div>

            <label className="flex items-start gap-default pad-sm rounded-xl border border-surface-border cursor-pointer hover:bg-surface-base transition-colors">
              <input
                type="radio"
                name="stemRole"
                value="reflector"
                checked={selectedRole === 'reflector'}
                onChange={() => setSelectedRole('reflector')}
                className="mt-tight w-4 h-4 text-brand-primary"
              />
              <div>
                <span className="label flex items-center gap-compact">
                  <Repeat className="w-4 h-4" />
                  {t('role.reflector.title', 'Reflector')}
                </span>
                <p className="text-xs text-text-muted mt-tight">
                  {t(
                    'role.reflector.description',
                    'Passive loopback — bounces frames back to a Test Master for end-to-end measurement.',
                  )}
                </p>
              </div>
            </label>

            <label className="flex items-start gap-default pad-sm rounded-xl border border-surface-border cursor-pointer hover:bg-surface-base transition-colors">
              <input
                type="radio"
                name="stemRole"
                value="test_master"
                checked={selectedRole === 'test_master'}
                onChange={() => setSelectedRole('test_master')}
                className="mt-tight w-4 h-4 text-brand-primary"
              />
              <div>
                <span className="label flex items-center gap-compact">
                  <Target className="w-4 h-4" />
                  {t('role.testMaster.title', 'Test Master')}
                </span>
                <p className="text-xs text-text-muted mt-tight">
                  {t(
                    'role.testMaster.description',
                    'Active testing — runs RFC 2544, Y.1564, Y.1731, MEF, TSN, and traffic-generation modules.',
                  )}
                </p>
              </div>
            </label>
          </div>

          {/* Password mode selection */}
          <div className="mb-section stack">
            <p className="text-xs font-semibold text-text-muted">{t('password.chooseMethod')}</p>

            {/* Custom password option */}
            <label className="flex items-start gap-default pad-sm rounded-xl border border-surface-border cursor-pointer hover:bg-surface-base transition-colors">
              <input
                type="radio"
                name="passwordMode"
                value="custom"
                checked={passwordMode === 'custom'}
                onChange={() => handlePasswordModeChange('custom')}
                className="mt-tight w-4 h-4 text-brand-primary"
              />
              <div>
                <span className="label flex items-center gap-compact">
                  <Lock className="w-4 h-4" />
                  {t('password.custom.title')}
                </span>
                <p className="text-xs text-text-muted mt-tight">
                  {t('password.custom.description')}
                </p>
              </div>
            </label>

            {/* Generated password option */}
            {suggestedPassword ? (
              <label className="flex items-start gap-default pad-sm rounded-xl border border-surface-border cursor-pointer hover:bg-surface-base transition-colors">
                <input
                  type="radio"
                  name="passwordMode"
                  value="generated"
                  checked={passwordMode === 'generated'}
                  onChange={() => handlePasswordModeChange('generated')}
                  className="mt-tight w-4 h-4 text-brand-primary"
                />
                <div className="flex-1">
                  <span className="label flex items-center gap-compact">
                    <Zap className="w-4 h-4" />
                    {t('password.generated.title')}
                  </span>
                  <p className="text-xs text-text-muted mt-tight">
                    {t('password.generated.description')}
                  </p>
                  {passwordMode === 'generated' ? (
                    <div className="mt-heading pad-xs rounded-lg bg-surface-sunken border border-surface-border">
                      <div className="flex items-center gap-compact">
                        <code className="flex-1 font-mono text-xs text-brand-primary select-all break-all">
                          {suggestedPassword}
                        </code>
                        <button
                          type="button"
                          onClick={handleCopyPassword}
                          className="shrink-0 p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-surface-base border border-surface-border"
                          title={t(
                            'buttons.copyTooltip',
                            'Copy the generated password to the clipboard so you can save it in a password manager',
                          )}
                          aria-label={t('buttons.copy', 'Copy password to clipboard')}
                        >
                          <Copy className="w-3.5 h-3.5" />
                        </button>
                      </div>
                      {copied ? (
                        <p className="text-xs text-status-success mt-tight">
                          {t('buttons.copied')}
                        </p>
                      ) : null}
                      <p className="text-xs text-status-warning mt-inline">
                        {t('password.generated.saveWarning')}
                      </p>
                    </div>
                  ) : null}
                </div>
              </label>
            ) : null}
          </div>

          {/* Custom password fields */}
          {passwordMode === 'custom' && (
            <>
              <div className="mb-content">
                <label htmlFor="setup-password" className="text-xs font-semibold text-text-muted">
                  {t('password.label')}
                </label>
                <div className="relative mt-tight">
                  <input
                    id="setup-password"
                    type={showPassword ? 'text' : 'password'}
                    {...register('password')}
                    className="w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row pr-icon text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                    placeholder={t('password.placeholder')}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
                    title={
                      showPassword
                        ? 'Hide the password (mask characters with dots)'
                        : 'Show the password in plain text to verify what you typed'
                    }
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                {errors.password ? (
                  <p className="text-xs text-status-error mt-tight">{errors.password.message}</p>
                ) : (
                  <p className="text-xs text-text-muted mt-tight">
                    {t('password.minLength', { count: MIN_PASSWORD_LENGTH })}
                  </p>
                )}
              </div>

              <div className="mb-section">
                <label
                  htmlFor="setup-confirm-password"
                  className="text-xs font-semibold text-text-muted"
                >
                  {t('password.confirm.label')}
                </label>
                <input
                  id="setup-confirm-password"
                  type={showPassword ? 'text' : 'password'}
                  {...register('confirmPassword')}
                  className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                  placeholder={t('password.confirm.placeholder')}
                />
                {errors.confirmPassword ? (
                  <p className="text-xs text-status-error mt-tight">
                    {errors.confirmPassword.message}
                  </p>
                ) : null}
              </div>
            </>
          )}

          {/* Cross-field error (passwords don't match) */}
          {crossFieldError ? (
            <div
              role="alert"
              className="mb-content pad-sm rounded-xl bg-status-error/10 border border-status-error/20 text-sm text-status-error"
            >
              {crossFieldError.message}
            </div>
          ) : null}

          {/* Submit error (network / server) */}
          {submitError !== null ? (
            <div
              role="alert"
              className="mb-content pad-sm rounded-xl bg-status-error/10 border border-status-error/20 text-sm text-status-error"
            >
              {submitError}
            </div>
          ) : null}

          {/* Submit button */}
          <button
            type="submit"
            disabled={isSubmitting}
            className="btn btn-primary w-full justify-center"
          >
            {isSubmitting ? t('buttons.settingUp') : t('buttons.completeSetup')}
          </button>
        </form>

        {/* Footer */}
        <p className="text-xs text-text-muted text-center mt-content">
          {t('footer.copyright', { year: new Date().getFullYear() })}
        </p>
      </div>
    </div>
  );
}
