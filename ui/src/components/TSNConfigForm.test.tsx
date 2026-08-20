/**
 * Last of the seven, and the one that already had its summary extracted —
 * TestSummarySection is now the shared TestSummary, so the last inline copy
 * of that block is gone.
 *
 * TSN's terms are standard numbers (802.1Qbv, 802.1Qbu, IEEE 1588) and the
 * metrics the DNT list already protects, so Spanish translates around them.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultTSNConfig, TSNConfigForm } from './TSNConfigForm';

function renderForm() {
  return render(
    <TSNConfigForm
      config={defaultTSNConfig}
      setConfig={() => {}}
      selectedTests={['tsn_latency', 'tsn_scheduling']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TSNConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('PTP Synchronization')).toBeInTheDocument();
    expect(screen.getByText('Traffic Scheduling (802.1Qbv)')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Sincronización PTP')).toBeInTheDocument();
    expect(screen.queryByText('PTP Synchronization')).not.toBeInTheDocument();
  });

  it('keeps the standard numbers verbatim', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Planificación del tráfico (802.1Qbv)')).toBeInTheDocument();
    expect(screen.getByText(/IEEE 1588/)).toBeInTheDocument();
  });

  it('uses the shared summary, so no form is left with its own copy', () => {
    renderForm();
    expect(screen.getByText('Test Summary')).toBeInTheDocument();
  });
});
