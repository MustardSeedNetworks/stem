// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * Profile Types
 *
 * Type definitions for test configuration profiles.
 * Aligned with Seed patterns for database-backed settings storage.
 *
 * Architecture:
 * - Profiles contain all test configuration as a single config object
 * - Default hierarchy: Profile → Backend Defaults → Hardcoded Defaults
 * - Profiles are the single source of truth for settings
 */

import type { RFC2544Config } from '../components/RFC2544ConfigForm';
import type { RFC2889Config } from '../components/Rfc2889ConfigForm';
import type { RFC6349Config } from '../components/Rfc6349ConfigForm';
import type { ReflectorProfile } from '../components/settings/types';
import type { TrafficGenConfig } from '../components/TrafficGenConfigForm';
import type { TSNConfig } from '../components/TsnConfigForm';
import type { Y1564Config } from '../components/Y1564ConfigForm';
import type { Y1731Config } from '../components/Y1731ConfigForm';

// =============================================================================
// Core Profile Types
// =============================================================================

/**
 * Test profile containing all configuration settings.
 * Stored in database with config_json blob.
 */
export interface Profile {
  /** Unique identifier (UUID) */
  id: string;
  /** Display name */
  name: string;
  /** Optional description */
  description?: string;
  /** All profile configuration */
  config: ProfileConfig;
  /** Whether this is the default profile */
  isDefault: boolean;
  /** Creation timestamp (ISO 8601) */
  createdAt: string;
  /** Last update timestamp (ISO 8601) */
  updatedAt: string;
}

/**
 * Complete configuration stored in a profile.
 * Contains all settings organized by category.
 */
export interface ProfileConfig {
  /** General settings */
  general?: GeneralSettings;
  /** Test configurations by standard */
  tests?: TestConfigs;
  /** Interface settings */
  interfaces?: InterfaceSettings;
  /** Thresholds for pass/fail criteria */
  thresholds?: ThresholdSettings;
  /** Display preferences */
  display?: DisplaySettings;
}

// =============================================================================
// Settings Categories
// =============================================================================

/**
 * General application settings.
 */
export interface GeneralSettings {
  /** Operating mode */
  mode: 'reflector' | 'test_master';
  /** Reflector profile when in reflector mode */
  reflectorProfile: ReflectorProfile;
  /** Default test duration in seconds */
  defaultDuration?: number;
  /** Auto-save results */
  autoSave?: boolean;
}

/**
 * Test configurations for each standard.
 */
export interface TestConfigs {
  /** RFC 2544 configuration */
  rfc2544?: RFC2544Config;
  /** RFC 2889 configuration */
  rfc2889?: RFC2889Config;
  /** RFC 6349 configuration */
  rfc6349?: RFC6349Config;
  /** Y.1564 / EtherSAM configuration */
  y1564?: Y1564Config;
  /** Y.1731 OAM configuration */
  y1731?: Y1731Config;
  /** TSN 802.1Qbv configuration */
  tsn?: TSNConfig;
  /** Traffic generator configuration */
  trafficGen?: TrafficGenConfig;
  /** Selected tests */
  selectedTests?: string[];
}

/**
 * Interface-specific settings.
 */
export interface InterfaceSettings {
  /** Preferred interface name */
  preferredInterface?: string;
  /** Interface-specific thresholds */
  interfaceThresholds?: Record<string, InterfaceThreshold>;
}

/**
 * Per-interface threshold configuration.
 */
export interface InterfaceThreshold {
  /** Minimum acceptable throughput (Mbps) */
  minThroughput?: number;
  /** Maximum acceptable latency (µs) */
  maxLatency?: number;
  /** Maximum acceptable jitter (µs) */
  maxJitter?: number;
  /** Maximum acceptable frame loss (%) */
  maxFrameLoss?: number;
}

/**
 * Threshold settings for pass/fail determination.
 */
export interface ThresholdSettings {
  /** Throughput thresholds */
  throughput?: {
    good: number;
    acceptable: number;
    unit: 'Mbps' | 'Gbps';
  };
  /** Latency thresholds */
  latency?: {
    good: number;
    acceptable: number;
    unit: 'us' | 'ms';
  };
  /** Jitter thresholds */
  jitter?: {
    good: number;
    acceptable: number;
    unit: 'us' | 'ms';
  };
  /** Frame loss thresholds */
  frameLoss?: {
    good: number;
    acceptable: number;
    unit: '%';
  };
}

