/**
 * Second of the seven forms. The suite initialises the real i18next against
 * the real locale files, so switching the language and asserting Spanish
 * proves this form reads from them — a label that regressed to a hardcoded
 * string would still render English and fail.
 *
 * Y.1564's vocabulary is where the do-not-translate rule earns its keep: CIR,
 * EIR, FLR and the standard's own name are terms an operator matches against
 * a service order, not words to render into Spanish.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultY1564Config, Y1564ConfigForm } from './Y1564ConfigForm';

function renderForm(overrides: Partial<typeof defaultY1564Config> = {}) {
  return render(
    <Y1564ConfigForm
      config={{ ...defaultY1564Config, ...overrides }}
      setConfig={() => {}}
      selectedTests={['y1564_full']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('Y1564ConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('Bandwidth Parameters')).toBeInTheDocument();
    expect(screen.getByText('SLA Thresholds')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Parámetros de ancho de banda')).toBeInTheDocument();
    expect(screen.getByText('Umbrales de SLA')).toBeInTheDocument();
    expect(screen.queryByText('Bandwidth Parameters')).not.toBeInTheDocument();
  });

  it('keeps the service-order vocabulary verbatim in Spanish', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('CIR (Mbps)')).toBeInTheDocument();
    expect(screen.getByText('EIR (Mbps)')).toBeInTheDocument();
    expect(screen.getByText('Configuración Y.1564 / MEF')).toBeInTheDocument();
  });

  it('translates the priority list without renumbering it', async () => {
    await i18n.changeLanguage('es');
    /* The priority list only exists on a tagged service. */
    renderForm({ vlanId: 100 });

    /* PCP values are 802.1p wire values; the names beside them are labels. */
    expect(screen.getByRole('option', { name: '0 - Best Effort (BE)' })).toBeInTheDocument();
  });
});
