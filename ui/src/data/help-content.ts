/**
 * @fileoverview The Stem - Help Content for WebUI
 * @description Barrel re-export — all help content now lives under data/help/*.
 *              This file exists solely to preserve the existing import paths.
 */

export { categories } from './help/categories';
export { glossary, searchGlossary } from './help/glossary';
export { getTestsByCategory, searchTests, tests } from './help/tests/index';
export { tutorials } from './help/tutorials';
export type {
  Category,
  Example,
  GlossaryEntry,
  Metric,
  Parameter,
  TestHelp,
  Tutorial,
  TutorialStep,
} from './help/types';
