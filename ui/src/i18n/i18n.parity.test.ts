/**
 * i18n.parity.test.ts — locks en/es locale parity in CI.
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

import enCli from '@locales/en/cli.json';
import enCommon from '@locales/en/common.json';
import enErrors from '@locales/en/errors.json';
import enHelp from '@locales/en/help.json';
import enModules from '@locales/en/modules.json';
import enParams from '@locales/en/params.json';
import enRecovery from '@locales/en/recovery.json';
import enSecurity from '@locales/en/security.json';
import enSettings from '@locales/en/settings.json';
import enSetup from '@locales/en/setup.json';
import esCli from '@locales/es/cli.json';
import esCommon from '@locales/es/common.json';
import esErrors from '@locales/es/errors.json';
import esHelp from '@locales/es/help.json';
import esModules from '@locales/es/modules.json';
import esParams from '@locales/es/params.json';
import esRecovery from '@locales/es/recovery.json';
import esSecurity from '@locales/es/security.json';
import esSettings from '@locales/es/settings.json';
import esSetup from '@locales/es/setup.json';
import { describe, expect, it } from 'vitest';

type Json = string | number | boolean | null | Json[] | { [k: string]: Json };

const FIXTURES: { ns: string; en: Json; es: Json }[] = [
  { ns: 'cli', en: enCli as Json, es: esCli as Json },
  { ns: 'common', en: enCommon as Json, es: esCommon as Json },
  { ns: 'errors', en: enErrors as Json, es: esErrors as Json },
  { ns: 'help', en: enHelp as Json, es: esHelp as Json },
  { ns: 'modules', en: enModules as Json, es: esModules as Json },
  { ns: 'params', en: enParams as Json, es: esParams as Json },
  { ns: 'recovery', en: enRecovery as Json, es: esRecovery as Json },
  { ns: 'security', en: enSecurity as Json, es: esSecurity as Json },
  { ns: 'settings', en: enSettings as Json, es: esSettings as Json },
  { ns: 'setup', en: enSetup as Json, es: esSetup as Json },
];

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
