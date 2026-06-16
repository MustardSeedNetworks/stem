/**
 * @fileoverview The Stem - Help Content Types
 * @description Shared interfaces for help content data structures.
 */

export interface TestHelp {
  id: string;
  name: string;
  standard: string;
  category: string;
  summary: string;
  techDesc: string;
  laymanDesc: string;
  whenToUse: string;
  whenNotToUse: string;
  parameters: Parameter[];
  metrics: Metric[];
  passCriteria: string;
  failMeaning: string;
  examples: Example[];
  tips: string[];
  seeAlso: string[];
}

export interface Parameter {
  name: string;
  flag: string;
  type: string;
  defaultValue: string;
  required: boolean;
  techDesc: string;
  laymanDesc: string;
  example: string;
}

export interface Metric {
  name: string;
  unit: string;
  goodRange: string;
  badMeaning: string;
}

export interface Example {
  desc: string;
  command: string;
  output?: string;
}

export interface Category {
  id: string;
  name: string;
  fullName: string;
  summary: string;
  description: string;
  tests: string[];
  whenToUse: string;
}

export interface GlossaryEntry {
  term: string;
  fullName: string;
  category: string;
  techDef: string;
  laymanDef: string;
  related: string[];
}

export interface TutorialStep {
  title: string;
  content: string;
  command?: string;
  expected?: string;
  tip?: string;
}

export interface Tutorial {
  id: string;
  title: string;
  duration: string;
  level: 'Beginner' | 'Intermediate' | 'Advanced';
  description: string;
  steps: TutorialStep[];
}
