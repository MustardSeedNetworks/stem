/**
 * useRoleSwitch tests.
 *
 * The switch changes what the box IS — a Test Master runs tests, a Reflector
 * bounces frames — so switching by accident, or switching twice while one
 * switch is in flight, is expensive. The guards around that are what these
 * assert, along with the failure surface when the backend refuses.
 */
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RoleContextValue } from '../contexts/RoleContext';
import i18n from '../i18n';
import { RailProductSlot } from '../shared/RailProductSlot';
import { useRoleSwitch } from './RoleSwitch';

const { roleContext } = vi.hoisted(() => ({
  roleContext: { current: null as unknown as RoleContextValue },
}));

vi.mock('../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/RoleContext')>()),
  useRole: () => roleContext.current,
}));

function Harness() {
  const { productSlot, overlay } = useRoleSwitch();
  return (
    <>
      {overlay}
      <RailProductSlot collapsed={false} {...productSlot} />
    </>
  );
}

function renderSwitch(overrides: Partial<RoleContextValue> = {}) {
  const setRole = vi.fn();
  const clearRoleSwitchError = vi.fn();
  roleContext.current = {
    role: 'reflector',
    setRole,
    isSwitchingRole: false,
    roleSwitchError: null,
    clearRoleSwitchError,
    ...overrides,
  };
  render(<Harness />);
  return { setRole, clearRoleSwitchError };
}

async function choose(name: string) {
  await userEvent.click(screen.getByTestId('rail-product-slot'));
  await userEvent.click(screen.getByRole('menuitemradio', { name }));
}

describe('useRoleSwitch', () => {
  afterEach(cleanup);

  it('shows the active role on the slot', () => {
    renderSwitch();

    expect(screen.getByTestId('rail-product-slot')).toHaveTextContent('Reflector');
  });

  it('does not switch without a confirmation', async () => {
    const { setRole } = renderSwitch();

    await choose('Test Master');

    // The dialog is up, but nothing has changed yet: a stray click cannot
    // cancel a running test.
    expect(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Switch role' }),
    ).toBeInTheDocument();
    expect(setRole).not.toHaveBeenCalled();
  });

  it('switches to the role the operator confirmed', async () => {
    const { setRole } = renderSwitch();

    await choose('Test Master');
    await userEvent.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Switch role' }),
    );

    expect(setRole).toHaveBeenCalledExactlyOnceWith('test_master');
  });

  it('leaves the role alone when the operator cancels', async () => {
    const { setRole } = renderSwitch();

    await choose('Test Master');
    await userEvent.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Cancel' }),
    );

    expect(setRole).not.toHaveBeenCalled();
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('asks nothing when the active role is chosen again', async () => {
    renderSwitch();

    await choose('Reflector');

    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('refuses a second switch while one is in flight', async () => {
    const { setRole } = renderSwitch({ isSwitchingRole: true });

    await choose('Test Master');

    expect(screen.queryByRole('dialog')).toBeNull();
    expect(setRole).not.toHaveBeenCalled();
  });

  it('surfaces a rejected switch as an alert', () => {
    renderSwitch({
      roleSwitchError: { kind: 'failed', detail: 'reflector dataplane unavailable' },
    });

    const error = screen.getByTestId('role-switch-error');
    expect(within(error).getByRole('alert')).toBeInTheDocument();
    // The alert title supplies "Role switch failed:"; the cause must not
    // carry it again.
    expect(error).toHaveTextContent(/Role switch failed:\s*reflector dataplane unavailable/);
  });

  it('names an unreadable reply from the locale, in Spanish too', async () => {
    await i18n.changeLanguage('es');
    try {
      renderSwitch({ roleSwitchError: { kind: 'unexpectedResponse' } });

      expect(screen.getByTestId('role-switch-error')).toHaveTextContent(
        /Fallo el cambio de rol:\s*respuesta inesperada del servidor/,
      );
    } finally {
      cleanup();
      await i18n.changeLanguage('en');
    }
  });

  it('lets the operator dismiss the error', async () => {
    const { clearRoleSwitchError } = renderSwitch({
      roleSwitchError: { kind: 'failed', detail: 'nope' },
    });

    await userEvent.click(
      within(screen.getByTestId('role-switch-error')).getByRole('button', { name: /dismiss/i }),
    );

    expect(clearRoleSwitchError).toHaveBeenCalledTimes(1);
  });

  it('shows no error when nothing failed', () => {
    renderSwitch();

    expect(screen.queryByTestId('role-switch-error')).toBeNull();
  });
});
