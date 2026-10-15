/**
 * @fileoverview Stem - Y.1564 Test Definitions
 * @description Help content for ITU-T Y.1564 service activation tests.
 */

import type { TestHelp } from '../types';

export const y1564Tests: Record<string, TestHelp> = {
  y1564_config: {
    id: 'y1564_config',
    name: 'Y.1564 Service Configuration Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Validates carrier ethernet service at 25%, 50%, 75%, and 100% of committed rate.',
    techDesc: `The Y.1564 Service Configuration Test validates that an Ethernet service
meets its SLA parameters at progressive load steps (25%, 50%, 75%, 100% of CIR).`,
    laymanDesc: `When you buy an ethernet service from a carrier, this test verifies
you're getting what you paid for at different load levels.`,
    whenNotToUse: '• For raw equipment benchmarking (use RFC 2544)',
    parameters: [
      {
        name: 'CIR',
        setting: 'y1564.cir',
        type: 'float (Mbps)',
        defaultValue: '1000',
        required: true,
        techDesc: 'Committed Information Rate',
        laymanDesc: 'The speed your contract guarantees',
      },
      {
        name: 'EIR',
        setting: 'y1564.eir',
        type: 'float (Mbps)',
        defaultValue: '0',
        required: false,
        techDesc: 'Excess Information Rate - bandwidth above CIR that may be available',
        laymanDesc: 'Extra burst speed you might get when network is not busy',
      },
      {
        name: 'CBS',
        setting: 'y1564.cbs',
        type: 'integer (KB)',
        defaultValue: '0',
        required: false,
        techDesc: 'Committed Burst Size - maximum burst at CIR',
        laymanDesc: 'How much data can be sent in a burst at guaranteed speed',
      },
      {
        name: 'EBS',
        setting: 'y1564.ebs',
        type: 'integer (KB)',
        defaultValue: '0',
        required: false,
        techDesc: 'Excess Burst Size - maximum burst at EIR',
        laymanDesc: 'How much data can be sent in a burst at excess speed',
      },
      {
        name: 'Frame Sizes',
        setting: 'y1564.frameSizes',
        type: 'comma-separated integers',
        defaultValue: '64,512,1518',
        required: false,
        techDesc: 'Ethernet frame sizes in bytes to test',
        laymanDesc: 'Different packet sizes to validate service with',
      },
      {
        name: 'Step Duration',
        setting: 'y1564.configStepDuration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of each configuration step (25%, 50%, 75%, 100%)',
        laymanDesc: 'How long to test at each speed level',
      },
      {
        name: 'FLR Threshold',
        setting: 'y1564.flrThreshold',
        type: 'float (percentage)',
        defaultValue: '0.0',
        required: false,
        techDesc: 'Frame Loss Ratio acceptance threshold',
        laymanDesc: 'Maximum acceptable packet loss percentage',
      },
      {
        name: 'FD Threshold',
        setting: 'y1564.fdThreshold',
        type: 'float (ms)',
        defaultValue: '10.0',
        required: false,
        techDesc: 'Frame Delay acceptance threshold in milliseconds',
        laymanDesc: 'Maximum acceptable delay',
      },
      {
        name: 'FDV Threshold',
        setting: 'y1564.fdvThreshold',
        type: 'float (ms)',
        defaultValue: '5.0',
        required: false,
        techDesc: 'Frame Delay Variation acceptance threshold in milliseconds',
        laymanDesc: 'Maximum acceptable jitter',
      },
    ],
    metrics: [
      {
        name: 'IR (Information Rate)',
        unit: 'Mbps',
        goodRange: 'Within 1% of configured CIR',
        badMeaning: 'Service not delivering promised bandwidth',
      },
      {
        name: 'FD (Frame Delay)',
        unit: 'milliseconds',
        goodRange: 'Below threshold at all steps',
        badMeaning: 'Exceeds SLA delay commitment',
      },
      {
        name: 'FLR (Frame Loss Ratio)',
        unit: 'percentage',
        goodRange: 'Below configured threshold',
        badMeaning: 'Packets being lost above acceptable level',
      },
      {
        name: 'FDV (Frame Delay Variation)',
        unit: 'milliseconds',
        goodRange: 'Below configured threshold',
        badMeaning: 'Jitter exceeds SLA commitment',
      },
    ],
    passCriteria: 'All metrics within thresholds at all four CIR steps',
    failMeaning: 'Service does not meet SLA',
    examples: [
      {
        desc: 'Test 100 Mbps service',
        command: 'stem test -i eth0 --peer 192.0.2.10 -t y1564_config --cir 100',
        output: 'All steps: PASS',
      },
    ],
    tips: ['CIR should match your contract exactly'],
    seeAlso: ['https://www.itu.int/rec/T-REC-Y.1564', 'y1564_perf', 'mef_config'],
  },

  y1564_perf: {
    id: 'y1564_perf',
    name: 'Y.1564 Service Performance Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Extended duration test to validate service quality over time.',
    techDesc: `The Y.1564 Service Performance Test validates sustained performance
over an extended period (typically 15 minutes to hours).`,
    laymanDesc: `After passing the initial speed test, can your network connection
maintain that performance for hours?`,
    whenNotToUse: '• Initial service turn-up (do Config Test first)',
    parameters: [
      {
        name: 'Duration',
        setting: 'y1564.perfTestDuration',
        type: 'integer (minutes)',
        defaultValue: '15',
        required: false,
        techDesc: 'Test duration in minutes',
        laymanDesc: 'How long to run the test',
      },
    ],
    metrics: [
      {
        name: 'Sustained Rate',
        unit: 'Mbps',
        goodRange: 'Within 1% of CIR for entire duration',
        badMeaning: 'Performance degrades over time',
      },
    ],
    passCriteria: 'All metrics within thresholds for entire duration',
    failMeaning: 'Service shows instability over time',
    examples: [
      {
        desc: '15-minute test',
        command: 'stem test -i eth0 --peer 192.0.2.10 -t y1564_perf --cir 100 --duration 900',
        output: 'Performance stable',
      },
    ],
    tips: [],
    seeAlso: ['https://www.itu.int/rec/T-REC-Y.1564', 'y1564_config'],
  },

  y1564: {
    id: 'y1564',
    name: 'Y.1564 Full SAC Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Complete Service Activation Test - Configuration followed by Performance.',
    techDesc: `The Full SAC Test combines both Configuration and Performance tests
into a complete validation sequence.`,
    laymanDesc: `The complete, official test for verifying a carrier ethernet service.
This is what carriers use to officially "turn up" a new service.`,
    whenNotToUse: '• Quick troubleshooting',
    parameters: [],
    metrics: [],
    passCriteria: 'Both Configuration and Performance tests pass',
    failMeaning: 'Service does not meet acceptance criteria',
    examples: [
      {
        desc: 'Full SAC test',
        command: 'stem test -i eth0 --peer 192.0.2.10 -t y1564 --cir 100',
        output: 'Service Accepted',
      },
    ],
    tips: [],
    seeAlso: ['https://www.itu.int/rec/T-REC-Y.1564', 'y1564_config', 'y1564_perf'],
  },
};
