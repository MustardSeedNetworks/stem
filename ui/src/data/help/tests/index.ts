/**
 * @fileoverview The Stem - Tests Index
 * @description Aggregates all test partials and provides search/lookup helpers.
 */

import type { TestHelp } from '../types';
import { categories } from '../categories';
import { customTests } from './custom';
import { mefTests } from './mef';
import { rfc2544Tests } from './rfc2544';
import { rfc2889Tests } from './rfc2889';
import { rfc6349Tests } from './rfc6349';
import { tsnTests } from './tsn';
import { y1564Tests } from './y1564';
import { y1731Tests } from './y1731';

export const tests: Record<string, TestHelp> = {
  ...rfc2544Tests,
  ...y1564Tests,
  ...rfc2889Tests,
  ...rfc6349Tests,
  ...y1731Tests,
  ...mefTests,
  ...tsnTests,
  ...customTests,
};

// Helper function to get tests by category
export function getTestsByCategory(categoryId: string): TestHelp[] {
  const cat = categories[categoryId];
  if (!cat) {
    return [];
  }
  return cat.tests.map((id) => tests[id]).filter(Boolean);
}

// Helper to search tests
export function searchTests(keyword: string): TestHelp[] {
  const lower = keyword.toLowerCase();
  return Object.values(tests).filter(
    (t) =>
      t.name.toLowerCase().includes(lower) ||
      t.summary.toLowerCase().includes(lower) ||
      t.techDesc.toLowerCase().includes(lower) ||
      t.laymanDesc.toLowerCase().includes(lower),
  );
}
