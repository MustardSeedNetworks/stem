/**
 * StopOutcomeMessage.test.tsx — the outcome of a stop request, as rendered.
 *
 * `handleStopTest` ignored `response.ok` and logged failures as "non-critical"
 * (#1080): the daemon refusing a stop and the daemon performing one produced
 * the same screen. These assert on the visible strings from the real locale
 * catalogs, in both locales, so a missing `es` key fails here rather than
 * shipping as an English word on a Spanish screen.
 *
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { StopOutcomeMessage } from './StopOutcomeMessage';

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('StopOutcomeMessage', () => {
  it('renders nothing while idle', () => {
    const { container } = render(<StopOutcomeMessage outcome={{ kind: 'idle' }} />);

    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing while the request is in flight — the button owns that', () => {
    const { container } = render(<StopOutcomeMessage outcome={{ kind: 'stopping' }} />);

    expect(container).toBeEmptyDOMElement();
  });

  it('confirms a completed stop in English', async () => {
    await i18n.changeLanguage('en');
    render(<StopOutcomeMessage outcome={{ kind: 'stopped' }} />);

    const message = screen.getByTestId('test-stop-message');
    expect(message).toHaveTextContent('Stopped');
    // A completed stop is not a failure: it must not be announced assertively
    // or coloured as an error.
    expect(message).toHaveAttribute('role', 'status');
  });

  it('confirms a completed stop in Spanish', async () => {
    await i18n.changeLanguage('es');
    render(<StopOutcomeMessage outcome={{ kind: 'stopped' }} />);

    expect(screen.getByTestId('test-stop-message')).toHaveTextContent('Detenido');
  });

  // The daemon's sentence, not ours: "No test is currently running" is the
  // whole point of the row, and paraphrasing it in the client would put the
  // UI's guess in front of the server's answer.
  it('renders the daemon refusal verbatim, as an alert', () => {
    render(
      <StopOutcomeMessage
        outcome={{ kind: 'stopRejected', message: 'No test is currently running' }}
      />,
    );

    const message = screen.getByTestId('test-stop-message');
    expect(message).toHaveTextContent('No test is currently running');
    expect(message).toHaveAttribute('role', 'alert');
  });

  it('renders a transport failure as an alert too', () => {
    render(<StopOutcomeMessage outcome={{ kind: 'stopFailed', message: 'Failed to stop test' }} />);

    const message = screen.getByTestId('test-stop-message');
    expect(message).toHaveTextContent('Failed to stop test');
    expect(message).toHaveAttribute('role', 'alert');
  });
});
