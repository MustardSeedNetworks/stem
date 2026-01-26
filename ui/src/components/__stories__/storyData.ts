// Shared story data for Stem component stories.

import type { InterfaceInfo } from '../settings/types';

export const sampleInterfaces: InterfaceInfo[] = [
  {
    name: 'eth0',
    mac: 'aa:bb:cc:dd:ee:ff',
    speed: 1000,
    state: 'up',
    driver: 'e1000e',
    physical: true,
    xdp: true,
    score: 95,
  },
  {
    name: 'eth1',
    mac: '11:22:33:44:55:66',
    speed: 1000,
    state: 'down',
    driver: 'igc',
    physical: true,
    xdp: false,
    score: 70,
  },
];

export const selectedRFC2544Tests: string[] = ['rfc2544_throughput', 'rfc2544_latency'];
export const selectedY1564Tests: string[] = ['y1564_config'];
export const selectedRFC2889Tests: string[] = ['rfc2889_forwarding'];
export const selectedRFC6349Tests: string[] = ['rfc6349_throughput', 'rfc6349_bdp'];
export const selectedY1731Tests: string[] = ['y1731_delay'];
export const selectedTSNTests: string[] = ['tsn_timing'];
export const selectedTrafficGenTests: string[] = ['custom_stream'];
export const selectedMEFTests: string[] = ['mef_config'];
