/**
 * Runtime type-guard tests for the API response validators.
 *
 * These three guards are the last line between an unexpected response shape
 * and the code that reads it, so what matters is that each one rejects the
 * near-misses — not just that it accepts a well-formed value. Every case below
 * that should be rejected differs from a valid one by a single field.
 */

import { describe, expect, it } from 'vitest';
import { isValidAuthResponse, isValidInterfaceArray, isValidStats } from './api';

describe('isValidInterfaceArray', () => {
  const validInterface = { name: 'eth0', mac: '00:11:22:33:44:55', speed: 1000 };

  it('accepts an array of well-formed interfaces', () => {
    expect(isValidInterfaceArray([validInterface])).toBe(true);
  });

  it('accepts an empty array', () => {
    // No interfaces is a legitimate answer from a host with none bound, and is
    // distinct from a malformed response.
    expect(isValidInterfaceArray([])).toBe(true);
  });

  it.each([
    ['not an array', validInterface],
    ['null', null],
    ['undefined', undefined],
    ['a JSON string that was never parsed', JSON.stringify([validInterface])],
  ])('rejects %s', (_label, value) => {
    expect(isValidInterfaceArray(value)).toBe(false);
  });

  it.each([
    ['name missing', { mac: '00:11:22:33:44:55', speed: 1000 }],
    ['mac missing', { name: 'eth0', speed: 1000 }],
    ['speed missing', { name: 'eth0', mac: '00:11:22:33:44:55' }],
    ['speed sent as a string', { ...validInterface, speed: '1000' }],
    ['name sent as a number', { ...validInterface, name: 0 }],
    ['a null entry', null],
  ])('rejects an array whose entry has %s', (_label, entry) => {
    expect(isValidInterfaceArray([entry])).toBe(false);
  });

  it('rejects when only one entry of several is malformed', () => {
    // every() short-circuits, so a guard that checked only the first element
    // would pass this.
    expect(isValidInterfaceArray([validInterface, validInterface, { name: 'eth2' }])).toBe(false);
  });
});

describe('isValidStats', () => {
  it.each([
    ['uptime alone', { uptime: 42 }],
    ['packetsReceived alone', { packetsReceived: 7 }],
    ['both', { uptime: 42, packetsReceived: 7 }],
  ])('accepts %s', (_label, value) => {
    expect(isValidStats(value)).toBe(true);
  });

  it.each([
    ['null', null],
    ['undefined', undefined],
    ['a number', 42],
    ['a string', 'uptime'],
    ['an object with neither field', { somethingElse: 1 }],
    ['uptime sent as a string', { uptime: '42' }],
    ['packetsReceived sent as a string', { packetsReceived: '7' }],
  ])('rejects %s', (_label, value) => {
    expect(isValidStats(value)).toBe(false);
  });
});

describe('isValidAuthResponse', () => {
  it('accepts a response carrying a string token', () => {
    expect(isValidAuthResponse({ token: 'abc' })).toBe(true);
  });

  it('accepts an empty token string', () => {
    // The guard is a shape check, not a credential check: whether an empty
    // token is usable is the auth store's decision, not this function's.
    expect(isValidAuthResponse({ token: '' })).toBe(true);
  });

  it.each([
    ['null', null],
    ['undefined', undefined],
    ['a bare string', 'abc'],
    ['an object with no token', { expiresIn: 3600 }],
    ['a numeric token', { token: 1234 }],
    ['a null token', { token: null }],
  ])('rejects %s', (_label, value) => {
    expect(isValidAuthResponse(value)).toBe(false);
  });
});
