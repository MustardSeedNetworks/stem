import { cleanup, fireEvent, render, renderHook, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { getGlossary } from '../data/help/glossary';
import { tests } from '../data/help/tests';
import i18n from '../i18n';
import { usePages } from '../pageRegistry';
import { HelpDrawer } from './HelpDrawer';

vi.mock('../hooks/useBuildVersion', () => ({
  useBuildVersion: () => ({
    version: 'v0.24.120',
    commit: 'test',
    buildTime: '2026-09-15',
    uiBuildHash: 'test',
  }),
}));

afterEach(async () => {
  cleanup();
  await i18n.changeLanguage('en');
});

describe('help content coverage', () => {
  it.each(['en', 'es'])('explains every routed page name in %s', async (language) => {
    await i18n.changeLanguage(language);
    const terms = getGlossary(i18n.getFixedT(language, 'help')).map((entry) =>
      entry.term.toLowerCase(),
    );
    const pages = renderHook(usePages).result.current;
    for (const page of pages) expect(terms).toContain(page.label.toLowerCase());
  });

  it('renders Spanish usage and lets related tests open their actual details', async () => {
    await i18n.changeLanguage('es');
    render(<HelpDrawer isOpen onClose={() => {}} initialTestId="throughput" />);
    expect(screen.getByText(/Validar equipos antes de instalarlos/)).toBeVisible();
    expect(screen.getByTestId('help-search')).toHaveFocus();
    expect(screen.getByTestId('help-standard-link')).toHaveAttribute(
      'href',
      'https://www.rfc-editor.org/rfc/rfc2544.html',
    );
    fireEvent.click(screen.getByRole('button', { name: 'Latency Test' }));
    expect(screen.getByRole('heading', { name: 'Latency Test' })).toBeVisible();
  });

  it('searches Spanish glossary definitions instead of the old English corpus', async () => {
    await i18n.changeLanguage('es');
    render(<HelpDrawer isOpen onClose={() => {}} initialTab="glossary" />);
    fireEvent.change(screen.getByTestId('help-search'), { target: { value: 'comprometida' } });
    expect(screen.getByText('Tasa de información comprometida')).toBeVisible();
    expect(screen.queryByText('Committed Information Rate')).not.toBeInTheDocument();
  });

  it('has an HTTPS source and valid related test destinations for every test', () => {
    expect(Object.values(tests)).toHaveLength(28);
    for (const test of Object.values(tests)) {
      const sources = test.seeAlso.filter((reference) => reference.startsWith('https://'));
      expect(sources.length, test.id).toBeGreaterThan(0);
      for (const reference of test.seeAlso) {
        if (reference.startsWith('https://')) expect(new URL(reference).protocol).toBe('https:');
        else expect(tests[reference], `${test.id}: ${reference}`).toBeDefined();
      }
    }
  });
});
