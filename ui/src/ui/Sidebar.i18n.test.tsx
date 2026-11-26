/**
 * Sidebar.i18n.test.tsx — the rail and its footer render real locale copy.
 *
 * #759. The footer actions once carried hardcoded English while a translated
 * string for each sat unused in the locale file, and nothing failed because
 * nothing asserted on the text.
 */
import { render, screen } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter } from 'react-router';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { SidebarLayout } from './Sidebar';

// The rail calls createElement(icon), so the fixture needs a real component.
const StubIcon = (): ReactElement => <svg aria-hidden="true" />;

const groups = [
  { label: 'Test', items: [{ label: 'Dashboard', path: '/', icon: StubIcon }] },
] as unknown as Parameters<typeof SidebarLayout>[0]['groups'];

const renderSidebar = (): void => {
  render(
    <MemoryRouter>
      <SidebarLayout
        groups={groups}
        version="1.2.3"
        onOpenHelp={() => undefined}
        onOpenSettings={() => undefined}
        onToggleTheme={() => undefined}
        isDark={false}
      >
        <div id="main-content" />
      </SidebarLayout>
    </MemoryRouter>,
  );
};

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('Sidebar — real locale copy', () => {
  it('labels the footer actions and the skip link in English', async () => {
    await i18n.changeLanguage('en');
    renderSidebar();

    // Both the desktop and mobile rails stay mounted — responsive classes
    // toggle display, not mount — so each action legitimately appears twice.
    expect(screen.getByText('Skip to main content')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Help' }).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: 'Settings' }).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: 'Switch to dark mode' }).length).toBeGreaterThan(
      0,
    );
  });

  it('translates them under es', async () => {
    await i18n.changeLanguage('es');
    renderSidebar();

    expect(screen.getByText('Saltar al contenido principal')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Ayuda' }).length).toBeGreaterThan(0);
    expect(screen.queryByText('Skip to main content')).not.toBeInTheDocument();
  });
});
