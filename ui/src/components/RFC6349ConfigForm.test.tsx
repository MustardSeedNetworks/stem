/**
 * Sixth of the seven. RFC 6349 is TCP throughput, so its vocabulary is
 * transport terms an operator reads in a capture — RTT, MSS, RWND, BDP — and
 * translating them would make the form harder to use, not easier.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultRFC6349Config, RFC6349ConfigForm } from './RFC6349ConfigForm';

function renderForm() {
  return render(
    <RFC6349ConfigForm
      config={defaultRFC6349Config}
      setConfig={() => {}}
      selectedTests={['rfc6349_throughput']}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('RFC6349ConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('Network Parameters')).toBeInTheDocument();
    expect(screen.getByText('TCP Parameters')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Parámetros de red')).toBeInTheDocument();
    expect(screen.queryByText('Network Parameters')).not.toBeInTheDocument();
  });

  it('keeps the transport vocabulary verbatim', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('MSS')).toBeInTheDocument();
    /* RWND and RTT appear in more than one place — a label and the summary —
       which is itself the point: the same term everywhere it is read. */
    expect(screen.getAllByText(/RWND/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/RTT/).length).toBeGreaterThan(0);
  });

  it('translates a mode name without renumbering its wire value', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    const downstream = screen.getByRole<HTMLOptionElement>('option', { name: 'Bajada' });
    expect(downstream.value).toBe('2');
  });
});
