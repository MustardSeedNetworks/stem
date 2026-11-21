import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { DaemonUnreachableBanner } from './DaemonUnreachableBanner';

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('DaemonUnreachableBanner', () => {
  it('names the daemon and says it is retrying when unreachable', async () => {
    await i18n.changeLanguage('en');
    render(<DaemonUnreachableBanner status="disconnected" />);
    expect(screen.getByRole('alert')).toHaveTextContent("Can't reach Stem daemon, retrying");
  });

  it('speaks Spanish under es', async () => {
    await i18n.changeLanguage('es');
    render(<DaemonUnreachableBanner status="disconnected" />);
    expect(screen.getByRole('alert')).toHaveTextContent(
      'No se puede conectar con el daemon de Stem, reintentando',
    );
  });

  it.each(['connected', 'checking'] as const)('renders nothing while %s', (status) => {
    render(<DaemonUnreachableBanner status={status} />);
    expect(screen.queryByTestId('daemon-unreachable-banner')).not.toBeInTheDocument();
  });
});
