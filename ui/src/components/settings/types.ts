/**
 * Settings Types
 *
 * Shared type definitions for settings components.
 * Aligned with Seed patterns for future profile integration.
 */

export interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number;
  state: string;
  driver: string;
  physical: boolean;
  xdp: boolean;
  score: number;
}

export type OperatingMode = 'reflector' | 'test_master';

export type ReflectorProfile = 'netally' | 'msn' | 'all' | 'custom';

/**
 * Test definition for rendering test checkboxes.
 */
export interface TestDefinition {
  id: string;
  name: string;
  desc: string;
  tooltip: string;
}

/**
 * Common props for test section components.
 */
export interface TestSectionProps {
  selectedTests: string[];
  onToggleTest: (testId: string) => void;
}

/**
 * Shared settings section props.
 */
export interface SettingsSectionProps {
  className?: string;
}
