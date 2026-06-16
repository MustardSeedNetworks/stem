/**
 * @fileoverview The Stem - RFC 6349 Test Definitions
 * @description Help content for RFC 6349 TCP throughput testing.
 */

import type { TestHelp } from '../types';

export const rfc6349Tests: Record<string, TestHelp> = {
  tcp_throughput: {
    id: 'tcp_throughput',
    name: 'TCP Throughput Test',
    standard: 'RFC 6349',
    category: 'RFC 6349',
    summary: 'Measures real application throughput using TCP.',
    techDesc: 'Tests TCP throughput accounting for protocol behavior.',
    laymanDesc: 'This measures REAL download/upload speeds you actually experience.',
    whenToUse: 'Application performance troubleshooting',
    whenNotToUse: 'Layer 2 equipment testing',
    parameters: [],
    metrics: [],
    passCriteria: 'TCP throughput meets requirements',
    failMeaning: 'Network may need optimization',
    examples: [],
    tips: [],
    seeAlso: ['path_analysis', 'throughput'],
  },

  path_analysis: {
    id: 'path_analysis',
    name: 'Path Analysis Test',
    standard: 'RFC 6349',
    category: 'RFC 6349',
    summary: "Analyzes what's limiting your network speed.",
    techDesc: 'Characterizes RTT, bottleneck bandwidth, and BDP.',
    laymanDesc: 'Answers WHY your connection is slow, not just HOW slow.',
    whenToUse: 'TCP troubleshooting',
    whenNotToUse: 'If you just need throughput numbers',
    parameters: [],
    metrics: [],
    passCriteria: 'Identifies optimization opportunities',
    failMeaning: 'N/A - diagnostic test',
    examples: [],
    tips: [],
    seeAlso: ['tcp_throughput'],
  },
};
