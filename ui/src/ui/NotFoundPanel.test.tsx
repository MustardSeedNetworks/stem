import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { NotFoundPanel } from './NotFoundPanel';

afterEach(async () => {
  await i18n.changeLanguage('en');
});

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/" element={<p>home page</p>} />
        <Route path="*" element={<NotFoundPanel />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('NotFoundPanel', () => {
  it('names Stem, stays on the bad address and links home', async () => {
    const user = userEvent.setup();
    renderAt('/nope/not-a-page');

    expect(screen.getByRole('heading', { name: 'Page not found' })).toBeInTheDocument();
    expect(screen.getByText('Stem has no page at this address.')).toBeInTheDocument();
    expect(document.title).toBe('Page not found · Stem');

    await user.click(screen.getByRole('link', { name: 'Back to Stem home' }));
    expect(screen.getByText('home page')).toBeInTheDocument();
  });

  it('renders Spanish under es', async () => {
    await i18n.changeLanguage('es');
    renderAt('/nope');
    expect(screen.getByRole('heading', { name: 'Página no encontrada' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Volver al inicio de Stem' })).toHaveAttribute(
      'href',
      '/',
    );
  });
});
