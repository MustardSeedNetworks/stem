/**
 * @fileoverview The Stem - MEF Test Definitions
 * @description Help content for Metro Ethernet Forum service tests.
 */

import type { TestHelp } from '../types';

export const mefTests: Record<string, TestHelp> = {
  mef_config: {
    id: 'mef_config',
    name: 'MEF Service Configuration Test',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Validates carrier ethernet service per MEF standards.',
    techDesc: 'Tests bandwidth profiles and CoS per MEF specifications.',
    laymanDesc: 'The official carrier ethernet validation per industry standards.',
    whenToUse: 'MEF-certified service validation',
    whenNotToUse: 'Simple single-class services',
    parameters: [],
    metrics: [],
    passCriteria: 'Bandwidth and CoS compliance',
    failMeaning: 'Service not meeting MEF specs',
    examples: [],
    tips: [],
    seeAlso: ['y1564_config'],
  },

  mef_performance: {
    id: 'mef_performance',
    name: 'MEF Performance Test',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Extended MEF service quality validation.',
    techDesc: 'Extended duration tests per MEF specifications.',
    laymanDesc: 'Long-running test for carrier service quality.',
    whenToUse: 'After MEF Config test passes',
    whenNotToUse: 'Quick spot checks',
    parameters: [],
    metrics: [],
    passCriteria: 'Performance within specs',
    failMeaning: 'Service shows instability',
    examples: [],
    tips: [],
    seeAlso: ['mef_config'],
  },

  mef_full: {
    id: 'mef_full',
    name: 'MEF Full Test Suite',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Complete MEF service validation.',
    techDesc: 'Complete MEF validation sequence.',
    laymanDesc: 'The complete MEF certification test.',
    whenToUse: 'Official MEF service acceptance',
    whenNotToUse: 'Troubleshooting',
    parameters: [],
    metrics: [],
    passCriteria: 'All MEF tests pass',
    failMeaning: 'Service not MEF compliant',
    examples: [],
    tips: [],
    seeAlso: ['mef_config', 'mef_performance'],
  },
};
