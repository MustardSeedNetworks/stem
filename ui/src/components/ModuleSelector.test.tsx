/**
 * ModuleSelector Component Tests
 *
 * The catalogue used to be fetched from '/api/modules' — no /v1, unlike every
 * other call in the app. That path does not 404: it falls through to the SPA
 * handler and returns index.html with HTTP 200, so `response.ok` was true,
 * `.json()` threw on HTML, and the catch silently substituted a hardcoded
 * list. The drawer looked populated and correct while never once reflecting
 * the daemon.
 *
 * These tests pin the three things that made that possible: the path, the
 * refusal to accept a non-JSON body, and the absence of a fallback that reads
 * as success.
 */

import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ModuleSelector } from './ModuleSelector';

const MODULES = {
  count: 1,
  modules: [
    {
      name: 'benchmark',
      displayName: 'Benchmark',
      description: 'RFC 2544 device benchmarking',
      color: '#dc2626',
      standard: 'RFC 2544',
      tests: ['rfc2544_throughput'],
    },
  ],
};

function mockFetch(response: Partial<Response> & { json?: () => Promise<unknown> }): void {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
      json: () => Promise.resolve(MODULES),
      ...response,
    }),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ModuleSelector', () => {
  it('requests the versioned API path', async () => {
    mockFetch({});
    render(<ModuleSelector selectedTests={[]} setSelectedTests={vi.fn()} />);

    await waitFor(() => {
      expect(fetch).toHaveBeenCalledWith('/api/v1/modules');
    });
    // The unversioned path is the bug: it resolves to the SPA, not the API.
    expect(fetch).not.toHaveBeenCalledWith('/api/modules');
  });

  it('renders modules returned by the daemon', async () => {
    mockFetch({});
    render(<ModuleSelector selectedTests={[]} setSelectedTests={vi.fn()} />);

    expect(await screen.findByText('Benchmark')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('surfaces an error when the response is HTML rather than JSON', async () => {
    // Exactly the old failure: HTTP 200 from the SPA handler.
    mockFetch({
      headers: new Headers({ 'content-type': 'text/html' }),
      json: () => Promise.reject(new Error('unexpected token <')),
    });
    render(<ModuleSelector selectedTests={[]} setSelectedTests={vi.fn()} />);

    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent(/could not load the test modules/i);
    // No hardcoded catalogue may stand in for the real one.
    expect(screen.queryByText('Benchmark')).not.toBeInTheDocument();
  });

  it('surfaces an error on a non-OK response', async () => {
    mockFetch({ ok: false, status: 500 });
    render(<ModuleSelector selectedTests={[]} setSelectedTests={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent(/HTTP 500/);
    expect(screen.queryByText('Benchmark')).not.toBeInTheDocument();
  });
});
