/**
 * i18n.parity.test.ts — locks en/es locale parity in CI.
 *
 * The fixture list is DERIVED from the locale directories, not hand-written.
 * It used to be ten entries maintained by hand against eleven declared
 * namespaces, and `pages` — every route title and description — was the one
 * nobody added (#1252); deleting an `es/pages.json` key failed nothing. A
 * hand-maintained list of the things to check is a list that will be one
 * short again, so the glob below is the gate's real subject.
 *
 * Asserts two invariants for every shipped namespace:
 *   1. KEY PARITY  — en and es JSON files have identical key sets at every
 *      depth. Adding or removing a key in one language without the other
 *      fails CI.
 *   2. DNT COMPLIANCE — every industry-standard "Do Not Translate" term
 *      (acronyms, RFC numbers, protocol names, metrics, units) that appears
 *      in an en value must appear in the matching es value. Translating
 *      `throughput` to `rendimiento` or `latency` to `latencia` fails this
 *      gate.
 *
 * Match is case-insensitive (so a term at the start of a sentence still
 * counts) and word-boundary anchored (so `EIR` doesn't match `their`).
 *
 * Note: product/module names like Stem / NIAC / Measure / Benchmark are NOT
 * enforced by this gate — they collide with common English vocabulary used
 * as verbs ("Measure round-trip delay") and are handled by translator
 * discipline + code review instead.
 */

import { describe, expect, it } from 'vitest';
import { namespaces } from './index';

type Json = string | number | boolean | null | Json[] | { [k: string]: Json };

/** `../../locales/<lang>/<ns>.json` -> `<ns>`. */
function namespaceOf(path: string): string {
  return path.replace(/^.*\//, '').replace(/\.json$/, '');
}

function localeModules(mods: Record<string, unknown>): Map<string, Json> {
  return new Map(
    Object.entries(mods).map(([path, mod]) => [
      namespaceOf(path),
      (mod as { default: Json }).default,
    ]),
  );
}

const EN = localeModules(import.meta.glob('../../locales/en/*.json', { eager: true }));
const ES = localeModules(import.meta.glob('../../locales/es/*.json', { eager: true }));

const FIXTURES: { ns: string; en: Json; es: Json }[] = [...EN.keys()].sort().flatMap((ns) => {
  const es = ES.get(ns);
  return es === undefined ? [] : [{ ns, en: EN.get(ns) as Json, es }];
});

/**
 * Standard terms that must NEVER be translated. Acronyms / RFC numbers /
 * protocol names / metric names / units. Product/module names are excluded
 * because they collide with common English vocabulary (see file header).
 */
const DNT_TERMS = [
  // Standards
  'RFC 2544',
  'Y.1564',
  'Y.1731',
  'RFC 2889',
  'RFC 6349',
  'MEF',
  'TSN',
  // Protocols & acronyms
  'ARP',
  'DHCP',
  'DNS',
  'BGP',
  'OSPF',
  'SNMP',
  'VLAN',
  'WebSocket',
  // Metrics, abbreviations, units
  'SNR',
  'FLR',
  'FDV',
  'CIR',
  'EIR',
  'Mbps',
  'dBm',
  'jitter',
  'throughput',
  'latency',
];

function flatKeyPaths(node: Json, prefix = ''): string[] {
  if (node === null || typeof node !== 'object') return [prefix];
  if (Array.isArray(node)) {
    return node.flatMap((v, i) => flatKeyPaths(v, `${prefix}[${i}]`));
  }
  return Object.entries(node).flatMap(([k, v]) =>
    flatKeyPaths(v, prefix === '' ? k : `${prefix}.${k}`),
  );
}

function flatStringEntries(node: Json, prefix = ''): [string, string][] {
  if (typeof node === 'string') return [[prefix, node]];
  if (node === null || typeof node !== 'object') return [];
  if (Array.isArray(node)) {
    return node.flatMap((v, i) => flatStringEntries(v, `${prefix}[${i}]`));
  }
  return Object.entries(node).flatMap(([k, v]) =>
    flatStringEntries(v, prefix === '' ? k : `${prefix}.${k}`),
  );
}

/** Word-boundary, case-insensitive regex per DNT term. */
const DNT_PATTERNS: { term: string; rx: RegExp }[] = DNT_TERMS.map((term) => ({
  term,
  rx: new RegExp(`(?:^|[^\\w])${term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}(?:[^\\w]|$)`, 'i'),
}));

describe('i18n parity — en/es key sets', () => {
  for (const { ns, en, es } of FIXTURES) {
    it(`${ns}: identical key sets in en and es`, () => {
      const enK = new Set(flatKeyPaths(en));
      const esK = new Set(flatKeyPaths(es));
      const enOnly = [...enK].filter((k) => !esK.has(k)).sort();
      const esOnly = [...esK].filter((k) => !enK.has(k)).sort();
      expect(enOnly, 'keys present in en but missing in es').toEqual([]);
      expect(esOnly, 'keys present in es but missing in en').toEqual([]);
    });
  }
});

describe('i18n DNT — standard terms appear verbatim in es', () => {
  for (const { ns, en, es } of FIXTURES) {
    it(`${ns}: DNT terms in en values appear (case-insensitive) in matching es`, () => {
      const enMap = new Map(flatStringEntries(en));
      const esMap = new Map(flatStringEntries(es));
      const violations: string[] = [];
      for (const [path, enVal] of enMap) {
        const esVal = esMap.get(path);
        if (!esVal) continue;
        for (const { term, rx } of DNT_PATTERNS) {
          if (rx.test(enVal) && !rx.test(esVal)) {
            violations.push(`${path}: en has "${term}" but es does not`);
          }
        }
      }
      expect(violations).toEqual([]);
    });
  }
});

describe('i18n parity — the fixture list covers everything shipped', () => {
  it('has a fixture for every namespace i18n/index.ts declares', () => {
    const covered = new Set(FIXTURES.map((f) => f.ns));
    const missing = [...namespaces].filter((ns) => !covered.has(ns)).sort();
    expect(missing, 'declared namespaces with no parity fixture').toEqual([]);
  });

  it('ships the same namespace files in en and es', () => {
    expect([...ES.keys()].sort()).toEqual([...EN.keys()].sort());
  });
});
