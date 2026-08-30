/**
 * Logger tests.
 *
 * logError/logWarn are the app's only structured error path, and they are
 * no-ops outside development. Both halves are worth pinning: that they DO
 * write in development (a logger that silently swallows everything is
 * indistinguishable from no logger at all) and that they write NOTHING in a
 * production build, which is what keeps stack traces and context objects out
 * of an operator's console.
 *
 * isDev is read from import.meta.env at module evaluation, so each case stubs
 * the environment and re-imports rather than trying to toggle it at runtime.
 */

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

async function loggerWith(dev: boolean) {
  vi.stubEnv('DEV', dev);
  vi.resetModules();

  return import('./logger');
}

let errorSpy: ReturnType<typeof vi.spyOn>;
let warnSpy: ReturnType<typeof vi.spyOn>;

beforeEach(() => {
  errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
  warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

describe('in development', () => {
  it('logError writes the message, stack and context', async () => {
    const { logError } = await loggerWith(true);
    const error = new Error('interface enumeration failed');

    logError(error, { component: 'InterfaceSection', action: 'refresh' });

    expect(errorSpy).toHaveBeenCalledTimes(1);
    const [prefix, payload] = errorSpy.mock.calls[0] as [string, Record<string, unknown>];
    expect(prefix).toBe('[STEM Error]');
    expect(payload).toMatchObject({
      message: 'interface enumeration failed',
      component: 'InterfaceSection',
      action: 'refresh',
    });
    expect(payload.stack).toContain('Error: interface enumeration failed');
    expect(typeof payload.timestamp).toBe('string');
  });

  it('logError stringifies a thrown non-Error and reports no stack', async () => {
    // A rejected fetch or a thrown string would otherwise log "undefined".
    const { logError } = await loggerWith(true);

    logError('plain string failure');

    const [, payload] = errorSpy.mock.calls[0] as [string, Record<string, unknown>];
    expect(payload.message).toBe('plain string failure');
    expect(payload.stack).toBeUndefined();
  });

  it('logWarn writes the message and context', async () => {
    const { logWarn } = await loggerWith(true);

    logWarn('reflector already running', { component: 'ReflectorPage' });

    expect(warnSpy).toHaveBeenCalledTimes(1);
    const [prefix, payload] = warnSpy.mock.calls[0] as [string, Record<string, unknown>];
    expect(prefix).toBe('[STEM Warning]');
    expect(payload).toMatchObject({
      message: 'reflector already running',
      component: 'ReflectorPage',
    });
  });

  it('logError works without a context argument', async () => {
    const { logError } = await loggerWith(true);

    logError(new Error('boom'));

    expect(errorSpy).toHaveBeenCalledTimes(1);
  });
});

describe('in a production build', () => {
  it('logError writes nothing', async () => {
    const { logError } = await loggerWith(false);

    logError(new Error('secret detail'), { additionalData: { token: 'sensitive' } });

    expect(errorSpy).not.toHaveBeenCalled();
  });

  it('logWarn writes nothing', async () => {
    const { logWarn } = await loggerWith(false);

    logWarn('noisy warning');

    expect(warnSpy).not.toHaveBeenCalled();
  });
});
