/**
 * Account Security page tests.
 *
 * Drives the page against a stubbed MFA API: what the operator is told about
 * their second factor, the enrol and disable flows end to end (including the
 * TOTP setup modal), and that every refusal is rendered rather than lost
 * (#824).
 */

import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { registerPasskey } from '../../../lib/webauthn-register';
import { MFAError, type MFAStatusResponse, mfaApi } from './mfaApi';
import { SecurityPage } from './SecurityPage';

vi.mock('../../../lib/webauthn-register', () => ({
  registerPasskey: vi.fn(),
}));

vi.mock('./mfaApi', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./mfaApi')>();
  return {
    ...actual,
    mfaApi: {
      status: vi.fn(),
      totpSetup: vi.fn(),
      totpVerify: vi.fn(),
      totpDisable: vi.fn(),
    },
  };
});

const SETUP = {
  secret: 'JBSWY3DPEHPK3PXP',
  provisioningUri: 'otpauth://totp/stem:admin',
  qrCodePngBase64: 'iVBO',
};

function status(over: Partial<MFAStatusResponse> = {}): MFAStatusResponse {
  return { totpEnabled: false, webauthnRegistered: false, webauthnCredentialCount: 0, ...over };
}

async function renderLoaded(initial: MFAStatusResponse = status()) {
  vi.mocked(mfaApi.status).mockResolvedValue(initial);
  render(<SecurityPage />);
  await screen.findByRole('heading', { name: 'Two-factor authentication' });
}

