/**
 * SettingsDrawer.i18n.test.tsx — the settings drawer renders real locale copy.
 *
 * #759: wrecking every EN locale file failed only 19 of 206 tests. The drawer
 * is the product's main configuration surface and had no copy assertions, so
 * its title, view-mode switch and close control could all regress silently.
 */
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ROLE_STORAGE_KEY, RoleProvider } from '../contexts/RoleContext';
import i18n from '../i18n';
import {
  defaultMEFConfig,
  defaultRFC2544Config,
  defaultRFC2889Config,
  defaultRFC6349Config,
  defaultTSNConfig,
  defaultY1564Config,
  defaultY1731Config,
} from '../types/settings';
import { SettingsDrawer } from './SettingsDrawer';
import { defaultTrafficGenConfig } from './TrafficGenConfigForm';

function renderDrawer() {
  return render(
    <RoleProvider>
      <SettingsDrawer
        isOpen={true}
        onClose={() => {}}
        selectedTests={[]}
        setSelectedTests={() => {}}
        rfc2544Config={defaultRFC2544Config}
        setRFC2544Config={() => {}}
        rfc2889Config={defaultRFC2889Config}
        setRFC2889Config={() => {}}
        rfc6349Config={defaultRFC6349Config}
        setRFC6349Config={() => {}}
        y1564Config={defaultY1564Config}
        setY1564Config={() => {}}
        y1731Config={defaultY1731Config}
        setY1731Config={() => {}}
        tsnConfig={defaultTSNConfig}
        setTSNConfig={() => {}}
        trafficGenConfig={defaultTrafficGenConfig}
        setTrafficGenConfig={() => {}}
        mefConfig={defaultMEFConfig}
        setMEFConfig={() => {}}
      />
    </RoleProvider>,
  );
}

beforeEach(() => {
  // The view-mode switch only renders for test_master (SettingsDrawer:154),
  // and RoleProvider seeds its role from localStorage.
  window.localStorage.setItem(ROLE_STORAGE_KEY, 'test_master');
  // LicenseSection renders inside the drawer and probes /api/license on mount.
  vi.spyOn(globalThis, 'fetch').mockImplementation((async () => ({
    ok: true,
    status: 200,
    json: async () => ({ activated: false, tier: 'free', deviceId: 'dev-1', features: [] }),
  })) as unknown as typeof fetch);
});

afterEach(async () => {
  window.localStorage.removeItem(ROLE_STORAGE_KEY);
  vi.restoreAllMocks();
  await i18n.changeLanguage('en');
});

describe('SettingsDrawer — real locale copy', () => {
  it('renders the English title and view switch', async () => {
    await i18n.changeLanguage('en');
    renderDrawer();

    await waitFor(() => {
      expect(screen.getAllByText('Settings').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('View by:')).toBeInTheDocument();
    expect(screen.getByText('Module')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    renderDrawer();

    await waitFor(() => {
      expect(screen.getAllByText('Configuración').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('Ver por:')).toBeInTheDocument();
    expect(screen.queryByText('View by:')).toBeNull();
  });

  it('localizes the close control for screen readers', async () => {
    await i18n.changeLanguage('es');
    renderDrawer();

    await waitFor(() => {
      expect(screen.getAllByText('Configuración').length).toBeGreaterThan(0);
    });
    // aria-label, not JSX text — invisible to the hardcoded-text detector.
    expect(screen.getAllByLabelText(/Cerrar/i).length).toBeGreaterThan(0);
  });
});
