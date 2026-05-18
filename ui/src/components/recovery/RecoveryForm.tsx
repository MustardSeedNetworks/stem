/**
 * @fileoverview Password Recovery Form Component
 * @description Allows admin to recover password using filesystem-based token.
 */

import { ArrowLeft, Eye, EyeOff, KeyRound, Lock, Timer } from 'lucide-react';
import type { FormEvent, ReactElement } from 'react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  alert,
  button,
  cn,
  icon,
  input,
  layout,
  radius,
  spacing,
  status,
} from '../../styles/theme';

/** Minimum password length (matches backend validation) */
const MIN_PASSWORD_LENGTH = 12;

/** Props for RecoveryForm component */
interface RecoveryFormProps {
  /** Callback when recovery is complete */
  onRecoveryComplete: () => void;
  /** Callback to return to login */
  onBackToLogin: () => void;
  /** Remaining time in seconds */
  remainingTime?: number;
  /** File path instructions */
  tokenFilePath?: string;
}

/** Recovery instructions from API */
interface RecoveryInstructions {
  triggerFile: string;
  tokenFile: string;
  expiryTime: string;
  steps: string[];
}

/** Props for PasswordInput sub-component */
interface PasswordInputProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  showPassword: boolean;
  onToggleVisibility: () => void;
  placeholder: string;
  hasError: boolean;
  showPasswordLabel: string;
  hidePasswordLabel: string;
  helperText?: string;
  errorText?: string;
}

/** Reusable password input with visibility toggle */
function PasswordInput({
  id,
  label,
  value,
  onChange,
  showPassword,
  onToggleVisibility,
  placeholder,
  hasError,
  showPasswordLabel,
  hidePasswordLabel,
  helperText,
  errorText,
}: PasswordInputProps): ReactElement {
  return (
    <div>
      <label htmlFor={id} className={cn('label block', spacing.margin.bottom.inline)}>
        {label}
      </label>
      <div className="relative">
        <Lock
          className={cn(icon.size.sm, 'absolute left-3 top-1/2 -translate-y-1/2 text-text-muted')}
        />
        <input
          id={id}
          type={showPassword ? 'text' : 'password'}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void => onChange(e.target.value)}
          className={cn(
            'w-full pl-10 pr-10',
            input.size.md,
            radius.xl,
            'border bg-surface-base text-text-primary',
            hasError ? 'border-status-error' : 'border-surface-border',
            'focus:outline-none focus:ring-2 focus:ring-brand-primary/30',
            hasError ? 'focus:border-status-error' : 'focus:border-brand-primary',
          )}
          placeholder={placeholder}
          required={true}
        />
        <button
          type="button"
          onClick={onToggleVisibility}
          className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
          title={showPassword ? hidePasswordLabel : showPasswordLabel}
          aria-label={showPassword ? hidePasswordLabel : showPasswordLabel}
        >
          {showPassword ? <EyeOff className={icon.size.sm} /> : <Eye className={icon.size.sm} />}
        </button>
      </div>
      {helperText !== undefined && (
        <p className={cn('caption mt-1', hasError ? status.text.error : 'text-text-muted')}>
          {helperText}
        </p>
      )}
      {errorText !== undefined && hasError && (
        <p className={cn('caption mt-1', status.text.error)}>{errorText}</p>
      )}
    </div>
  );
}

/** Props for InstructionsPanel sub-component */
interface InstructionsPanelProps {
  instructions: RecoveryInstructions;
  tokenFilePath: string;
  title: string;
  tokenFileLabel: string;
}

