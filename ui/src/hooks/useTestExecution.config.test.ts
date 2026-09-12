/**
 * @fileoverview buildTestConfig — which configuration block reaches the daemon.
 * @description The mapping from a test type to its config block decides what
 *              the daemon is actually told to run. A wrong entry does not
 *              fail: it silently measures something other than what the
 *              operator selected, with plausible-looking numbers. Every
 *              registered prefix is pinned here for that reason.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { describe, expect, it } from 'vitest';
import { defaultRFC2544Config } from '../components/RFC2544ConfigForm';
import { defaultRFC2889Config } from '../components/RFC2889ConfigForm';
import { defaultRFC6349Config } from '../components/RFC6349ConfigForm';
import { defaultTrafficGenConfig } from '../components/TrafficGenConfigForm';
import { defaultTSNConfig } from '../components/TSNConfigForm';
import { defaultY1564Config } from '../components/Y1564ConfigForm';
import { defaultY1731Config } from '../components/Y1731ConfigForm';
import { buildTestConfig } from './useTestExecution';

const configs = {
  rfc2544: defaultRFC2544Config,
  rfc2889: defaultRFC2889Config,
  rfc6349: defaultRFC6349Config,
  y1564: defaultY1564Config,
  y1731: defaultY1731Config,
  tsn: defaultTSNConfig,
  trafficGen: defaultTrafficGenConfig,
};

describe('buildTestConfig', () => {
  it.each([
    ['rfc2544_throughput', 'rfc2544'],
    ['rfc2544_latency', 'rfc2544'],
    ['rfc2889_congestion', 'rfc2889'],
    ['rfc6349_tcp_throughput', 'rfc6349'],
    ['y1564_config', 'y1564'],
    ['y1731_delay', 'y1731'],
    ['tsn_latency', 'tsn'],
  ])('sends %s under the %s block', (testType, expectedKey) => {
    const built = buildTestConfig(testType, configs);

    expect(built).toBeDefined();
    expect(Object.keys(built ?? {})).toEqual([expectedKey]);
    expect((built as Record<string, unknown>)[expectedKey]).toBe(
      configs[expectedKey as keyof typeof configs],
    );
  });

  it('sends custom_stream under trafficGen rather than by prefix', () => {
    const built = buildTestConfig('custom_stream', configs);

    expect(Object.keys(built ?? {})).toEqual(['trafficGen']);
    expect((built as Record<string, unknown>).trafficGen).toBe(configs.trafficGen);
  });

  it('sends no config for a type it does not recognise', () => {
    expect(buildTestConfig('reflect', configs)).toBeUndefined();
    expect(buildTestConfig('', configs)).toBeUndefined();
  });

  // A prefix match must not be so loose that one standard claims another's
  // test: rfc2544 and rfc2889 share three characters.
  it('does not let one standard claim another by prefix', () => {
    const built = buildTestConfig('rfc2889_broadcast', configs);

    expect(Object.keys(built ?? {})).toEqual(['rfc2889']);
  });
});
