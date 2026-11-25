import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import { LanguageSection } from './LanguageSection';

beforeEach(async () => {
  window.localStorage.removeItem('stem-language');
  await i18n.changeLanguage('en');
});

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('LanguageSection', () => {
  it('offers both locales with the active one selected', () => {
    render(<LanguageSection />);
    const select = screen.getByRole('combobox', { name: 'Language' });
    expect(select).toHaveValue('en');
    expect(screen.getByRole('option', { name: 'Español' })).toBeInTheDocument();
  });

  it('switches the interface to Spanish, persists it and sets the document language', async () => {
    const user = userEvent.setup();
    render(<LanguageSection />);

    await user.selectOptions(screen.getByRole('combobox', { name: 'Language' }), 'es');

    expect(i18n.resolvedLanguage).toBe('es');
    expect(window.localStorage.getItem('stem-language')).toBe('es');
    expect(document.documentElement.lang).toBe('es');
    expect(screen.getByRole('combobox', { name: 'Idioma' })).toHaveValue('es');
  });
});
