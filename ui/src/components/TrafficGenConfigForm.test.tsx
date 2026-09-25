/**
 * Third of the seven, and the one that carried the summary-heading
 * divergence: this form said "Traffic Summary" while the other six said
 * "Test Summary". The shared TestSummary resolves that to one heading, so
 * the first assertion here is that this form now says what the others say.
 */
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import i18n from 'i18next';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { defaultTrafficGenConfig, TrafficGenConfigForm } from './TrafficGenConfigForm';

function renderForm(overrides: Partial<typeof defaultTrafficGenConfig> = {}) {
  return render(
    <TrafficGenConfigForm
      config={{ ...defaultTrafficGenConfig, ...overrides }}
      setConfig={() => {}}
      selectedTests={['trafficgen_burst', 'custom_stream']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TrafficGenConfigForm — i18n', () => {
  it('calls its recap what the other six call theirs', () => {
    renderForm();
    expect(screen.getByText('Test Summary')).toBeInTheDocument();
    expect(screen.queryByText('Traffic Summary')).not.toBeInTheDocument();
  });

  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('Traffic Parameters')).toBeInTheDocument();
    expect(screen.getByText('MAC Addresses (Optional)')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Parámetros de tráfico')).toBeInTheDocument();
    expect(screen.getByText('Direcciones MAC (opcional)')).toBeInTheDocument();
    expect(screen.queryByText('Traffic Parameters')).not.toBeInTheDocument();
  });

  it('keeps frame sizes as numbers with a translated qualifier', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByRole('option', { name: '64 B (mín.)' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '9000 B (jumbo)' })).toBeInTheDocument();
  });
});

/* The rate presets write through setValue, like the frame-size checkboxes on
   the RFC 2544 and Y.1564 forms (#1465). */
describe('TrafficGenConfigForm — store sync', () => {
  it('sends a rate preset to the run config on its own', async () => {
    const setConfig = vi.fn();
    render(
      <TrafficGenConfigForm
        config={defaultTrafficGenConfig}
        setConfig={setConfig}
        selectedTests={['trafficgen_burst', 'custom_stream']}
      />,
    );

    await userEvent.click(screen.getByRole('button', { name: '25%' }));

    await waitFor(() =>
      expect(setConfig).toHaveBeenLastCalledWith(expect.objectContaining({ ratePct: 25 })),
    );
  });
});
