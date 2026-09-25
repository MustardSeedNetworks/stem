import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, spyOn, userEvent, within } from 'storybook/test';
import { mfaApi } from './mfaApi';
import { SecurityPage } from './SecurityPage';
import { TotpSetupModal } from './TotpSetupModal';

// Storybook has no daemon, so the page's status fetch is answered here. Each
// play function waits for the loaded card before the a11y check runs, or axe
// would only ever see the loading line.
const meta: Meta<typeof SecurityPage> = {
  title: 'Pages/Account/Security',
  component: SecurityPage,
};

export default meta;
type Story = StoryObj<typeof SecurityPage>;

export const TotpDisabled: Story = {
  beforeEach: () => {
    spyOn(mfaApi, 'status').mockResolvedValue({
      totpEnabled: false,
      webauthnRegistered: false,
      webauthnCredentialCount: 0,
    });
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      await canvas.findByRole('button', { name: 'Enable authenticator app' }),
    ).toBeEnabled();
    await expect(canvas.getByRole('button', { name: 'Add a passkey' })).toBeEnabled();
  },
};

export const DisableDialog: Story = {
  beforeEach: () => {
    spyOn(mfaApi, 'status').mockResolvedValue({
      totpEnabled: true,
      webauthnRegistered: true,
      webauthnCredentialCount: 2,
    });
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: 'Disable authenticator app' }));
    const dialog = within(await canvas.findByRole('dialog'));
    await expect(dialog.getByRole('button', { name: 'Cancel' })).toHaveAttribute('type', 'button');
    await expect(dialog.getByRole('button', { name: 'Disable' })).toHaveAttribute('type', 'submit');
  },
};

// 1x1 transparent PNG: the modal only needs a decodable image.
const PIXEL_PNG =
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=';

export const TotpSetup: StoryObj<typeof TotpSetupModal> = {
  render: (args) => <TotpSetupModal {...args} />,
  args: {
    setup: {
      secret: 'JBSWY3DPEHPK3PXP',
      provisioningUri: 'otpauth://totp/Stem:admin?secret=JBSWY3DPEHPK3PXP',
      qrCodePngBase64: PIXEL_PNG,
    },
    onComplete: fn(),
    onCancel: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const dialog = within(within(canvasElement).getByRole('dialog'));
    await userEvent.click(dialog.getByRole('button', { name: 'Cancel' }));
    await expect(args.onCancel).toHaveBeenCalledOnce();
  },
};
