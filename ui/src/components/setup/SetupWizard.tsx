/**
 * @fileoverview Initial Setup Wizard Component
 * @description Guides users through the first-time setup process for The Stem application.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Activity, Copy, Eye, EyeOff, Lock, Zap } from 'lucide-react';
import type { FormEvent, ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

/** Minimum password length (matches backend validation) */
const MIN_PASSWORD_LENGTH = 12;

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
  const [passwordMode, setPasswordMode] = useState<'generated' | 'custom'>('custom');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [copied, setCopied] = useState(false);

  // Update password fields when switching to generated mode
  useEffect(() => {
    if (passwordMode === 'generated' && suggestedPassword) {
      setPassword(suggestedPassword);
      setConfirmPassword(suggestedPassword);
    }
  }, [passwordMode, suggestedPassword]);

  const handlePasswordModeChange = (mode: 'generated' | 'custom'): void => {
    setPasswordMode(mode);
    if (mode === 'generated' && suggestedPassword) {
      setPassword(suggestedPassword);
      setConfirmPassword(suggestedPassword);
      setShowPassword(true);
    } else {
      setPassword('');
      setConfirmPassword('');
      setShowPassword(false);
    }
    setError(null);
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

  const handleSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);

    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(t('errors.passwordTooShort', { count: MIN_PASSWORD_LENGTH }));
      return;
    }

    if (password !== confirmPassword) {
      setError(t('errors.passwordMismatch'));
      return;
    }

    setIsSubmitting(true);

    try {
      // Step 1: Complete setup (set password on server)
      const response = await fetch('/api/v1/setup/complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ password, setupToken }),
      });

      if (!response.ok) {
        const data = await (response.json() as Promise<{
          error?: string;
          message?: string;
        }>);
        setError(data.error ?? data.message ?? t('errors.setupFailed'));
        return;
      }

      // Step 2: Automatically log in with the new password
      const loginSuccess = await onLogin(username, password);

      if (!loginSuccess) {
        setError(t('errors.loginFailed'));
        onComplete();
        return;
      }

      // Step 3: Setup complete and user is logged in
      onComplete();
    } catch {
      setError(t('errors.networkError'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="w-16 h-16 mx-auto mb-4 flex items-center justify-center rounded-2xl bg-[var(--color-brand-primary)] text-white">
            <Activity className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
            {t('welcome.title')}
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">{t('welcome.subtitle')}</p>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="rounded-3xl border border-[var(--color-surface-border)] bg-[var(--color-surface-raised)] p-6 shadow-2xl"
        >
          {/* Username display */}
          <div className="mb-6 p-3 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
            <p className="text-sm text-[var(--color-text-muted)]">
              {t('username.label')}{' '}
              <strong className="text-[var(--color-text-primary)]">{username}</strong>
            </p>
            <p className="text-xs text-[var(--color-text-muted)] mt-1">
              {t('username.description')}
            </p>
          </div>

          {/* Password mode selection */}
          <div className="mb-6 space-y-3">
            <p className="text-xs font-semibold text-[var(--color-text-muted)]">
              {t('password.chooseMethod')}
            </p>

            {/* Custom password option */}
            <label className="flex items-start gap-3 p-3 rounded-xl border border-[var(--color-surface-border)] cursor-pointer hover:bg-[var(--color-surface-base)] transition-colors">
              <input
                type="radio"
                name="passwordMode"
                value="custom"
                checked={passwordMode === 'custom'}
                onChange={() => handlePasswordModeChange('custom')}
                className="mt-1 w-4 h-4 text-[var(--color-brand-primary)]"
              />
              <div>
                <span className="text-sm font-medium text-[var(--color-text-primary)] flex items-center gap-2">
                  <Lock className="w-4 h-4" />
                  {t('password.custom.title')}
                </span>
                <p className="text-xs text-[var(--color-text-muted)] mt-1">
                  {t('password.custom.description')}
                </p>
              </div>
            </label>

            {/* Generated password option */}
            {suggestedPassword ? (
              <label className="flex items-start gap-3 p-3 rounded-xl border border-[var(--color-surface-border)] cursor-pointer hover:bg-[var(--color-surface-base)] transition-colors">
                <input
                  type="radio"
                  name="passwordMode"
                  value="generated"
                  checked={passwordMode === 'generated'}
                  onChange={() => handlePasswordModeChange('generated')}
                  className="mt-1 w-4 h-4 text-[var(--color-brand-primary)]"
                />
                <div className="flex-1">
                  <span className="text-sm font-medium text-[var(--color-text-primary)] flex items-center gap-2">
                    <Zap className="w-4 h-4" />
                    {t('password.generated.title')}
                  </span>
                  <p className="text-xs text-[var(--color-text-muted)] mt-1">
                    {t('password.generated.description')}
                  </p>
                  {passwordMode === 'generated' ? (
                    <div className="mt-3 p-2 rounded-lg bg-[var(--color-surface-sunken)] border border-[var(--color-surface-border)]">
                      <div className="flex items-center gap-2">
                        <code className="flex-1 font-mono text-xs text-[var(--color-brand-primary)] select-all break-all">
                          {suggestedPassword}
                        </code>
                        <button
                          type="button"
                          onClick={handleCopyPassword}
                          className="shrink-0 p-1.5 rounded-md text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]"
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
                        <p className="text-xs text-[var(--color-status-success)] mt-1">
                          {t('buttons.copied')}
                        </p>
                      ) : null}
                      <p className="text-xs text-[var(--color-status-warning)] mt-2">
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
              <div className="mb-4">
                <label
                  htmlFor="setup-password"
                  className="text-xs font-semibold text-[var(--color-text-muted)]"
                >
                  {t('password.label')}
                </label>
                <div className="relative mt-1">
                  <input
                    id="setup-password"
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                      setPassword(e.target.value)
                    }
                    className="w-full rounded-xl border border-[var(--color-surface-border)] bg-[var(--color-surface-base)] px-3 py-2 pr-10 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30"
                    placeholder={t('password.placeholder')}
                    required={true}
                    minLength={MIN_PASSWORD_LENGTH}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
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
                <p className="text-xs text-[var(--color-text-muted)] mt-1">
                  {t('password.minLength', { count: MIN_PASSWORD_LENGTH })}
                </p>
              </div>

              <div className="mb-6">
                <label
                  htmlFor="setup-confirm-password"
                  className="text-xs font-semibold text-[var(--color-text-muted)]"
                >
                  {t('password.confirm.label')}
                </label>
                <input
                  id="setup-confirm-password"
                  type={showPassword ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                    setConfirmPassword(e.target.value)
                  }
                  className="mt-1 w-full rounded-xl border border-[var(--color-surface-border)] bg-[var(--color-surface-base)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30"
                  placeholder={t('password.confirm.placeholder')}
                  required={true}
                />
              </div>
            </>
          )}

          {/* Error display */}
          {error ? (
            <div
              role="alert"
              className="mb-4 p-3 rounded-xl bg-[var(--color-status-error)]/10 border border-[var(--color-status-error)]/20 text-sm text-[var(--color-status-error)]"
            >
              {error}
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
        <p className="text-xs text-[var(--color-text-muted)] text-center mt-4">
          {t('footer.copyright', { year: new Date().getFullYear() })}
        </p>
      </div>
    </div>
  );
}
