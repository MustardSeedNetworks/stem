/**
 * Fourth of the seven. Y.1731's field names are the protocol's own
 * identifiers — MEP, MEG, CCM — and an operator matches them against a
 * network design document, so Spanish translates the prose around them and
 * leaves them alone.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultY1731Config, Y1731ConfigForm } from './Y1731ConfigForm';

function renderForm() {
  return render(
    <Y1731ConfigForm
      config={defaultY1731Config}
      setConfig={() => {}}
      selectedTests={['y1731_delay', 'y1731_loss']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('Y1731ConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('MEP/MEG Configuration')).toBeInTheDocument();
    expect(screen.getByText('Measurement Parameters')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Parámetros de medición')).toBeInTheDocument();
    expect(screen.queryByText('Measurement Parameters')).not.toBeInTheDocument();
  });

  it('leaves the protocol identifiers alone', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('MEP ID')).toBeInTheDocument();
    expect(screen.getByText('MEG ID')).toBeInTheDocument();
    expect(screen.getByText('Configuración OAM Y.1731')).toBeInTheDocument();
  });

  it('offers only the OAM frame sizes, not the sweep list', () => {
    renderForm();

    expect(screen.getByRole('option', { name: '64 B (min)' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: '1518 B (max)' })).toBeInTheDocument();
    /* Jumbo frames are not an OAM measurement size, and adopting the shared
       sweep list wholesale would have quietly offered them. */
    expect(screen.queryByRole('option', { name: /9000/ })).not.toBeInTheDocument();
  });
});
