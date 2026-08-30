/**
 * SetupWizard tests.
 *
 * This is the only way into a fresh stem, so its failure modes are the
 * expensive ones: a submit that never settles strands the operator on a
 * disabled button with no error and no retry, and a role chosen here is the
 * role the box runs as until someone changes it.
 */
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RoleContextValue } from '../../contexts/RoleContext';
import { SetupWizard } from './SetupWizard';

const { roleContext } = vi.hoisted(() => ({
  roleContext: { current: null as unknown as RoleContextValue },
}));

vi.mock('../../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../contexts/RoleContext')>()),
  useRole: () => roleContext.current,
}));

const GOOD_PASSWORD = 'correct-horse-battery';

function renderWizard(props: Partial<Parameters<typeof SetupWizard>[0]> = {}) {
  const onComplete = vi.fn();
  const onLogin = vi.fn().mockResolvedValue(true);
  const setRole = vi.fn();

  roleContext.current = {
    role: 'reflector',
    setRole,
    isSwitchingRole: false,
    roleSwitchError: null,
    clearRoleSwitchError: vi.fn(),
  };

  render(
    <SetupWizard onComplete={onComplete} onLogin={onLogin} setupToken="token-abc" {...props} />,
  );
  return { onComplete, onLogin, setRole };
}

async function fillPassword(password = GOOD_PASSWORD) {
  await userEvent.type(screen.getByLabelText(/^password$/i), password);
  await userEvent.type(screen.getByLabelText(/confirm password/i), password);
}

function radio(value: string): HTMLInputElement {
  const found = screen.getAllByRole('radio').find((el) => (el as HTMLInputElement).value === value);
  if (!found) {
    throw new Error(`no radio with value ${value}`);
  }
  return found as HTMLInputElement;
}

function submit() {
  return userEvent.click(screen.getByRole('button', { name: /complete setup/i }));
}

describe('SetupWizard — completing setup', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    cleanup();
  });

  it('posts the password with the one-time setup token', async () => {
    renderWizard();
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(vi.mocked(fetch)).toHaveBeenCalledWith(
        '/api/v1/setup/complete',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ password: GOOD_PASSWORD, setupToken: 'token-abc' }),
        }),
      );
    });
  });

  it('persists the role the operator picked, not the one it started on', async () => {
    const { setRole } = renderWizard();
    await userEvent.click(radio('test_master'));
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(setRole).toHaveBeenCalledWith('test_master');
    });
  });

  it('logs in with the new password and hands control back', async () => {
    const { onLogin, onComplete } = renderWizard({ username: 'operator' });
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(onLogin).toHaveBeenCalledWith('operator', GOOD_PASSWORD);
    });
    expect(onComplete).toHaveBeenCalledTimes(1);
  });

  it('still leaves the wizard when setup succeeded but the auto-login did not', async () => {
    const { onComplete } = renderWizard({ onLogin: vi.fn().mockResolvedValue(false) });
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/login failed/i);
    });
    // The password is already set — trapping the operator here would leave
    // them with no way in at all.
    expect(onComplete).toHaveBeenCalledTimes(1);
  });
});

describe('SetupWizard — validation and errors', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    cleanup();
  });

  it('refuses mismatched passwords without calling the server', async () => {
    vi.stubGlobal('fetch', vi.fn());
    renderWizard();
    await userEvent.type(screen.getByLabelText(/^password$/i), GOOD_PASSWORD);
    await userEvent.type(screen.getByLabelText(/confirm password/i), 'something-else-entirely');
    await submit();

    await waitFor(() => {
      expect(screen.getByText(/do not match/i)).toBeInTheDocument();
    });
    // Before the schema fix this submitted anyway, with a body carrying only
    // the setup token and no password at all.
    expect(vi.mocked(fetch)).not.toHaveBeenCalled();
  });

  it('refuses a password below the backend minimum without calling the server', async () => {
    vi.stubGlobal('fetch', vi.fn());
    renderWizard();
    await fillPassword('short');
    await submit();

    await waitFor(() => {
      expect(screen.getByText(/at least 12 characters/i)).toBeInTheDocument();
    });
    expect(vi.mocked(fetch)).not.toHaveBeenCalled();
  });

  it("shows the server's own reason for rejecting the setup", async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        json: async () => ({ error: 'setup token already used' }),
      }),
    );
    renderWizard();
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('setup token already used');
    });
  });

  it('gives the setup request a deadline', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) });
    vi.stubGlobal('fetch', fetchMock);
    renderWizard();
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalled();
    });
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    expect(init.signal).toBeInstanceOf(AbortSignal);
  });

  it('reports a hung server instead of leaving the button disabled forever', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockRejectedValue(new DOMException('The operation timed out.', 'TimeoutError')),
    );
    renderWizard();
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/did not respond/i);
    });
    expect(screen.getByRole('button', { name: /complete setup/i })).toBeEnabled();
  });

  it('reports an unreachable server distinctly from one that did not answer', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    renderWizard();
    await fillPassword();
    await submit();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/unable to reach server/i);
    });
  });
});

describe('SetupWizard — generated password', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    cleanup();
  });

  it('shows the suggested password in the clear so it can be saved', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) }));
    renderWizard({ suggestedPassword: 'generated-passphrase-1' });

    await userEvent.click(radio('generated'));

    // A password the operator has one chance to save is useless behind dots.
    expect(screen.getByText('generated-passphrase-1')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /copy password/i })).toBeInTheDocument();
  });

  it('submits the generated password, not an empty one', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) }));
    renderWizard({ suggestedPassword: 'generated-passphrase-1' });

    await userEvent.click(radio('generated'));
    await submit();

    await waitFor(() => {
      expect(vi.mocked(fetch)).toHaveBeenCalledWith(
        '/api/v1/setup/complete',
        expect.objectContaining({
          body: JSON.stringify({
            password: 'generated-passphrase-1',
            setupToken: 'token-abc',
          }),
        }),
      );
    });
  });

  it('clears the generated value when the operator switches back to a custom one', async () => {
    vi.stubGlobal('fetch', vi.fn());
    renderWizard({ suggestedPassword: 'generated-passphrase-1' });

    await userEvent.click(radio('generated'));
    await userEvent.click(radio('custom'));

    expect((screen.getByLabelText(/^password$/i) as HTMLInputElement).value).toBe('');
    expect((screen.getByLabelText(/confirm password/i) as HTMLInputElement).value).toBe('');
  });

  it('offers no generated option when the server suggested nothing', () => {
    vi.stubGlobal('fetch', vi.fn());
    renderWizard();

    expect(screen.getAllByRole('radio').map((el) => (el as HTMLInputElement).value)).not.toContain(
      'generated',
    );
  });
});
