/**
 * TestResults.i18n.test.tsx — the result header renders real locale copy.
 *
 * The module line was `Module: {result.module}`, JSX text that ends at an
 * expression, which the shared gate could not see until .github#100. These
 * import the real i18n instance so a missing key or an English fragment fails.
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { TestResults } from './TestResults';

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TestResults — real locale copy', () => {
  it('names the module in Spanish under es', async () => {
    await i18n.changeLanguage('es');
    render(
      <TestResults
        testStatus="completed"
        result={{
          testType: 'rfc2544_throughput',
          module: 'benchmark',
          status: 'completed',
          success: true,
        }}
      />,
    );

    expect(screen.getByText('Módulo: benchmark')).toBeInTheDocument();
    expect(screen.queryByText(/Module:/)).toBeNull();
  });
});
