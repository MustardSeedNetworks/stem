/**
 * RoleGuard tests.
 *
 * RoleGuard is what tells an operator that the stem is in the wrong mode for
 * the module they are looking at. It deliberately still renders its children
 * rather than hiding them, so the failure mode it guards against is a silent
 * one: a page that looks usable while the stem cannot actually run it. These
 * tests pin both halves — the banner appears exactly when the role does not
 * match, and switching goes through the same confirmation the header uses
 * rather than changing the stem's mode on a single click.
 *
 * useRole is stubbed rather than wrapping in a real RoleProvider: the provider
 * talks to /api/v1/mode, and this component's contract is with the context
 * value, not with the endpoint.
 */

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { StemRole } from '../contexts/RoleContext';
import { RoleGuard } from './RoleGuard';

const setRole = vi.fn();
let currentRole: StemRole = 'reflector';

vi.mock('../contexts/RoleContext', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../contexts/RoleContext')>();

  return {
    ...actual,
    useRole: () => ({
      role: currentRole,
      setRole,
      isSwitchingRole: false,
      roleSwitchError: null,
      clearRoleSwitchError: vi.fn(),
    }),
  };
});

beforeEach(() => {
  setRole.mockClear();
  currentRole = 'reflector';
});

describe('when the role already matches', () => {
  it('renders the children with no banner', () => {
    currentRole = 'test_master';

    render(
      <RoleGuard requires="test_master" moduleName="RFC 2544">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByText('module body')).toBeInTheDocument();
    expect(screen.queryByTestId('role-guard-banner')).not.toBeInTheDocument();
  });
});

describe('when the role does not match', () => {
  it('shows the banner and still renders the children', () => {
    render(
      <RoleGuard requires="test_master" moduleName="RFC 2544">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByTestId('role-guard-banner')).toBeInTheDocument();
    // The children stay: hiding them would leave the operator with a blank
    // page and no explanation.
    expect(screen.getByText('module body')).toBeInTheDocument();
  });

  it('names the module in the message so the banner is specific', () => {
    render(
      <RoleGuard requires="test_master" moduleName="RFC 2544">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByTestId('role-guard-banner')).toHaveTextContent(
      'This stem is currently configured as Reflector. Switch to Test Master to run RFC 2544.',
    );
  });

  it('falls back to the role name when no module name is given', () => {
    render(
      <RoleGuard requires="test_master">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByTestId('role-guard-banner')).toHaveTextContent(
      'Switch to Test Master to run Test Master.',
    );
  });

  it('uses the reflector message in the other direction', () => {
    currentRole = 'test_master';

    render(
      <RoleGuard requires="reflector">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByTestId('role-guard-banner')).toHaveTextContent(
      'This stem is currently configured as Test Master. Switch to Reflector to use the loopback reflector.',
    );
  });

  it('exposes the banner to assistive technology as a status', () => {
    render(
      <RoleGuard requires="test_master">
        <p>module body</p>
      </RoleGuard>,
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });
});

describe('switching role from the banner', () => {
  it('does not change the role on the first click — it asks first', async () => {
    const user = userEvent.setup();
    render(
      <RoleGuard requires="test_master" moduleName="RFC 2544">
        <p>module body</p>
      </RoleGuard>,
    );

    await user.click(screen.getByRole('button', { name: 'Switch role' }));

    // Switching cancels any in-progress test, so a single stray click must not
    // do it.
    expect(setRole).not.toHaveBeenCalled();
    expect(screen.getByText('Switch to Test Master?')).toBeInTheDocument();
  });

  it('switches to the required role once confirmed', async () => {
    const user = userEvent.setup();
    render(
      <RoleGuard requires="test_master" moduleName="RFC 2544">
        <p>module body</p>
      </RoleGuard>,
    );

    await user.click(screen.getByRole('button', { name: 'Switch role' }));
    await user.click(screen.getByTestId('confirm-modal-confirm'));

    expect(setRole).toHaveBeenCalledExactlyOnceWith('test_master');
  });

  it('switches to reflector when that is what the module requires', async () => {
    currentRole = 'test_master';
    const user = userEvent.setup();
    render(
      <RoleGuard requires="reflector">
        <p>module body</p>
      </RoleGuard>,
    );

    await user.click(screen.getByRole('button', { name: 'Switch role' }));
    await user.click(screen.getByTestId('confirm-modal-confirm'));

    expect(setRole).toHaveBeenCalledExactlyOnceWith('reflector');
  });

  it('leaves the role alone when the confirmation is cancelled', async () => {
    const user = userEvent.setup();
    render(
      <RoleGuard requires="test_master">
        <p>module body</p>
      </RoleGuard>,
    );

    await user.click(screen.getByRole('button', { name: 'Switch role' }));
    await user.click(screen.getByTestId('confirm-modal-cancel'));

    expect(setRole).not.toHaveBeenCalled();
    expect(screen.queryByText('Switch to Test Master?')).not.toBeInTheDocument();
  });
});
