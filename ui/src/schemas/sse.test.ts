import { describe, expect, it } from 'vitest';
import {
  parseSseEnvelope,
  parseSseModeChanged,
  parseSseReflectorStats,
  parseSseTestProgress,
} from './sse';

describe('parseSseEnvelope', () => {
  it('accepts a minimal envelope', () => {
    expect(parseSseEnvelope({ type: 'heartbeat', payload: null })).not.toBeNull();
  });

  it('rejects missing type', () => {
    expect(parseSseEnvelope({ payload: {} })).toBeNull();
  });

  it('rejects non-string type', () => {
    expect(parseSseEnvelope({ type: 42, payload: null })).toBeNull();
  });

  it('rejects non-objects', () => {
    expect(parseSseEnvelope(null)).toBeNull();
    expect(parseSseEnvelope('whoops')).toBeNull();
  });
});

describe('parseSseModeChanged', () => {
  it('parses a valid mode_changed frame', () => {
    const env = parseSseEnvelope({
      type: 'mode_changed',
      payload: { status: 'updated', mode: 'reflector', previous: 'test_master' },
    });
    expect(env).not.toBeNull();
    if (!env) return;
    const parsed = parseSseModeChanged(env);
    expect(parsed).not.toBeNull();
    expect(parsed?.payload.mode).toBe('reflector');
  });

  it('returns null when envelope is a different type', () => {
    const env = parseSseEnvelope({ type: 'heartbeat', payload: null });
    expect(env).not.toBeNull();
    if (!env) return;
    expect(parseSseModeChanged(env)).toBeNull();
  });

  it('rejects an invalid mode value', () => {
    const env = parseSseEnvelope({
      type: 'mode_changed',
      payload: { status: 'updated', mode: 'passthrough', previous: 'reflector' },
    });
    expect(env).not.toBeNull();
    if (!env) return;
    expect(parseSseModeChanged(env)).toBeNull();
  });
});

describe('parseSseReflectorStats', () => {
  it('accepts a minimal reflector_stats frame', () => {
    const env = parseSseEnvelope({
      type: 'reflector_stats',
      payload: { running: true, ratePps: 1234.5 },
    });
    expect(env).not.toBeNull();
    if (!env) return;
    const parsed = parseSseReflectorStats(env);
    expect(parsed?.payload.running).toBe(true);
    expect(parsed?.payload.ratePps).toBe(1234.5);
  });

  it('accepts an empty payload (forward-compat)', () => {
    const env = parseSseEnvelope({ type: 'reflector_stats', payload: {} });
    expect(env).not.toBeNull();
    if (!env) return;
    expect(parseSseReflectorStats(env)).not.toBeNull();
  });

  it('rejects non-numeric stats fields', () => {
    const env = parseSseEnvelope({
      type: 'reflector_stats',
      payload: { ratePps: 'fast' },
    });
    expect(env).not.toBeNull();
    if (!env) return;
    expect(parseSseReflectorStats(env)).toBeNull();
  });
});

describe('parseSseTestProgress', () => {
  it('accepts a test_progress frame with arbitrary progress payload', () => {
    const env = parseSseEnvelope({
      type: 'test_progress',
      payload: { testId: 'rfc2544-1', progress: { phase: 'throughput', percent: 42 } },
    });
    expect(env).not.toBeNull();
    if (!env) return;
    const parsed = parseSseTestProgress(env);
    expect(parsed?.payload.testId).toBe('rfc2544-1');
  });

  it('rejects missing testId', () => {
    const env = parseSseEnvelope({ type: 'test_progress', payload: { progress: {} } });
    expect(env).not.toBeNull();
    if (!env) return;
    expect(parseSseTestProgress(env)).toBeNull();
  });
});
