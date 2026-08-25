/**
 * SetupWizard.i18n.test.tsx — first-run setup renders real locale copy.
 *
 * #759: wrecking every EN locale file failed only 19 of 206 tests. Setup is
 * the first screen an operator ever sees and had no copy assertions at all,
 * so its 30 keys could go missing or ship untranslated without CI noticing.
 */
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { RoleProvider } from '../../contexts/RoleContext';
import i18n from '../../i18n';
import { SetupWizard } from './SetupWizard';

function renderWizard() {
  return render(
    <RoleProvider>
      <SetupWizard
        onComplete={() => {}}
        onLogin={async () => true}
        suggestedPassword="Correct-Horse-Battery-9"
        setupToken="token"
      />
    </RoleProvider>,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('SetupWizard — real locale copy', () => {
  it('renders the English welcome step', async () => {
    await i18n.changeLanguage('en');
    renderWizard();

    await waitFor(() => {
      expect(screen.getByText('Welcome to The Stem')).toBeInTheDocument();
    });
    expect(screen.getByText('Set up your admin password to get started')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    renderWizard();

    await waitFor(() => {
      expect(screen.getByText('Bienvenido a The Stem')).toBeInTheDocument();
    });
    expect(
      screen.getByText('Configure su contraseña de administrador para comenzar'),
    ).toBeInTheDocument();
    expect(screen.queryByText('Set up your admin password to get started')).toBeNull();
  });

  it('keeps the product name verbatim in both locales, per the glossary', async () => {
    await i18n.changeLanguage('es');
    renderWizard();

    // "The Stem" is a glossary term: the copy around it translates, the name
    // itself must survive verbatim. It appears in more than one node, so this
    // asserts presence rather than uniqueness.
    await waitFor(() => {
      expect(screen.getAllByText(/The Stem/).length).toBeGreaterThan(0);
    });
  });
});