/** Instructions panel showing recovery steps */
function InstructionsPanel({
  instructions,
  tokenFilePath,
  title,
  tokenFileLabel,
}: InstructionsPanelProps): ReactElement {
  return (
    <div
      className={cn(
        'bg-surface-sunken border border-surface-border',
        radius.xl,
        spacing.pad.default,
        spacing.margin.bottom.content,
      )}
    >
      <h3 className={cn('heading-4', spacing.margin.bottom.inline)}>{title}</h3>
      <ol className="caption text-text-muted space-y-1 list-decimal list-inside">
        {instructions.steps.map((step) => (
          <li key={step}>{step}</li>
        ))}
      </ol>
      {tokenFilePath !== '' && (
        <p className={cn('caption text-text-muted', spacing.margin.top.inline)}>
          {tokenFileLabel}: <code className="code">{tokenFilePath}</code>
        </p>
      )}
    </div>
  );
}

/**
 * RecoveryForm Component
 *
 * Form for recovering admin password using filesystem access.
 * User must have SSH/filesystem access to read the recovery token.
 */
export function RecoveryForm({
  onRecoveryComplete,
  onBackToLogin,
  remainingTime: initialRemainingTime = 0,
  tokenFilePath = '',
}: RecoveryFormProps): ReactElement {
  const { t } = useTranslation('recovery');
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [remainingTime, setRemainingTime] = useState(initialRemainingTime);
  const [instructions, setInstructions] = useState<RecoveryInstructions | null>(null);

  // Fetch recovery instructions on mount
  useEffect(() => {
    const fetchInstructions = async (): Promise<void> => {
      try {
        const response = await fetch('/api/v1/recovery/instructions');
        if (response.ok) {
          const data = await (response.json() as Promise<RecoveryInstructions>);
          setInstructions(data);
        }
      } catch {
        // Instructions are optional, don't error
      }
    };
    fetchInstructions().catch(() => {
      // Already handled silently inside
    });
  }, []);

  // Countdown timer for token expiry
  useEffect(() => {
    if (remainingTime <= 0) {
      return;
    }

    const interval = setInterval(() => {
      setRemainingTime((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [remainingTime]);

  // Format remaining time as MM:SS
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  // Password validation
  const passwordValid = password.length >= MIN_PASSWORD_LENGTH;
  const passwordsMatch = password === confirmPassword;
  const canSubmit = token.trim() !== '' && passwordValid && passwordsMatch && !isSubmitting;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);

    if (!passwordValid) {
      setError(t('errors.passwordTooShort', { count: MIN_PASSWORD_LENGTH }));
      return;
    }

    if (!passwordsMatch) {
      setError(t('errors.passwordMismatch'));
      return;
    }

    setIsSubmitting(true);

    try {
      const response = await fetch('/api/v1/recovery/complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token: token.trim(),
          password,
        }),
      });

      const data = await (response.json() as Promise<{
        success?: boolean;
        message?: string;
        error?: string;
      }>);

      if (response.ok && data.success === true) {
        onRecoveryComplete();
      } else {
        setError(data.message ?? data.error ?? t('errors.recoveryFailed'));
      }
    } catch {
      setError(t('errors.networkError'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      className={cn('fixed inset-0 z-50 bg-black/60 backdrop-blur-sm', layout.flex.center, 'p-4')}
    >
      <div className="w-full max-w-md">
        {/* Header */}
        <div className={cn('text-center', spacing.margin.bottom.section)}>
          <div
            className={cn(
              cn('w-16 h-16 mx-auto rounded-2xl text-text-inverse', status.bg.warning),
              layout.flex.center,
              spacing.margin.bottom.content,
            )}
          >
            <KeyRound className={icon.size.xl} />
          </div>
          <h1 className="heading-2 text-text-primary">{t('title')}</h1>
          <p className={cn('body-small text-text-muted', spacing.margin.top.inline)}>
            {t('subtitle')}
          </p>
        </div>

        {/* Timer Warning */}
        {remainingTime > 0 && (
          <div
            className={cn(
              alert.base,
              remainingTime < 120 ? alert.variant.warning : alert.variant.info,
              spacing.margin.bottom.content,
              layout.flex.center,
            )}
          >
            <Timer className={icon.size.sm} />
            <span className="body-small ml-2">
              {t('timeRemaining')}: {formatTime(remainingTime)}
            </span>
          </div>
        )}

        {/* Instructions Panel */}
        {instructions !== null && (
          <InstructionsPanel
            instructions={instructions}
            tokenFilePath={tokenFilePath}
            title={t('instructions.title')}
            tokenFileLabel={t('instructions.tokenFile')}
          />
        )}

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className={cn(
            'bg-surface-raised border border-surface-border shadow-2xl',
            radius.xl,
            spacing.pad.lg,
            spacing.stack.lg,
          )}
        >
          {/* Token Input */}
          <div>
            <label
              htmlFor="recovery-token"
              className={cn('label block', spacing.margin.bottom.inline)}
            >
              {t('token.label')}
            </label>
            <div className="relative">
              <KeyRound
                className={cn(
                  icon.size.sm,
                  'absolute left-3 top-1/2 -translate-y-1/2 text-text-muted',
                )}
              />
              <input
                id="recovery-token"
                type="text"
                value={token}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  setToken(e.target.value)
                }
                className={cn(
                  'w-full pl-10',
                  input.size.md,
                  radius.xl,
                  'border border-surface-border bg-surface-base text-text-primary',
                  'focus:outline-none focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 font-mono',
                )}
                placeholder={t('token.placeholder')}
                autoComplete="off"
                spellCheck={false}
                required={true}
              />
            </div>
          </div>

          {/* New Password Input */}
          <PasswordInput
            id="recovery-password"
            label={t('password.label')}
            value={password}
            onChange={setPassword}
            showPassword={showPassword}
            onToggleVisibility={() => setShowPassword(!showPassword)}
            placeholder={t('password.placeholder')}
            hasError={password !== '' && !passwordValid}
            showPasswordLabel={t('buttons.showPassword')}
            hidePasswordLabel={t('buttons.hidePassword')}
            helperText={t('password.minLength', { count: MIN_PASSWORD_LENGTH })}
          />

          {/* Confirm Password Input */}
          <PasswordInput
            id="recovery-confirm-password"
            label={t('password.confirmLabel')}
            value={confirmPassword}
            onChange={setConfirmPassword}
            showPassword={showConfirmPassword}
            onToggleVisibility={() => setShowConfirmPassword(!showConfirmPassword)}
            placeholder={t('password.confirmPlaceholder')}
            hasError={confirmPassword !== '' && !passwordsMatch}
            showPasswordLabel={t('buttons.showPassword')}
            hidePasswordLabel={t('buttons.hidePassword')}
            errorText={t('errors.passwordMismatch')}
          />

          {/* Error display */}
          {error !== null && (
            <div role="alert" aria-live="assertive" className={cn(alert.base, alert.variant.error)}>
              {error}
            </div>
          )}

          {/* Submit button */}
          <button
            type="submit"
            disabled={!canSubmit}
            className={cn(
              'w-full',
              button.size.md,
              'bg-brand-primary text-text-inverse',
              radius.md,
              'font-medium hover:bg-brand-accent',
              'focus:outline-none focus:ring-2 focus:ring-brand-primary',
              'focus:ring-offset-2 focus:ring-offset-surface-base',
              'disabled:opacity-50 disabled:cursor-not-allowed',
            )}
          >
            {isSubmitting ? t('buttons.submitting') : t('buttons.submit')}
          </button>

          {/* Back to Login */}
          <button
            type="button"
            onClick={onBackToLogin}
            className={cn(
              'w-full',
              button.size.sm,
              'text-text-muted hover:text-text-primary',
              layout.flex.center,
            )}
          >
            <ArrowLeft className={icon.size.sm} />
            <span className="ml-2">{t('buttons.backToLogin')}</span>
          </button>
        </form>

        {/* Security Note */}
        <p className={cn('caption text-text-muted text-center', spacing.margin.top.content)}>
          {t('securityNote')}
        </p>
      </div>
    </div>
  );
}
