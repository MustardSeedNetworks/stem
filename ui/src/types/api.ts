/**
 * @fileoverview Shared API types for The Stem
 * @description Centralized type definitions for API responses and requests
 */

/** Test status values */
export type TestStatus = 'idle' | 'starting' | 'running' | 'completed' | 'cancelled' | 'error';

/** Network interface information from /api/v1/interfaces */
export interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number;
  duplex: string;
  state: string;
  driver: string;
  physical: boolean;
  xdp: boolean;
  dpdk: boolean;
  score: number;
}

/** Runtime statistics from /api/v1/stats */
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
  errorMessage?: string;
}

/** Initial stats state */
export const initialStats: Stats = {
  packetsReceived: 0,
  packetsSent: 0,
  bytesReceived: 0,
  bytesSent: 0,
  currentPps: 0,
  currentMbps: 0,
  uptime: 0,
  testStatus: 'idle',
  currentTest: null,
};

/** Test result from completed test */
export interface TestResult {
  testType: string;
  module: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
  success?: boolean;
  error?: string;
  metrics?: Record<string, number | string>;
  data?: Record<string, unknown>;
}

/** License information */
export interface LicenseInfo {
  valid: boolean;
  tier: string;
  expiresAt?: string;
  daysRemaining?: number;
  features?: string[];
}

/** Auth response from login */
export interface AuthResponse {
  token: string;
  refreshToken: string;
  expiresIn: number;
}

/** Validate InterfaceInfo array response */
export function isValidInterfaceArray(data: unknown): data is InterfaceInfo[] {
  if (!Array.isArray(data)) {
    return false;
  }
  return data.every(
    (item) =>
      typeof item === 'object' &&
      item !== null &&
      typeof item.name === 'string' &&
      typeof item.mac === 'string' &&
      typeof item.speed === 'number',
  );
}

/** Validate Stats response */
export function isValidStats(data: unknown): data is Partial<Stats> {
  if (typeof data !== 'object' || data === null) {
    return false;
  }
  const obj = data as Record<string, unknown>;
  // At minimum, uptime should be a number
  return typeof obj.uptime === 'number' || typeof obj.packetsReceived === 'number';
}

/** Validate AuthResponse */
export function isValidAuthResponse(data: unknown): data is AuthResponse {
  if (typeof data !== 'object' || data === null) {
    return false;
  }
  const obj = data as Record<string, unknown>;
  return typeof obj.token === 'string';
}
