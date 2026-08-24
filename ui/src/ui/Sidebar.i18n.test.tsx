/**
 * Sidebar.i18n.test.tsx — the rail and its footer render real locale copy.
 *
 * #759. The sidebar footer is the surface where #750 found `tooltips.header.*`
 * keys that had been authored and never wired: the buttons carried hardcoded
 * English while a translated string for each sat unused in the locale file.
 * Nothing failed, because nothing asserted on the text.
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
        onOpenProfiles={() => undefined}
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
  it('labels the footer buttons and the skip link in English', async () => {
    await i18n.changeLanguage('en');
    renderSidebar();

    // Both the desktop and mobile rails stay mounted — responsive classes
    // toggle display, not mount — so each label legitimately appears twice.
    expect(screen.getByText('Skip to main content')).toBeInTheDocument();
    expect(screen.getAllByText('Help').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Settings').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Profiles').length).toBeGreaterThan(0);
  });

  it('translates them under es', async () => {
    await i18n.changeLanguage('es');
    renderSidebar();

    expect(screen.getByText('Saltar al contenido principal')).toBeInTheDocument();
    expect(screen.getAllByText('Perfiles').length).toBeGreaterThan(0);
    expect(screen.queryByText('Skip to main content')).not.toBeInTheDocument();
  });

  it('gives the help button its descriptive tooltip, not just its label', async () => {
    await i18n.changeLanguage('en');
    renderSidebar();

    // The distinction matters: `labels.help` is the visible word, while
    // `tooltips.chrome.help` is the sentence #750 wired in from the unused keys.
    expect(
      screen.getAllByTitle('Open the help drawer with test references, tutorials, and a glossary')
        .length,
    ).toBeGreaterThan(0);
  });
});
