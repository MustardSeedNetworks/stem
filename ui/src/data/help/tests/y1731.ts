/**
 * @fileoverview The Stem - Y.1731 Test Definitions
 * @description Help content for ITU-T Y.1731 OAM tests.
 */

import type { TestHelp } from '../types';

export const y1731Tests: Record<string, TestHelp> = {
  frame_delay: {
    id: 'frame_delay',
    name: 'Frame Delay Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Precise one-way and two-way delay measurements using OAM.',
    techDesc: 'Y.1731 DMM/DMR for precise delay measurement.',
    laymanDesc: 'Super-precise timing measurements for carrier networks.',
    whenToUse: 'SLA monitoring in production',
    whenNotToUse: 'Initial service turn-up',
    parameters: [],
    metrics: [],
    passCriteria: 'Delay within SLA',
    failMeaning: 'Exceeds SLA threshold',
    examples: [],
    tips: [],
    seeAlso: ['latency'],
  },

  y1731_frame_loss: {
    id: 'y1731_frame_loss',
    name: 'Frame Loss Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Monitors packet loss on production carrier networks.',
    techDesc: 'Y.1731 LMM/LMR for continuous loss monitoring.',
    laymanDesc: 'Continuously monitors if packets are being lost without disrupting traffic.',
    whenToUse: 'Continuous service monitoring',
    whenNotToUse: 'Initial service testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Loss within SLA',
    failMeaning: 'SLA may be violated',
    examples: [],
    tips: [],
    seeAlso: ['frame_loss'],
  },

  synthetic_loss: {
    id: 'synthetic_loss',
    name: 'Synthetic Loss Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Continuous reliability monitoring using test signals.',
    techDesc: 'SLM/SLR for loss measurement independent of user traffic.',
    laymanDesc: 'Sends special test signals to continuously check network health.',
    whenToUse: 'Links with variable traffic',
    whenNotToUse: 'High-traffic links',
    parameters: [],
    metrics: [],
    passCriteria: '0% synthetic loss',
    failMeaning: 'Network path has problems',
    examples: [],
    tips: [],
    seeAlso: ['y1731_frame_loss'],
  },

  loopback: {
    id: 'loopback',
    name: 'Loopback Test',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Quick connectivity check using OAM loopback.',
    techDesc: 'Y.1731 LBM/LBR for connectivity verification.',
    laymanDesc: 'A "ping" for carrier ethernet networks.',
    whenToUse: 'Quick connectivity verification',
    whenNotToUse: 'Performance testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Response received',
    failMeaning: 'Connectivity problem',
    examples: [],
    tips: [],
    seeAlso: ['frame_delay'],
  },
};
