/**
 * RecoveryForm tests.
 *
 * This is the filesystem-token path back into an account whose password is
 * lost — the one flow that bypasses normal login. It had no tests at all
 * (0 of 15 functions covered).
 *
 * The case that matters most is the one an eyeball review would not notice:
 * the form treats a response as successful only when it is BOTH `response.ok`
 * AND `data.success === true`. A 200 that does not say `success` — a proxy
 * page, a partial failure, a changed contract — must not call
 * `onRecoveryComplete`, because that is what lets the caller in.
 */

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { RecoveryForm } from './RecoveryForm';

const VALID_PASSWORD = 'correct-horse-battery';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

/** Instructions are fetched on mount; most tests do not care about them. */
function noInstructions(): Response {
  return new Response('nope', { status: 404 });
}

let fetchMock: ReturnType<typeof vi.fn>;
let onRecoveryComplete: ReturnType<typeof vi.fn>;
let onBackToLogin: ReturnType<typeof vi.fn>;

function renderForm(remainingTime = 0) {
  return render(
    <RecoveryForm
      onRecoveryComplete={onRecoveryComplete}
      onBackToLogin={onBackToLogin}
      remainingTime={remainingTime}
    />,
  );
}

/** Fill the three fields with a valid, matching set. */
async function fillValid(user: ReturnType<typeof userEvent.setup>, token = '  tok-123  ') {
  await user.type(screen.getByLabelText('Recovery Token'), token);
  await user.type(screen.getByLabelText('New Password'), VALID_PASSWORD);
  await user.type(screen.getByLabelText('Confirm Password'), VALID_PASSWORD);
}

function submitCall(): { url: string; init: RequestInit } | undefined {
  const call = fetchMock.mock.calls.find(([url]) => url === '/api/v1/recovery/complete');
  return call ? { url: call[0] as string, init: call[1] as RequestInit } : undefined;
}

/**
 * Route the two endpoints separately, building a fresh Response per call.
 *
 * A single mockResolvedValue would hand the same Response object to both
 * fetches, and a Response body can only be read once — the instructions fetch
 * would consume it and the submit would then fail as a network error, which is
 * indistinguishable from the behaviour under test.
 */
function mockEndpoints(complete?: () => Response): void {
  fetchMock.mockImplementation((url: string) =>
    Promise.resolve(
      url === '/api/v1/recovery/complete' && complete ? complete() : noInstructions(),
    ),
  );
}

