/**
 * ModuleGate.test.tsx — a Pro module on Free says so before the run.
 *
 * Driven through the real LicenseProvider and the wire payload rather than a
 * mocked hook: a mocked `hasFeature` keeps passing while the endpoint sends no
 * `features` at all, which is exactly how seed's gates read `undefined` for a
 * release (seed#2688).
 */
import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { LicenseProvider } from '../contexts/LicenseContext';
import i18n from '../i18n';
import type { LicenseStatus } from '../types/generated/license-status';
import { ModuleGate } from './ModuleGate';

const mockFetch = vi.fn<() => Promise<Response>>();
vi.mock('../stores/auth-store', () => ({
  authFetch: (): Promise<Response> => mockFetch(),
}));

function payload(features: string[]): LicenseStatus {
  return {
    activated: features.length > 0,
    isTrialMode: false,
    tier: features.length > 0 ? 2 : 0,
    tierName: features.length > 0 ? 'Professional' : 'Reflector',
    daysRemaining: 0,
    features,
    deviceHash: 'abc123',
  };
}

function answerWith(features: string[]): void {
  mockFetch.mockResolvedValue({
    ok: true,
    json: () => Promise.resolve(payload(features)),
  } as unknown as Response);
}

function renderGate(
  path: '/tests/benchmark' | '/tests/certify' = '/tests/benchmark',
  preview = true,
): void {
  render(
    <LicenseProvider>
      <ModuleGate path={path} preview={preview}>
        <button type="button">frame size</button>
      </ModuleGate>
    </LicenseProvider>,
  );
}

describe('ModuleGate', () => {
  beforeEach(async () => {
    mockFetch.mockReset();
    await i18n.changeLanguage('en');
  });

  it('pitches the module and leaves the form readable when it is not licensed', async () => {
    answerWith([]);
    renderGate();

    await waitFor(() => expect(screen.getByTestId('module-gate-pitch')).toBeVisible());
    expect(screen.getByText('Benchmark is a Professional module')).toBeVisible();
    expect(screen.getByText(/RFC 2544 throughput, latency, frame loss/)).toBeVisible();
    // The form is the preview: still rendered, still readable, just inert.
    const form = screen.getByText('frame size').parentElement;
    expect(form).toHaveAttribute('inert');
    // Not aria-hidden — an operator decides whether to buy it by reading it.
    expect(form).not.toHaveAttribute('aria-hidden');
    expect(form).toHaveAttribute('aria-describedby', 'module-gate-pitch-benchmark');
  });

  it('leaves children the operator still needs live under the pitch', async () => {
    // preview={false} is the empty state: its Settings link is how an operator
    // configures an interface at all, so an inert page body would strand them.
    answerWith([]);
    renderGate('/tests/benchmark', false);

    await waitFor(() => expect(screen.getByTestId('module-gate-pitch')).toBeVisible());
    const button = screen.getByText('frame size');
    expect(button.closest('[inert]')).toBeNull();
    button.focus();
    expect(document.activeElement).toBe(button);
  });

  it('renders the form alone once the licence grants the module', async () => {
    answerWith(['rfc2544']);
    renderGate();

    await waitFor(() => expect(screen.getByText('frame size')).toBeVisible());
    expect(screen.queryByTestId('module-gate')).toBeNull();
  });

  it('is ungated when the licence grants any one of a module’s standards', async () => {
    // Certify covers RFC 2889, RFC 6349 and TSN. Holding one of the three
    // should show the form, not a pitch for what is already paid for.
    answerWith(['tsn']);
    renderGate('/tests/certify');

    await waitFor(() => expect(screen.getByText('frame size')).toBeVisible());
    expect(screen.queryByTestId('module-gate')).toBeNull();
  });

  it('shows the form, not a pitch, while the status is unknown', () => {
    mockFetch.mockReturnValue(new Promise(() => undefined));
    renderGate();

    // A null flash of an already-visible form is worse than a pitch that
    // arrives a moment late.
    expect(screen.getByText('frame size')).toBeVisible();
    expect(screen.queryByTestId('module-gate')).toBeNull();
  });

  it('does not gate a paying operator when the daemon is unreachable', async () => {
    mockFetch.mockRejectedValue(new Error('network'));
    renderGate();

    await waitFor(() => expect(screen.getByText('frame size')).toBeVisible());
    expect(screen.queryByTestId('module-gate')).toBeNull();
  });

  it('renders the pitch in Spanish', async () => {
    await i18n.changeLanguage('es');
    answerWith([]);
    renderGate();

    await waitFor(() =>
      expect(screen.getByText('Benchmark es un módulo Professional')).toBeVisible(),
    );
    expect(screen.queryByText('Benchmark is a Professional module')).toBeNull();
    // No reset here: beforeEach owns the language, and switching it back while
    // the gate is still mounted re-renders it outside act().
  });
});