/**
 * Display and UI preferences.
 */
export interface DisplaySettings {
  /** Dark mode preference */
  darkMode?: boolean;
  /** Preferred view mode for test selection */
  testViewMode?: 'standard' | 'module';
  /** Show advanced options */
  showAdvanced?: boolean;
  /** Decimal precision for results */
  decimalPrecision?: number;
}

// =============================================================================
// API Types
// =============================================================================

/**
 * Profile list item (minimal data for listings).
 */
export interface ProfileListItem {
  id: string;
  name: string;
  description?: string;
  isDefault: boolean;
  updatedAt: string;
}

/**
 * Request to create a new profile.
 */
export interface CreateProfileRequest {
  name: string;
  description?: string;
  config: ProfileConfig;
  isDefault?: boolean;
}

/**
 * Request to update an existing profile.
 */
export interface UpdateProfileRequest {
  name?: string;
  description?: string;
  config?: Partial<ProfileConfig>;
  isDefault?: boolean;
}

/**
 * Profile import/export format.
 */
export interface ProfileExport {
  version: '1.0';
  exportedAt: string;
  profiles: Profile[];
}

/**
 * Profile import request.
 */
export interface ProfileImportRequest {
  version: string;
  profiles: Profile[];
  /** Whether to overwrite existing profiles with same name */
  overwrite?: boolean;
}

/**
 * Profile import result.
 */
export interface ProfileImportResult {
  created: number;
  updated: number;
  skipped: number;
  errors: string[];
}

// =============================================================================
// Store Types
// =============================================================================

/**
 * Profile store state.
 */
export interface ProfileStoreState {
  /** All available profiles */
  profiles: Profile[];
  /** Currently active profile */
  activeProfile: Profile | null;
  /** Backend default settings */
  backendDefaults: ProfileConfig | null;
  /** Loading state */
  isLoading: boolean;
  /** Error message if any */
  error: string | null;
}

/**
 * Profile store actions.
 */
export interface ProfileStoreActions {
  /** Load all profiles from backend */
  loadProfiles: () => Promise<void>;
  /** Get active profile from backend */
  loadActiveProfile: () => Promise<void>;
  /** Switch to a different profile */
  switchProfile: (profileId: string) => Promise<boolean>;
  /** Create a new profile */
  createProfile: (request: CreateProfileRequest) => Promise<Profile>;
  /** Update an existing profile */
  updateProfile: (id: string, request: UpdateProfileRequest) => Promise<Profile>;
  /** Delete a profile */
  deleteProfile: (id: string) => Promise<void>;
  /** Duplicate a profile */
  duplicateProfile: (id: string, newName: string) => Promise<Profile>;
  /** Set a profile as default */
  setDefaultProfile: (id: string) => Promise<void>;
  /** Update settings in active profile */
  updateSettings: <K extends keyof ProfileConfig>(
    category: K,
    settings: ProfileConfig[K],
  ) => Promise<void>;
  /** Export profiles */
  exportProfiles: () => Promise<ProfileExport>;
  /** Import profiles */
  importProfiles: (data: ProfileImportRequest) => Promise<ProfileImportResult>;
}

// =============================================================================
// Default Values
// =============================================================================

/**
 * Default general settings.
 */
export const DEFAULT_GENERAL_SETTINGS: GeneralSettings = {
  mode: 'test_master',
  reflectorProfile: 'all',
  defaultDuration: 60,
  autoSave: true,
};

/**
 * Default threshold settings.
 */
export const DEFAULT_THRESHOLD_SETTINGS: ThresholdSettings = {
  throughput: { good: 900, acceptable: 800, unit: 'Mbps' },
  latency: { good: 1000, acceptable: 5000, unit: 'us' },
  jitter: { good: 100, acceptable: 500, unit: 'us' },
  frameLoss: { good: 0, acceptable: 0.1, unit: '%' },
};

/**
 * Default display settings.
 */
export const DEFAULT_DISPLAY_SETTINGS: DisplaySettings = {
  darkMode: false,
  testViewMode: 'standard',
  showAdvanced: false,
  decimalPrecision: 2,
};

/**
 * Default profile configuration.
 */
export const DEFAULT_PROFILE_CONFIG: ProfileConfig = {
  general: DEFAULT_GENERAL_SETTINGS,
  thresholds: DEFAULT_THRESHOLD_SETTINGS,
  display: DEFAULT_DISPLAY_SETTINGS,
};
