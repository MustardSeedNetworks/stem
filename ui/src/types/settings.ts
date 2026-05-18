/**
 * @fileoverview The Stem - TypeScript Type Definitions
 * @description Defines all TypeScript interfaces and types used throughout the WebUI.
 *              Includes interface info, test configurations, license info, and app settings.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

/** Network interface information from the backend */
export interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number; // Mbps
  duplex: string; // full, half, unknown
  state: string; // up, down
  driver: string;
  physical: boolean;
  xdp: boolean;
  dpdk: boolean;
  score: number;
  mtu: number;
  ipv4: string;
  ipv6: string;
}

// Runtime statistics
export interface Stats {
  packetsReceived: number;
  packetsSent: number;
  bytesReceived: number;
  bytesSent: number;
  currentPps: number;
  currentMbps: number;
  uptime: number;
  testStatus: TestStatus;
  currentTest: string | null;
}

export type TestStatus = 'idle' | 'running' | 'completed' | 'error';

// License tiers
export type LicenseTier = 1 | 2;

// License information
export interface LicenseInfo {
  key: string;
  tier: LicenseTier;
  activated: boolean;
  expiresAt: string;
  deviceCount: number;
  maxDevices: number;
}

// Application mode
export type AppMode = 'reflector' | 'test_master';

// Theme preference
export type Theme = 'light' | 'dark' | 'system';

// Reflector profile presets
export type ReflectorProfile = 'netally' | 'msn' | 'all' | 'custom';

// Reflector configuration
export interface ReflectorConfig {
  profile: ReflectorProfile;
  signatureFilter: string[];
  ouiFilter: string;
  portFilter: number;
}

// RFC 2544 test configuration
export interface RFC2544Config {
  throughput: boolean;
  latency: boolean;
  frameLoss: boolean;
  backToBack: boolean;
  systemRecovery: boolean;
  reset: boolean;
  frameSizes: number[];
  duration: number;
  iterations: number;
}

// Y.1564 / EtherSAM configuration
export interface Y1564Config {
  configurationTest: boolean;
  performanceTest: boolean;
  fullTest: boolean;
  cir: number; // Committed Information Rate (Kbps)
  eir: number; // Excess Information Rate (Kbps)
  delayThreshold: number; // ms
  jitterThreshold: number; // ms
  lossThreshold: number; // percentage
}

// RFC 2889 LAN Switch configuration
export interface RFC2889Config {
  forwardingRate: boolean;
  addressCaching: boolean;
  addressLearning: boolean;
  broadcast: boolean;
  congestionControl: boolean;
  portCount: number;
  macAddressCount: number;
}

// RFC 6349 TCP configuration
export interface RFC6349Config {
  tcpThroughput: boolean;
  pathAnalysis: boolean;
  mss: number; // Maximum Segment Size
  rwnd: number; // Receive Window
  duration: number;
}

// Y.1731 OAM configuration
export interface Y1731Config {
  delay: boolean;
  loss: boolean;
  syntheticLoss: boolean;
  loopback: boolean;
  mepId: number;
  megId: string;
  domain: string;
}

// MEF Service configuration
export interface MEFConfig {
  configuration: boolean;
  performance: boolean;
  fullTest: boolean;
  serviceType: 'epl' | 'evpl' | 'eplan' | 'evplan';
  cosClasses: number;
}

// TSN 802.1Qbv configuration
export interface TSNConfig {
  gateTiming: boolean;
  trafficIsolation: boolean;
  scheduledLatency: boolean;
  fullSuite: boolean;
  cyclePeriod: number; // microseconds
  gateControlList: string;
}

// Complete test configuration
export interface TestConfig {
  rfc2544: RFC2544Config;
  y1564: Y1564Config;
  rfc2889: RFC2889Config;
  rfc6349: RFC6349Config;
  y1731: Y1731Config;
  mef: MEFConfig;
  tsn: TSNConfig;
}

// Application settings
export interface AppSettings {
  mode: AppMode;
  license: LicenseInfo;
  interface: {
    name: string;
    autoSelect: boolean;
  };
  reflector: ReflectorConfig;
  tests: TestConfig;
  appearance: {
    theme: Theme;
  };
}

// Default configurations
export const defaultRFC2544Config: RFC2544Config = {
  throughput: true,
  latency: true,
  frameLoss: true,
  backToBack: true,
  systemRecovery: false,
  reset: false,
  frameSizes: [64, 128, 256, 512, 1024, 1280, 1518],
  duration: 60,
  iterations: 1,
};

export const defaultY1564Config: Y1564Config = {
  configurationTest: true,
  performanceTest: true,
  fullTest: false,
  cir: 1000000, // 1 Gbps
  eir: 0,
  delayThreshold: 50,
  jitterThreshold: 10,
  lossThreshold: 0.1,
};

export const defaultRFC2889Config: RFC2889Config = {
  forwardingRate: true,
  addressCaching: false,
  addressLearning: false,
  broadcast: false,
  congestionControl: false,
  portCount: 2,
  macAddressCount: 100,
};

export const defaultRFC6349Config: RFC6349Config = {
  tcpThroughput: true,
  pathAnalysis: false,
  mss: 1460,
  rwnd: 65535,
  duration: 60,
};

export const defaultY1731Config: Y1731Config = {
  delay: true,
  loss: true,
  syntheticLoss: false,
  loopback: false,
  mepId: 1,
  megId: 'MSN-MEG',
  domain: 'customer',
};

export const defaultMEFConfig: MEFConfig = {
  configuration: true,
  performance: true,
  fullTest: false,
  serviceType: 'epl',
  cosClasses: 4,
};

export const defaultTSNConfig: TSNConfig = {
  gateTiming: true,
  trafficIsolation: false,
  scheduledLatency: false,
  fullSuite: false,
  cyclePeriod: 1000,
  gateControlList: '',
};