beforeEach(() => {
  onRecoveryComplete = vi.fn();
  onBackToLogin = vi.fn();
  fetchMock = vi.fn();
  vi.stubGlobal('fetch', fetchMock);
  mockEndpoints();
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('recovery instructions', () => {
  it('renders the steps the server reports', async () => {
    fetchMock.mockImplementationOnce(() =>
      Promise.resolve(
        jsonResponse({
          triggerFile: '/var/lib/stem/.recovery',
          tokenFile: '/var/lib/stem/.recovery-token',
          expiryTime: '15m',
          steps: ['touch /var/lib/stem/.recovery', 'read /var/lib/stem/.recovery-token'],
        }),
      ),
    );

    renderForm();

    expect(await screen.findByText('touch /var/lib/stem/.recovery')).toBeInTheDocument();
  });

  it('survives a 200 whose body is not the expected shape', async () => {
    // Regression: the body was cast rather than parsed, so a 200 without
    // `steps` reached instructions.steps.map() and took the whole recovery
    // form down -- the one way back into a locked-out account.
    fetchMock.mockImplementationOnce(() => Promise.resolve(jsonResponse({ unexpected: true })));

    renderForm();

    expect(await screen.findByLabelText('Recovery Token')).toBeInTheDocument();
    expect(screen.queryByText('Recovery Instructions')).not.toBeInTheDocument();
  });

  it('renders the form anyway when instructions cannot be fetched', async () => {
    // Instructions are a convenience; losing them must not block recovery.
    fetchMock.mockImplementationOnce(() => Promise.reject(new Error('offline')));

    renderForm();

    expect(await screen.findByLabelText('Recovery Token')).toBeInTheDocument();
  });
});

describe('submitting a valid form', () => {
  it('posts the trimmed token and the new password', async () => {
    const user = userEvent.setup();
    mockEndpoints(() => jsonResponse({ success: true }));
    renderForm();

    await fillValid(user);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    await waitFor(() => expect(submitCall()).toBeDefined());
    const { init } = submitCall() as { init: RequestInit };
    expect(init.method).toBe('POST');
    // The token is pasted out of a file, so surrounding whitespace is normal
    // and must not reach the server. This asserts the end-to-end property
    // rather than one implementation of it: RecoveryCompleteSchema already
    // applies v.trim(), so the .trim() in the request body is belt-and-braces
    // and removing it alone does not fail this — which is the point. Whichever
    // layer does the trimming, the wire stays clean.
    expect(JSON.parse(init.body as string)).toEqual({
      token: 'tok-123',
      password: VALID_PASSWORD,
    });
  });

  it('calls onRecoveryComplete only when the server says success', async () => {
    const user = userEvent.setup();
    mockEndpoints(() => jsonResponse({ success: true }));
    renderForm();

    await fillValid(user);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    await waitFor(() => expect(onRecoveryComplete).toHaveBeenCalledTimes(1));
  });
});

describe('a response that is not a success', () => {
  it.each([
    [
      '200 without a success field',
      () => jsonResponse({ message: 'partially done' }),
      'partially done',
    ],
    [
      '200 with success:false',
      () => jsonResponse({ success: false, error: 'bad token' }),
      'bad token',
    ],
    [
      '200 with neither message nor error',
      () => jsonResponse({ success: false }),
      'Recovery failed',
    ],
    ['401 with a message', () => jsonResponse({ message: 'token expired' }, 401), 'token expired'],
  ])('%s shows the error and does not complete recovery', async (_label, complete, expected) => {
    const user = userEvent.setup();
    mockEndpoints(complete);
    renderForm();

    await fillValid(user);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    expect(await screen.findByText(expected)).toBeInTheDocument();
    // The half that matters: a non-success answer must never admit the caller.
    expect(onRecoveryComplete).not.toHaveBeenCalled();
  });

  it('reports a network failure without completing recovery', async () => {
    const user = userEvent.setup();
    fetchMock.mockImplementation((url: string) =>
      url === '/api/v1/recovery/complete'
        ? Promise.reject(new Error('connection refused'))
        : Promise.resolve(noInstructions()),
    );
    renderForm();

    await fillValid(user);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    expect(
      await screen.findByText('Unable to reach server. Please try again.'),
    ).toBeInTheDocument();
    expect(onRecoveryComplete).not.toHaveBeenCalled();
  });
});

describe('validation', () => {
  it('does not submit when the passwords do not match', async () => {
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByLabelText('Recovery Token'), 'tok-123');
    await user.type(screen.getByLabelText('New Password'), VALID_PASSWORD);
    await user.type(screen.getByLabelText('Confirm Password'), `${VALID_PASSWORD}-typo`);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    // The message has to reach the operator, not just block the submit: a
    // form that silently does nothing on click reads as broken.
    await waitFor(() => expect(screen.getByText(/do not match/i)).toBeInTheDocument());
    expect(submitCall()).toBeUndefined();
    expect(onRecoveryComplete).not.toHaveBeenCalled();
  });

  it('does not submit a password below the minimum length', async () => {
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByLabelText('Recovery Token'), 'tok-123');
    await user.type(screen.getByLabelText('New Password'), 'short');
    await user.type(screen.getByLabelText('Confirm Password'), 'short');
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    await waitFor(() => expect(submitCall()).toBeUndefined());
  });

  it('does not submit without a token', async () => {
    const user = userEvent.setup();
    renderForm();

    await user.type(screen.getByLabelText('New Password'), VALID_PASSWORD);
    await user.type(screen.getByLabelText('Confirm Password'), VALID_PASSWORD);
    await user.click(screen.getByRole('button', { name: 'Reset Password' }));

    await waitFor(() => expect(submitCall()).toBeUndefined());
  });
});

describe('password visibility', () => {
  it('reveals and re-hides the new password', async () => {
    const user = userEvent.setup();
    renderForm();

    const field = screen.getByLabelText('New Password');
    expect(field).toHaveAttribute('type', 'password');

    await user.click(screen.getAllByRole('button', { name: 'Show password' })[0]);
    expect(field).toHaveAttribute('type', 'text');

    await user.click(screen.getAllByRole('button', { name: 'Hide password' })[0]);
    expect(field).toHaveAttribute('type', 'password');
  });
});

describe('navigation', () => {
  it('goes back to login without touching the recovery endpoint', async () => {
    const user = userEvent.setup();
    renderForm();

    await user.click(screen.getByRole('button', { name: 'Back to Login' }));

    expect(onBackToLogin).toHaveBeenCalledTimes(1);
    expect(submitCall()).toBeUndefined();
  });
});

describe('token expiry countdown', () => {
  it('counts down and stops at zero', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      renderForm(2);

      // The label interpolates the clock into surrounding text, so the value
      // is not its own text node — match the rendered line instead.
      const clock = () => screen.getByText(/Time remaining:/).textContent ?? '';

      await waitFor(() => expect(clock()).toContain('0:02'));

      await vi.advanceTimersByTimeAsync(1000);
      await waitFor(() => expect(clock()).toContain('0:01'));

      // At zero the whole timer block unmounts — the token has expired, so a
      // countdown is no longer meaningful. A clock that kept ticking would
      // still be here reading 0:-1.
      await vi.advanceTimersByTimeAsync(2000);
      await waitFor(() => expect(screen.queryByText(/Time remaining:/)).not.toBeInTheDocument());

      await vi.advanceTimersByTimeAsync(3000);
      expect(screen.queryByText(/Time remaining:/)).not.toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });
});
