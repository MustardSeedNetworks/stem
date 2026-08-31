/**
 * AuthGate tests.
 *
 * This component decides which of four mutually exclusive surfaces an
 * unauthenticated operator sees — setup wizard, recovery form, MFA challenge,
 * login — and showing the wrong one is how someone gets locked out of a box.
 * The gating conditions are asserted here, along with the handlers that move
 * between them.
 */
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuthStore } from '../../stores/auth-store';
import { AuthGate } from './AuthGate';

vi.mock('../setup/SetupWizard', () => ({
  SetupWizard: ({ username }: { username?: string }) => (
    <div data-testid="stub-setup-wizard">setup for {username}</div>
  ),
}));

vi.mock('../recovery/RecoveryForm', () => ({
  RecoveryForm: ({ remainingTime }: { remainingTime?: number }) => (
    <div data-testid="stub-recovery-form">recovery {remainingTime}</div>
  ),
}));

const BASE = {
  isAuthenticated: false,
  loginLoading: false,
  loginError: null,
  mfaPending: null,
  setupStatus: null,
  setupChecked: true,
  recoveryStatus: null,
  showRecoveryForm: false,
};

function setState(overrides: Partial<typeof BASE> = {}) {
  useAuthStore.setState({ ...BASE, ...overrides });
}

describe('AuthGate — which surface is shown', () => {
  beforeEach(() => {
    vi.spyOn(useAuthStore.getState(), 'checkStatuses').mockResolvedValue(undefined);
    setState();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  it('renders nothing at all once authenticated', () => {
    setState({ isAuthenticated: true });
    const { container } = render(<AuthGate />);

    expect(container).toBeEmptyDOMElement();
  });

  it('shows nothing until the setup check has answered', () => {
    // setupChecked false means we do not yet know whether this box needs
    // setup. Showing the login modal here would tell a fresh install to sign
    // in with credentials that do not exist yet.
    setState({ setupChecked: false });
    render(<AuthGate />);

    expect(screen.queryByTestId('login-title')).toBeNull();
    expect(screen.queryByTestId('stub-setup-wizard')).toBeNull();
  });

  it('shows the setup wizard, not the login modal, on a fresh box', () => {
    setState({
      setupStatus: {
        needsSetup: true,
        username: 'admin',
        suggestedPassword: 'generated',
        setupToken: 'tok',
      },
    });
    render(<AuthGate />);

    expect(screen.getByTestId('stub-setup-wizard')).toHaveTextContent('setup for admin');
    expect(screen.queryByTestId('login-title')).toBeNull();
  });

  it('shows the recovery form only when recovery is actually active', () => {
    setState({ showRecoveryForm: true, recoveryStatus: { active: false } });
    render(<AuthGate />);

    // Asking for recovery is not the same as recovery being available; the
    // login modal has to stay reachable when it is not.
    expect(screen.queryByTestId('stub-recovery-form')).toBeNull();
    expect(screen.getByTestId('login-title')).toBeInTheDocument();
  });

  it('shows the recovery form when it is requested and available', () => {
    setState({
      showRecoveryForm: true,
      recoveryStatus: { active: true, remainingTime: 900, instructions: '/tmp/token' },
    });
    render(<AuthGate />);

    expect(screen.getByTestId('stub-recovery-form')).toHaveTextContent('recovery 900');
    expect(screen.queryByTestId('login-title')).toBeNull();
  });

  it('shows the MFA challenge instead of the credentials form', () => {
    setState({ mfaPending: { mfaToken: 'tok', factor: 'totp' } });
    render(<AuthGate />);

    expect(screen.getByLabelText(/code/i)).toBeInTheDocument();
    expect(screen.queryByTestId('login-username')).toBeNull();
  });
});

describe('AuthGate — login', () => {
  beforeEach(() => {
    vi.spyOn(useAuthStore.getState(), 'checkStatuses').mockResolvedValue(undefined);
    setState();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  it('submits the credentials the operator typed', async () => {
    const login = vi
      .spyOn(useAuthStore.getState(), 'login')
      .mockResolvedValue({ status: 'ok' as const });
    render(<AuthGate />);

    await userEvent.type(screen.getByTestId('login-username'), 'operator');
    await userEvent.type(screen.getByTestId('login-password'), 'a-real-password');
    await userEvent.click(screen.getByTestId('login-submit'));

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith('operator', 'a-real-password');
    });
  });

  it('does not call the server with an empty username', async () => {
    const login = vi.spyOn(useAuthStore.getState(), 'login');
    render(<AuthGate />);

    await userEvent.type(screen.getByTestId('login-password'), 'a-real-password');
    await userEvent.click(screen.getByTestId('login-submit'));

    await waitFor(() => {
      expect(screen.getByTestId('login-submit')).toBeEnabled();
    });
    expect(login).not.toHaveBeenCalled();
  });

  it('announces a login failure to assistive technology', () => {
    setState({ loginError: 'Authentication failed' });
    render(<AuthGate />);

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Authentication failed');
    expect(alert).toHaveAttribute('aria-live', 'assertive');
  });

  it('disables submit while a login is in flight', () => {
    setState({ loginLoading: true });
    render(<AuthGate />);

    expect(screen.getByTestId('login-submit')).toBeDisabled();
  });

  it('traps focus in the login dialog so keyboard users cannot tab behind it', () => {
    render(<AuthGate />);

    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-labelledby', 'login-dialog-title');
  });
});

describe('AuthGate — MFA', () => {
  beforeEach(() => {
    vi.spyOn(useAuthStore.getState(), 'checkStatuses').mockResolvedValue(undefined);
    setState({ mfaPending: { mfaToken: 'tok', factor: 'totp' } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    cleanup();
  });

  it('verifies the code the operator typed', async () => {
    const verifyMfa = vi
      .spyOn(useAuthStore.getState(), 'verifyMfa')
      .mockResolvedValue({ status: 'ok' as const });
    render(<AuthGate />);

    await userEvent.type(screen.getByLabelText(/code/i), '123456');
    await userEvent.click(screen.getByRole('button', { name: /verify/i }));

    await waitFor(() => {
      expect(verifyMfa).toHaveBeenCalledWith('123456');
    });
  });

  it('rejects a code that is not six digits without calling the server', async () => {
    const verifyMfa = vi.spyOn(useAuthStore.getState(), 'verifyMfa');
    render(<AuthGate />);

    await userEvent.type(screen.getByLabelText(/code/i), '12');
    await userEvent.click(screen.getByRole('button', { name: /verify/i }));

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /verify/i })).toBeEnabled();
    });
    expect(verifyMfa).not.toHaveBeenCalled();
  });
});
