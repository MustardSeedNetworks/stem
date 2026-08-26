/**
 * LicenseSection.i18n.test.tsx — the licence panel renders real locale copy.
 *
 * #759: wrecking every EN locale file failed only 19 of 206 tests, because the
 * suite asserts on testids and on English hardcoded in components. These assert
 * visible strings in both locales against the real `ui/locales`
 * JSON, so a key that goes missing or ships untranslated fails here.
 *
 * The error paths are covered too: five of this component's user-visible
 * messages were English string literals passed to setError, which the
 * hardcoded-text detector cannot see — it reads JSX text nodes, not arguments.
 */
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../i18n';
import { LicenseSection } from './LicenseSection';

function stubLicense(body: unknown, ok = true) {
  vi.spyOn(globalThis, 'fetch').mockImplementation((async () => ({
    ok,
    status: ok ? 200 : 500,
    json: async () => body,
  })) as unknown as typeof fetch);
}

beforeEach(() => {
  stubLicense({ activated: false, tier: 'free', deviceId: 'dev-1', features: [] });
});

afterEach(async () => {
  vi.restoreAllMocks();
  await i18n.changeLanguage('en');
});

describe('LicenseSection — real locale copy', () => {
  it('renders the English panel, title included', async () => {
    await i18n.changeLanguage('en');
    render(<LicenseSection />);

    // 'Activate License' is both the form heading and the button label.
    await waitFor(() => {
      expect(screen.getAllByText('Activate License').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('License')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    render(<LicenseSection />);

    await waitFor(() => {
      expect(screen.getAllByText('Activar Licencia').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('Licencia')).toBeInTheDocument();
    expect(screen.queryByText('Activate License')).toBeNull();
    expect(screen.queryByText('License')).toBeNull();
  });

  it('localizes the activation failure, which is a string literal not JSX', async () => {
    await i18n.changeLanguage('es');
    render(<LicenseSection />);

    await waitFor(() => {
      expect(screen.getAllByText('Activar Licencia').length).toBeGreaterThan(0);
    });

    // The activate button is disabled until a key is present, so the failure
    // path is only reachable with one typed in.
    await userEvent.type(screen.getByPlaceholderText('XXXX-XXXX-XXXX-XXXX'), 'AAAA-BBBB-CCCC-DDDD');
    stubLicense({ success: false, message: '' });
    const buttons = screen.getAllByRole('button', { name: /Activar Licencia/ });
    const activate = buttons[buttons.length - 1];
    if (activate) {
      await userEvent.click(activate);
    }

    await waitFor(() => {
      expect(screen.getByText('Error en la activacion de licencia')).toBeInTheDocument();
    });
  });
});
