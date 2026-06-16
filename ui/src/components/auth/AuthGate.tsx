/**
 * @fileoverview AuthGate — unauthenticated overlays (setup / recovery / login).
 * @description Renders the first-run SetupWizard, the password RecoveryForm, and
 *              the login + MFA modal. Self-contained on the auth-store: it owns
 *              the login/MFA forms, the focus trap, and the on-mount status
 *              check. Renders nothing once authenticated. Extracted from App.tsx
 *              during the W5.5 providers+routing decomposition.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import { Lock } from 'lucide-react';
import { type ReactElement, useCallback, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useFocusTrap } from '../../hooks/useFocusTrap';
import { LoginSchema, MfaVerifySchema } from '../../schemas/auth';
import { useAuthStore } from '../../stores/auth-store';
import { RecoveryForm } from '../recovery/RecoveryForm';
import { SetupWizard } from '../setup/SetupWizard';

export function AuthGate(): ReactElement {
  // Authentication state + flows live in the auth-store (tokens are in httpOnly
  // cookies, inaccessible to JS; the store mirrors a boolean flag to
  // localStorage). Render reads state via selectors; handlers call store
  // actions via useAuthStore.getState().
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const loginLoading = useAuthStore((s) => s.loginLoading);
  const loginError = useAuthStore((s) => s.loginError);
  const mfaPending = useAuthStore((s) => s.mfaPending);
  const setupStatus = useAuthStore((s) => s.setupStatus);
  const setupChecked = useAuthStore((s) => s.setupChecked);
  const recoveryStatus = useAuthStore((s) => s.recoveryStatus);
  const showRecoveryForm = useAuthStore((s) => s.showRecoveryForm);
  const setShowRecoveryForm = useAuthStore((s) => s.setShowRecoveryForm);

  // react-hook-form instances for login + MFA verify (#332).
  const loginForm = useForm<{ username: string; password: string }>({
    resolver: valibotResolver(LoginSchema),
    defaultValues: { username: '', password: '' },
    mode: 'onBlur',
  });
  const mfaForm = useForm<{ code: string }>({
    resolver: valibotResolver(MfaVerifySchema),
    defaultValues: { code: '' },
    mode: 'onBlur',
  });

  // Focus trap for login modal (no onEscape - user must authenticate)
  const loginModalRef = useFocusTrap<HTMLDivElement>({
    isActive: !isAuthenticated,
    autoFocus: true,
    restoreFocus: false, // No element to restore to after login
  });

  // Check setup + recovery status on mount (store action; before auth check).
  useEffect(() => {
    void useAuthStore.getState().checkStatuses();
  }, []);

  // Login + MFA submit handlers dispatch to the auth-store. The store owns
  // loginLoading / loginError / mfaPending / isAuthenticated, so the JSX reacts.
  const handleLogin = useCallback((values: { username: string; password: string }): void => {
    void useAuthStore.getState().login(values.username, values.password);
  }, []);

  const handleMFAVerify = useCallback(
    (values: { code: string }): void => {
      void useAuthStore
        .getState()
        .verifyMfa(values.code)
        .then((result) => {
          if (result.status === 'ok') {
            mfaForm.reset();
          }
        });
    },
    [mfaForm],
  );

  // Silent login used by the setup wizard once it provisions an admin.
  const handleSetupLogin = useCallback(
    (username: string, password: string): Promise<boolean> =>
      useAuthStore.getState().performLogin(username, password),
    [],
  );

  // Setup/recovery completion + back-to-login dispatch to the auth-store.
  const handleSetupComplete = useCallback(() => {
    useAuthStore.getState().setSetupStatus(null);
  }, []);

  const handleRecoveryComplete = useCallback(() => {
    const { setRecoveryStatus, setShowRecoveryForm, setLoginError } = useAuthStore.getState();
    setRecoveryStatus(null);
    setShowRecoveryForm(false);
    setLoginError('Password has been reset. Please sign in with your new password.');
  }, []);

  const handleBackToLogin = useCallback(() => {
    useAuthStore.getState().setShowRecoveryForm(false);
  }, []);

  return (
    <>
      {/* Setup Wizard - shown before login if initial setup required */}
      {setupChecked && setupStatus?.needsSetup ? (
        <SetupWizard
          onComplete={handleSetupComplete}
          onLogin={handleSetupLogin}
          suggestedPassword={setupStatus.suggestedPassword}
          username={setupStatus.username}
          setupToken={setupStatus.setupToken}
        />
      ) : null}

      {/* Recovery Form - shown when user clicks "Forgot Password" and recovery is available */}
      {showRecoveryForm && recoveryStatus?.active ? (
        <RecoveryForm
          onRecoveryComplete={handleRecoveryComplete}
          onBackToLogin={handleBackToLogin}
          remainingTime={recoveryStatus.remainingTime}
          tokenFilePath={recoveryStatus.instructions}
        />
      ) : null}

      {/* Login Modal - shown after setup complete or if setup not needed */}
      {!isAuthenticated && setupChecked && !setupStatus?.needsSetup && !showRecoveryForm ? (
        <div className="fixed inset-0 z-50 flex-center bg-scrim/60 backdrop-blur-sm">
          <div
            ref={loginModalRef}
            role="dialog"
            aria-modal="true"
            aria-labelledby="login-dialog-title"
            className="w-full max-w-md rounded-3xl border border-surface-border bg-surface-raised pad-lg shadow-2xl"
          >
            <h2
              id="login-dialog-title"
              data-testid="login-title"
              className="flex items-center gap-compact heading-3 text-text-primary"
            >
              <Lock className="w-5 h-5 text-brand-primary" />
              Sign in to continue
            </h2>
            <p className="text-sm text-text-muted">Authenticate with your Stem credentials.</p>
            {mfaPending ? (
              <form className="mt-6 stack-lg" onSubmit={mfaForm.handleSubmit(handleMFAVerify)}>
                <p className="text-sm text-text-muted">
                  Enter the code from your authenticator app to continue.
                </p>
                <div>
                  <label htmlFor="stem-login-mfa" className="text-xs font-semibold text-text-muted">
                    Verification code
                  </label>
                  <input
                    id="stem-login-mfa"
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]{6}"
                    autoComplete="one-time-code"
                    {...mfaForm.register('code')}
                    className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm font-mono tracking-widest text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                  />
                  {mfaForm.formState.errors.code ? (
                    <p className="mt-tight text-xs text-status-error">
                      {mfaForm.formState.errors.code.message}
                    </p>
                  ) : null}
                </div>
                {loginError ? (
                  <p role="alert" aria-live="assertive" className="text-xs text-status-error">
                    {loginError}
                  </p>
                ) : null}
                <button
                  type="submit"
                  className="btn btn-primary w-full justify-center"
                  disabled={loginLoading}
                >
                  {loginLoading ? 'Verifying...' : 'Verify'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    useAuthStore.getState().cancelMfa();
                    mfaForm.reset();
                  }}
                  className="w-full mt-inline text-sm text-text-muted hover:text-brand-primary"
                >
                  Use different account
                </button>
              </form>
            ) : (
              <form className="mt-6 stack-lg" onSubmit={loginForm.handleSubmit(handleLogin)}>
                <div>
                  <label
                    htmlFor="stem-login-username"
                    className="text-xs font-semibold text-text-muted"
                  >
                    Username
                  </label>
                  <input
                    id="stem-login-username"
                    data-testid="login-username"
                    type="text"
                    autoComplete="username"
                    {...loginForm.register('username')}
                    className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                  />
                  {loginForm.formState.errors.username ? (
                    <p className="mt-tight text-xs text-status-error">
                      {loginForm.formState.errors.username.message}
                    </p>
                  ) : null}
                </div>
                <div>
                  <label
                    htmlFor="stem-login-password"
                    className="text-xs font-semibold text-text-muted"
                  >
                    Password
                  </label>
                  <input
                    id="stem-login-password"
                    data-testid="login-password"
                    type="password"
                    autoComplete="current-password"
                    {...loginForm.register('password')}
                    className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                  />
                  {loginForm.formState.errors.password ? (
                    <p className="mt-tight text-xs text-status-error">
                      {loginForm.formState.errors.password.message}
                    </p>
                  ) : null}
                </div>
                {loginError ? (
                  <p role="alert" aria-live="assertive" className="text-xs text-status-error">
                    {loginError}
                  </p>
                ) : null}
                <button
                  type="submit"
                  data-testid="login-submit"
                  className="btn btn-primary w-full justify-center"
                  disabled={loginLoading}
                >
                  {loginLoading ? 'Signing in...' : 'Sign In'}
                </button>

                {/* Forgot Password link - only shown when recovery is available */}
                {recoveryStatus?.active ? (
                  <button
                    type="button"
                    onClick={() => setShowRecoveryForm(true)}
                    className="w-full mt-content text-sm text-text-muted hover:text-brand-primary"
                  >
                    Forgot password?
                  </button>
                ) : null}
              </form>
            )}
          </div>
        </div>
      ) : null}
    </>
  );
}