function enterCode(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

beforeEach(() => {
  vi.mocked(mfaApi.status).mockReset();
  vi.mocked(mfaApi.totpSetup).mockReset();
  vi.mocked(mfaApi.totpVerify).mockReset();
  vi.mocked(mfaApi.totpDisable).mockReset();
  vi.mocked(registerPasskey).mockReset();
});

describe('SecurityPage status', () => {
  it('shows loading until the status answers', async () => {
    vi.mocked(mfaApi.status).mockReturnValue(new Promise(() => undefined));
    render(<SecurityPage />);

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('reports a disabled factor and offers to enable it', async () => {
    await renderLoaded(status({ webauthnCredentialCount: 2 }));

    expect(screen.getByText('Disabled')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Enable authenticator app' })).toBeInTheDocument();
    expect(screen.getByTestId('passkey-count')).toHaveTextContent('Registered passkeys: 2');
  });

  it('reports an enabled factor and offers to disable it', async () => {
    await renderLoaded(status({ totpEnabled: true }));

    expect(screen.getByText('Enabled')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Disable authenticator app' })).toBeInTheDocument();
  });

  it('renders a status failure as an alert', async () => {
    vi.mocked(mfaApi.status).mockRejectedValue(new MFAError(500, 'status store offline'));
    render(<SecurityPage />);

    expect(await screen.findByRole('alert')).toHaveTextContent('status store offline');
    expect(screen.getByTestId('passkey-count')).toHaveTextContent('Registered passkeys: 0');
  });
});

describe('enrolling TOTP', () => {
  it('shows the secret, verifies the code and refreshes the status', async () => {
    vi.mocked(mfaApi.totpSetup).mockResolvedValue(SETUP);
    vi.mocked(mfaApi.totpVerify).mockResolvedValue({ success: true, totpEnabled: true });
    await renderLoaded();
    vi.mocked(mfaApi.status).mockResolvedValue(status({ totpEnabled: true }));

    fireEvent.click(screen.getByRole('button', { name: 'Enable authenticator app' }));
    const dialog = await screen.findByRole('dialog', { name: 'Set up authenticator app' });
    expect(within(dialog).getByText(SETUP.secret)).toBeInTheDocument();
    expect(within(dialog).getByAltText('TOTP QR code')).toHaveAttribute(
      'src',
      `data:image/png;base64,${SETUP.qrCodePngBase64}`,
    );

    enterCode('Verification code', '123456');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Verify and enable' }));

    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    expect(mfaApi.totpVerify).toHaveBeenCalledWith('123456');
    expect(await screen.findByText('Enabled')).toBeInTheDocument();
  });

  it('rejects a malformed code without calling the daemon', async () => {
    vi.mocked(mfaApi.totpSetup).mockResolvedValue(SETUP);
    await renderLoaded();

    fireEvent.click(screen.getByRole('button', { name: 'Enable authenticator app' }));
    const dialog = await screen.findByRole('dialog');
    enterCode('Verification code', '12ab');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Verify and enable' }));

    expect(screen.getByLabelText('Verification code')).toBeInvalid();
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    expect(mfaApi.totpVerify).not.toHaveBeenCalled();
  });

  it('keeps the modal open and shows a refused code', async () => {
    vi.mocked(mfaApi.totpSetup).mockResolvedValue(SETUP);
    vi.mocked(mfaApi.totpVerify).mockRejectedValue(new MFAError(401, 'invalid code'));
    await renderLoaded();

    fireEvent.click(screen.getByRole('button', { name: 'Enable authenticator app' }));
    const dialog = await screen.findByRole('dialog');
    enterCode('Verification code', '654321');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Verify and enable' }));

    expect(await within(dialog).findByText('invalid code')).toBeInTheDocument();
  });

  it('closes the modal on cancel without enrolling', async () => {
    vi.mocked(mfaApi.totpSetup).mockResolvedValue(SETUP);
    await renderLoaded();

    fireEvent.click(screen.getByRole('button', { name: 'Enable authenticator app' }));
    const dialog = await screen.findByRole('dialog');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(mfaApi.totpVerify).not.toHaveBeenCalled();
  });

  it('renders a refused setup request', async () => {
    vi.mocked(mfaApi.totpSetup).mockRejectedValue(new MFAError(409, 'already enrolled'));
    await renderLoaded();

    fireEvent.click(screen.getByRole('button', { name: 'Enable authenticator app' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('already enrolled');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });
});

describe('disabling TOTP', () => {
  async function openDisable() {
    await renderLoaded(status({ totpEnabled: true }));
    fireEvent.click(screen.getByRole('button', { name: 'Disable authenticator app' }));
    return screen.getByRole('dialog', { name: 'Disable two-factor authentication' });
  }

  it('sends password and code, then shows the factor disabled', async () => {
    vi.mocked(mfaApi.totpDisable).mockResolvedValue({ success: true, totpEnabled: false });
    const dialog = await openDisable();
    vi.mocked(mfaApi.status).mockResolvedValue(status());

    enterCode('Current password', 'hunter2');
    enterCode('Current code', '000111');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Disable' }));

    expect(await screen.findByText('Disabled')).toBeInTheDocument();
    expect(mfaApi.totpDisable).toHaveBeenCalledWith('hunter2', '000111');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('requires both fields before calling the daemon', async () => {
    const dialog = await openDisable();

    fireEvent.click(within(dialog).getByRole('button', { name: 'Disable' }));

    expect(await within(dialog).findByText('Password is required')).toBeInTheDocument();
    expect(within(dialog).getByText('Code must be exactly 6 digits')).toBeInTheDocument();
    expect(mfaApi.totpDisable).not.toHaveBeenCalled();
  });

  it('shows a refused disable and stays open', async () => {
    vi.mocked(mfaApi.totpDisable).mockRejectedValue(new MFAError(403, 'wrong password'));
    const dialog = await openDisable();

    enterCode('Current password', 'nope');
    enterCode('Current code', '000111');
    fireEvent.click(within(dialog).getByRole('button', { name: 'Disable' }));

    expect(await within(dialog).findByText('wrong password')).toBeInTheDocument();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('cancels back to the enabled card', async () => {
    const dialog = await openDisable();

    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Disable authenticator app' })).toBeInTheDocument();
  });
});

describe('adding a passkey', () => {
  it('confirms the registration and refreshes the count', async () => {
    let finish: () => void = () => undefined;
    vi.mocked(registerPasskey).mockReturnValue(
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
    );
    await renderLoaded();
    vi.mocked(mfaApi.status).mockResolvedValue(status({ webauthnCredentialCount: 1 }));

    fireEvent.click(screen.getByRole('button', { name: 'Add a passkey' }));
    expect(await screen.findByRole('button', { name: 'Waiting for browser...' })).toBeDisabled();
    finish();

    expect(await screen.findByText('Passkey registered.')).toBeInTheDocument();
    expect(screen.getByTestId('passkey-count')).toHaveTextContent('Registered passkeys: 1');
    expect(screen.getByRole('button', { name: 'Add a passkey' })).toBeEnabled();
  });

  it('renders a failed registration and re-enables the button', async () => {
    vi.mocked(registerPasskey).mockRejectedValue(new Error('The operation was aborted'));
    await renderLoaded();

    fireEvent.click(screen.getByRole('button', { name: 'Add a passkey' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('The operation was aborted');
    expect(screen.getByRole('button', { name: 'Add a passkey' })).toBeEnabled();
    expect(screen.queryByText('Passkey registered.')).not.toBeInTheDocument();
  });
});
