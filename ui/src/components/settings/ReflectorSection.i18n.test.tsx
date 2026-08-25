/**
 * ReflectorSection.i18n.test.tsx — the reflector profile picker renders real
 * locale copy.
 *
 * #759 asked for copy assertions on the reflector surface. Writing them found
 * that none of this component's eight labels had ever resolved: the keys read
 * `settings.reflector.netally` with a dot, which i18next parses as a path
 * inside the default `common` namespace rather than as the `settings`
 * namespace, so every one fell through to a hardcoded English default in both
 * locales. The keys did not exist in any catalog either.
 *
 * Nothing caught it because the keys were carried on an object and passed as
 * `t(p.nameKey, p.nameDefault)`: dynamic, so the key checker could not see
 * them, and a banned `t(key, fallback)` in a shape semgrep's rules miss
 * because the fallback is an object property rather than a literal argument.
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { ReflectorSection } from './ReflectorSection';

function renderSection() {
  return render(<ReflectorSection profile="msn" onProfileChange={() => {}} />);
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('ReflectorSection — real locale copy', () => {
  it('renders the English profile descriptions', async () => {
    await i18n.changeLanguage('en');
    renderSection();

    expect(screen.getByText('ITO signatures only')).toBeInTheDocument();
    expect(screen.getByText('All signature types')).toBeInTheDocument();
    expect(screen.getByText('Manual configuration')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    renderSection();

    expect(screen.getByText('Solo firmas ITO')).toBeInTheDocument();
    expect(screen.getByText('Todos los tipos de firma')).toBeInTheDocument();
    expect(screen.getByText('Configuración manual')).toBeInTheDocument();
    expect(screen.queryByText('ITO signatures only')).toBeNull();
    expect(screen.queryByText('Manual configuration')).toBeNull();
  });

  it('keeps the vendor names verbatim in both locales', async () => {
    await i18n.changeLanguage('es');
    renderSection();

    // NetAlly and MSN are proper nouns; the descriptions around them translate.
    expect(screen.getByText('NetAlly')).toBeInTheDocument();
    expect(screen.getByText('MSN')).toBeInTheDocument();
  });
});
