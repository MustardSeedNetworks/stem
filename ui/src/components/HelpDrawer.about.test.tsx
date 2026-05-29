/**
 * HelpDrawer.about.test.tsx — asserts the in-app About/version panel exists.
 *
 * The audit identified "no About/version panel" as the one stem GUI-help gap.
 * PR-C adds a version line in the HelpDrawer header showing `Stem v<version>`
 * sourced from the /__version endpoint via useBuildVersion. This test fails
 * if that surface disappears.
 */
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { HelpDrawer } from './HelpDrawer';

describe('HelpDrawer — About / version surface', () => {
  it('renders the version badge using /__version payload', async () => {
    const payload = {
      version: '0.21.4',
      commit: 'deadbee',
      buildTime: '2026-05-29T00:00:00Z',
      uiBuildHash: 'abc123',
    };
    vi.spyOn(globalThis, 'fetch').mockImplementation((async () => ({
      ok: true,
      json: async () => payload,
    })) as unknown as typeof fetch);

    render(<HelpDrawer isOpen={true} onClose={() => undefined} />);
    const badge = await screen.findByTestId('help-drawer-version');
    expect(badge).toBeInTheDocument();
    await waitFor(() => {
      expect(badge).toHaveTextContent(/v0\.21\.4/);
    });
    expect(badge).toHaveTextContent(/Stem/);
  });

  it('renders a fallback version when /__version is unreachable', async () => {
    vi.spyOn(globalThis, 'fetch').mockImplementation((async () => {
      throw new Error('network');
    }) as unknown as typeof fetch);

    render(<HelpDrawer isOpen={true} onClose={() => undefined} />);
    const badge = await screen.findByTestId('help-drawer-version');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveTextContent(/Stem v/);
  });
});
