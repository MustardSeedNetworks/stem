/**
 * TestProgressBar.i18n.test.tsx — the run-plan progress surface renders real
 * locale copy, in both languages.
 *
 * #1252. This component called no `t()` at all: Starting…/Running/Completed,
 * `Elapsed:`, `ETA:`, `Step n of m` and the per-step status were English
 * literals, and its own unit test asserted those literals as correct. The
 * step status is the interesting one — it arrives from the server as a bare
 * string, so the temptation is `t(`status.${step.status}`)`, which the
 * extractor cannot see and which ships a missing key.
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { TestProgress } from './TestProgressBar';
import { TestProgressBar } from './TestProgressBar';

function progress(overrides: Partial<TestProgress> = {}): TestProgress {
  return {
    status: 'running',
    currentTest: 'Y.1731 delay',
    currentStep: 2,
    stepsTotal: 3,
    phase: 'Executing y1731_delay',
    elapsedSeconds: 15,
    estimatedRemainingSeconds: 90,
    steps: [
      { testType: 'rfc2544_throughput', module: 'benchmark', status: 'passed' },
      { testType: 'y1731_delay', module: 'measure', status: 'running' },
      { testType: 'y1564_service', module: 'servicetest', status: 'skipped' },
    ],
    ...overrides,
  };
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TestProgressBar — real locale copy', () => {
  it('renders the English strings from the locale file', async () => {
    await i18n.changeLanguage('en');
    render(<TestProgressBar progress={progress()} />);

    expect(screen.getByText('Elapsed: 0:15')).toBeInTheDocument();
    expect(screen.getByText('ETA: ~2m')).toBeInTheDocument();
    expect(screen.getByText(/Step 2 of 3/)).toBeInTheDocument();
    expect(screen.getByText('(Running)')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAccessibleName('Run plan progress');
  });

  it('translates every one of them under es', async () => {
    await i18n.changeLanguage('es');
    render(<TestProgressBar progress={progress()} />);

    expect(screen.getByText('Transcurrido: 0:15')).toBeInTheDocument();
    expect(screen.getByText('ETA: ~2 min')).toBeInTheDocument();
    expect(screen.getByText(/Paso 2 de 3/)).toBeInTheDocument();
    expect(screen.getByText('(Ejecutando)')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAccessibleName('Progreso del plan de ejecución');
    expect(screen.queryByText(/Elapsed:/)).toBeNull();
    expect(screen.queryByText(/Step 2 of 3/)).toBeNull();
  });

  // The server's step status is a bare string; a template-literal key would
  // leave these untranslated and unextracted.
  it('translates the per-step status, not the raw server string', async () => {
    await i18n.changeLanguage('es');
    render(<TestProgressBar progress={progress()} />);

    const steps = screen.getByRole('list', { name: 'Pasos del plan de ejecución' });
    expect(steps).toHaveTextContent('Aprobado');
    expect(steps).toHaveTextContent('Ejecutando');
    expect(steps).toHaveTextContent('Omitido');
    expect(steps).not.toHaveTextContent('passed');
    expect(steps).not.toHaveTextContent('skipped');
  });

  it('translates the terminal ETA and the idle status label', async () => {
    await i18n.changeLanguage('es');
    render(
      <TestProgressBar progress={progress({ estimatedRemainingSeconds: 0, status: 'starting' })} />,
    );

    expect(screen.getByText('ETA: Completado')).toBeInTheDocument();
  });
});
