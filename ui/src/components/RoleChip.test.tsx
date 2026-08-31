/**
 * RoleChip tests.
 *
 * The chip changes what the box IS — a Test Master runs tests, a Reflector
 * bounces frames — so switching by accident, or switching twice while one
 * switch is in flight, is expensive. The guards around that are what these
 * assert, along with the failure surface when the backend refuses.
 */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RoleContextValue } from '../contexts/RoleContext';
import { RoleChip } from './RoleChip';

const { roleContext } = vi.hoisted(() => ({
  roleContext: { current: null as unknown as RoleContextValue },
}));

vi.mock('../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/RoleContext')>()),
  useRole: () => roleContext.current,
}));

function renderChip(overrides: Partial<RoleContextValue> = {}) {
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
  render(<RoleChip />);
  return { setRole, clearRoleSwitchError };
}

describe('RoleChip', () => {
  afterEach(cleanup);

  it('marks the active role for assistive technology', () => {
    renderChip();

    expect(screen.getByTestId('role-chip-reflector')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('role-chip-test_master')).toHaveAttribute('aria-pressed', 'false');
  });

  it('does not switch without a confirmation', async () => {
    const { setRole } = renderChip();

    await userEvent.click(screen.getByTestId('role-chip-test_master'));

    // The modal is up, but nothing has changed yet — the whole point of the
    // chip is that a stray click cannot cancel a running test.
    expect(screen.getByTestId('confirm-modal-confirm')).toBeInTheDocument();
    expect(setRole).not.toHaveBeenCalled();
  });

  it('switches to the role the operator confirmed', async () => {
    const { setRole } = renderChip();

    await userEvent.click(screen.getByTestId('role-chip-test_master'));
    await userEvent.click(screen.getByTestId('confirm-modal-confirm'));

    expect(setRole).toHaveBeenCalledWith('test_master');
  });

  it('leaves the role alone when the operator cancels', async () => {
    const { setRole } = renderChip();

    await userEvent.click(screen.getByTestId('role-chip-test_master'));
    await userEvent.click(screen.getByTestId('confirm-modal-cancel'));

    expect(setRole).not.toHaveBeenCalled();
    expect(screen.queryByTestId('confirm-modal-confirm')).toBeNull();
  });

  it('ignores a click on the role already in effect', async () => {
    const { setRole } = renderChip();

    await userEvent.click(screen.getByTestId('role-chip-reflector'));

    expect(screen.queryByTestId('confirm-modal-confirm')).toBeNull();
    expect(setRole).not.toHaveBeenCalled();
  });

  it('refuses a second switch while one is in flight', async () => {
    const { setRole } = renderChip({ isSwitchingRole: true });

    // Asserted on the disabled attribute, which is what actually stops the
    // click. handleClick also returns early on isSwitchingRole, but a click on
    // a disabled button never reaches it -- so asserting "no modal appeared"
    // alone passes with that guard removed and proves nothing.
    expect(screen.getByTestId('role-chip-test_master')).toBeDisabled();
    expect(screen.getByTestId('role-chip-reflector')).toBeDisabled();

    await userEvent.click(screen.getByTestId('role-chip-test_master'));
    expect(setRole).not.toHaveBeenCalled();
  });

  it('shows a spinner while the switch is in flight', () => {
    renderChip({ isSwitchingRole: true });

    expect(screen.getByTestId('role-chip-spinner')).toBeInTheDocument();
    expect(screen.getByLabelText(/role/i)).toHaveAttribute('aria-busy', 'true');
  });

  it('surfaces a rejected switch as an alert', () => {
    renderChip({ roleSwitchError: 'reflector dataplane unavailable' });

    const error = screen.getByTestId('role-chip-error');
    expect(error).toHaveAttribute('role', 'alert');
    expect(error).toHaveTextContent('reflector dataplane unavailable');
  });

  it('lets the operator dismiss the error', async () => {
    const { clearRoleSwitchError } = renderChip({ roleSwitchError: 'nope' });

    await userEvent.click(screen.getByTestId('role-chip-error-dismiss'));

    expect(clearRoleSwitchError).toHaveBeenCalledTimes(1);
  });

  it('shows no error tag when nothing failed', () => {
    renderChip();

    expect(screen.queryByTestId('role-chip-error')).toBeNull();
  });
});
