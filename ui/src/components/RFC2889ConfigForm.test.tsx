/**
 * Fifth of the seven. RFC 2889's traffic patterns are wire values with names
 * beside them, so the test pins that translating the name cannot renumber the
 * value the daemon receives.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultRFC2889Config, RFC2889ConfigForm } from './RFC2889ConfigForm';

function renderForm() {
  return render(
    <RFC2889ConfigForm
      config={defaultRFC2889Config}
      setConfig={() => {}}
      selectedTests={['rfc2889_forwarding', 'rfc2889_learning']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('RFC2889ConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('Switch Configuration')).toBeInTheDocument();
    expect(screen.getByText('Test Parameters')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Configuración del switch')).toBeInTheDocument();
    expect(screen.queryByText('Switch Configuration')).not.toBeInTheDocument();
  });

  it('translates a pattern name without renumbering its wire value', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    const fullMesh = screen.getByRole<HTMLOptionElement>('option', { name: 'Malla completa' });
    expect(fullMesh.value).toBe('0');
    const broadcast = screen.getByRole<HTMLOptionElement>('option', { name: 'Difusión' });
    expect(broadcast.value).toBe('2');
  });

  it('keeps the standard name verbatim in the help text', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getAllByTitle(/RFC 2889/).length).toBeGreaterThan(0);
  });
});
