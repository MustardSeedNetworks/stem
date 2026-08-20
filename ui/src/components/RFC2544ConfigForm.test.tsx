/**
 * The seven test-config forms are the product's primary configuration
 * surface and had no i18n at all: a Spanish operator got a translated shell
 * around an entirely English form, and CI reported nothing, because the i18n
 * job validates the locale files that exist rather than the surfaces that
 * never call t().
 *
 * These pin the wiring for the first form migrated. The test setup
 * initialises the real i18next against the real locale files, so asserting
 * on Spanish text asserts that this form reads from them — a form that
 * regressed to a hardcoded string would still render English here.
 */
import { render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { afterEach, describe, expect, it } from 'vitest';
import { defaultRFC2544Config, RFC2544ConfigForm } from './RFC2544ConfigForm';

const ALL_TESTS = ['rfc2544_throughput', 'rfc2544_frame_loss'];

function renderForm() {
  return render(
    <RFC2544ConfigForm
      config={defaultRFC2544Config}
      setConfig={() => {}}
      selectedTests={ALL_TESTS}
    />,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('RFC2544ConfigForm — i18n', () => {
  it('renders its labels from the locale files', () => {
    renderForm();
    expect(screen.getByText('Duration per Test (s)')).toBeInTheDocument();
    expect(screen.getByText('Frame Sizes')).toBeInTheDocument();
  });

  it('follows the active language instead of staying English', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText('Duración por prueba (s)')).toBeInTheDocument();
    expect(screen.getByText('Tamaños de trama')).toBeInTheDocument();
    expect(screen.queryByText('Duration per Test (s)')).not.toBeInTheDocument();
  });

  it('keeps standard and metric terms verbatim in Spanish', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    /* The locale gate's do-not-translate list covers these; asserting it at
       the surface as well means a form that hand-builds a translated string
       from parts cannot slip past the JSON-level check. The standard's name
       reaches the user through a frame-size checkbox's title. */
    expect(screen.getAllByTitle(/RFC 2544/).length).toBeGreaterThan(0);
    expect(screen.getByText('Prueba de throughput')).toBeInTheDocument();
  });

  it('says which tests are selected in the summary, in the active language', async () => {
    await i18n.changeLanguage('es');
    renderForm();

    expect(screen.getByText(/Pruebas seleccionadas/)).toBeInTheDocument();
    expect(screen.getByText(/Tiempo estimado/)).toBeInTheDocument();
  });
});
