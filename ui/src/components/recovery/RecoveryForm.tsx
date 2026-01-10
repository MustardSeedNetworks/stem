/**
 * @fileoverview Password Recovery Form Component
 * @description Allows admin to recover password using filesystem-based token.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { ArrowLeft, Eye, EyeOff, KeyRound, Lock, Timer } from 'lucide-react';
import type { FormEvent, ReactElement } from 'react';
import { useEffect, useState } from 'react';

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

/**
 * RecoveryForm Component
 *
 * Form for recovering admin password using filesystem access.
 * User must have SSH/filesystem access to read the recovery token.
 */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Form component with multiple UI states
export function RecoveryForm({
  onRecoveryComplete,
  onBackToLogin,
  remainingTime: initialRemainingTime = 0,
  tokenFilePath = '',
}: RecoveryFormProps): ReactElement {
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
          const data = (await response.json()) as RecoveryInstructions;
          setInstructions(data);
        }
      } catch {
        // Instructions are optional, don't error
      }
    };
    void fetchInstructions();
  }, []);

  // Countdown timer for token expiry
  useEffect(() => {
    if (remainingTime <= 0) return;

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
  const canSubmit = token.trim() && passwordValid && passwordsMatch && !isSubmitting;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);

    if (!passwordValid) {
      setError(`Password must be at least ${MIN_PASSWORD_LENGTH} characters`);
      return;
    }

    if (!passwordsMatch) {
      setError('Passwords do not match');
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

      const data = (await response.json()) as {
        success?: boolean;
        message?: string;
        error?: string;
      };

      if (response.ok && data.success) {
        onRecoveryComplete();
      } else {
        setError(data.message ?? data.error ?? 'Recovery failed');
      }
    } catch {
      setError('Unable to reach server. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="w-16 h-16 mx-auto mb-4 flex items-center justify-center rounded-2xl bg-[var(--color-status-warning)] text-white">
            <KeyRound className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">Password Recovery</h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            Reset your password using filesystem access.
          </p>
        </div>

        {/* Timer Warning */}
        {remainingTime > 0 && (
          <div
            className={`mb-4 p-3 rounded-xl border flex items-center justify-center gap-2 ${
              remainingTime < 120
                ? 'bg-[var(--color-status-warning)]/10 border-[var(--color-status-warning)]/20 text-[var(--color-status-warning)]'
                : 'bg-[var(--color-status-info)]/10 border-[var(--color-status-info)]/20 text-[var(--color-status-info)]'
            }`}
          >
            <Timer className="w-4 h-4" />
            <span className="text-sm">Time remaining: {formatTime(remainingTime)}</span>
          </div>
        )}

        {/* Instructions Panel */}
        {instructions && (
          <div className="mb-4 p-4 rounded-xl bg-[var(--color-surface-sunken)] border border-[var(--color-surface-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)] mb-2">
              Recovery Instructions
            </h3>
            <ol className="text-xs text-[var(--color-text-muted)] space-y-1 list-decimal list-inside">
              {instructions.steps.map((step) => (
                <li key={step}>{step}</li>
              ))}
            </ol>
            {tokenFilePath && (
              <p className="text-xs text-[var(--color-text-muted)] mt-2">
                Token file:{' '}
                <code className="font-mono bg-[var(--color-surface-base)] px-1 rounded">
                  {tokenFilePath}
                </code>
              </p>
            )}
          </div>
        )}

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="rounded-3xl border border-[var(--color-surface-border)] bg-[var(--color-surface-raised)] p-6 shadow-2xl"
        >
          {/* Token Input */}
          <div className="mb-4">
            <label
              htmlFor="recovery-token"
              className="text-xs font-semibold text-[var(--color-text-muted)]"
            >
              Recovery Token
            </label>
            <div className="relative mt-1">
              <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
              <input
                id="recovery-token"
                type="text"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                className="w-full rounded-xl border border-[var(--color-surface-border)] bg-[var(--color-surface-base)] pl-10 pr-3 py-2 text-sm text-[var(--color-text-primary)] font-mono focus:border-[var(--color-brand-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30"
                placeholder="Paste token from .recovery-token file"
                autoComplete="off"
                spellCheck={false}
                required
              />
            </div>
          </div>

          {/* New Password Input */}
          <div className="mb-4">
            <label
              htmlFor="recovery-password"
              className="text-xs font-semibold text-[var(--color-text-muted)]"
            >
              New Password
            </label>
            <div className="relative mt-1">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
              <input
                id="recovery-password"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={`w-full rounded-xl border bg-[var(--color-surface-base)] pl-10 pr-10 py-2 text-sm text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30 ${
                  password && !passwordValid
                    ? 'border-[var(--color-status-error)] focus:border-[var(--color-status-error)]'
                    : 'border-[var(--color-surface-border)] focus:border-[var(--color-brand-primary)]'
                }`}
                placeholder="Enter new password"
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            <p
              className={`text-xs mt-1 ${password && !passwordValid ? 'text-[var(--color-status-error)]' : 'text-[var(--color-text-muted)]'}`}
            >
              Minimum {MIN_PASSWORD_LENGTH} characters
            </p>
          </div>

          {/* Confirm Password Input */}
          <div className="mb-6">
            <label
              htmlFor="recovery-confirm-password"
              className="text-xs font-semibold text-[var(--color-text-muted)]"
            >
              Confirm Password
            </label>
            <div className="relative mt-1">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
              <input
                id="recovery-confirm-password"
                type={showConfirmPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className={`w-full rounded-xl border bg-[var(--color-surface-base)] pl-10 pr-10 py-2 text-sm text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30 ${
                  confirmPassword && !passwordsMatch
                    ? 'border-[var(--color-status-error)] focus:border-[var(--color-status-error)]'
                    : 'border-[var(--color-surface-border)] focus:border-[var(--color-brand-primary)]'
                }`}
                placeholder="Confirm new password"
                required
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
                aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
              >
                {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
            {confirmPassword && !passwordsMatch && (
              <p className="text-xs mt-1 text-[var(--color-status-error)]">
                Passwords do not match
              </p>
            )}
          </div>

          {/* Error display */}
          {error && (
            <div
              role="alert"
              className="mb-4 p-3 rounded-xl bg-[var(--color-status-error)]/10 border border-[var(--color-status-error)]/20 text-sm text-[var(--color-status-error)]"
            >
              {error}
            </div>
          )}

          {/* Submit button */}
          <button
            type="submit"
            disabled={!canSubmit}
            className="btn btn-primary w-full justify-center disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Resetting Password...' : 'Reset Password'}
          </button>

          {/* Back to Login */}
          <button
            type="button"
            onClick={onBackToLogin}
            className="w-full mt-3 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] flex items-center justify-center gap-2"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Login
          </button>
        </form>

        {/* Security Note */}
        <p className="text-xs text-[var(--color-text-muted)] text-center mt-4">
          Recovery tokens are single-use and expire after 15 minutes.
        </p>
      </div>
    </div>
  );
}
