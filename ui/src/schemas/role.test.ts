import { describe, expect, it } from 'vitest';
import { isStemRole, parseApiErrorBody, parseModeUpdateResponse } from './role';

describe('parseModeUpdateResponse', () => {
  it('accepts a valid response', () => {
    const result = parseModeUpdateResponse({
      status: 'updated',
      mode: 'reflector',
      previous: 'test_master',
    });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.value.mode).toBe('reflector');
      expect(result.value.previous).toBe('test_master');
    }
  });

  it('rejects unknown mode value', () => {
    const result = parseModeUpdateResponse({
      status: 'updated',
      mode: 'passthrough', // not a valid StemRole
      previous: 'reflector',
    });
    expect(result.ok).toBe(false);
  });

  it('rejects missing status', () => {
    const result = parseModeUpdateResponse({
      mode: 'reflector',
      previous: 'test_master',
    });
    expect(result.ok).toBe(false);
  });

  it('rejects non-object input', () => {
    expect(parseModeUpdateResponse(null).ok).toBe(false);
    expect(parseModeUpdateResponse('reflector').ok).toBe(false);
    expect(parseModeUpdateResponse(42).ok).toBe(false);
  });

  it('accepts mode == previous (unchanged status)', () => {
    // "unchanged" status has mode == previous; sanity-confirm the
    // schema doesn't accidentally reject it.
    const result = parseModeUpdateResponse({
      status: 'unchanged',
      mode: 'reflector',
      previous: 'reflector',
    });
    expect(result.ok).toBe(true);
  });
});

describe('parseApiErrorBody', () => {
  it('returns parsed body when fields are valid', () => {
    expect(parseApiErrorBody({ message: 'something broke' })).toEqual({
      message: 'something broke',
    });
    expect(parseApiErrorBody({ error: 'fallback msg' })).toEqual({
      error: 'fallback msg',
    });
    expect(parseApiErrorBody({})).toEqual({});
  });

  it('returns null for non-objects', () => {
    expect(parseApiErrorBody(null)).toBeNull();
    expect(parseApiErrorBody('oops')).toBeNull();
    expect(parseApiErrorBody(42)).toBeNull();
  });

  it('returns null when fields are wrong type', () => {
    expect(parseApiErrorBody({ message: 42 })).toBeNull();
    expect(parseApiErrorBody({ error: true })).toBeNull();
  });
});

describe('isStemRole', () => {
  it('accepts the two valid roles', () => {
    expect(isStemRole('reflector')).toBe(true);
    expect(isStemRole('test_master')).toBe(true);
  });

  it('rejects everything else', () => {
    expect(isStemRole('passthrough')).toBe(false);
    expect(isStemRole('')).toBe(false);
    expect(isStemRole(null)).toBe(false);
    expect(isStemRole(undefined)).toBe(false);
    expect(isStemRole(42)).toBe(false);
  });
});
