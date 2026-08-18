/**
 * Test Setup and Utilities
 *
 * Purpose: Shared test configuration and mock utilities for Vitest.
 * Provides mock implementations of browser APIs (localStorage, fetch, etc.)
 * and common test helpers used across the test suite.
 *
 * Dependencies: vitest, @testing-library/jest-dom
 * Applied In: All test files via vitest configuration
 */

import '@testing-library/jest-dom';
import { afterEach, beforeEach, vi } from 'vitest';

// ============================================================
// Real i18n
// ============================================================
// Initialising the real i18next (rather than mocking react-i18next with a
// fixed table) means a test that asserts on user-visible text is asserting
// on the actual locale files. The mock this replaces returned the key
// itself for anything not in its ~30-entry table and ignored t()'s options
// argument entirely, so defaultValue and interpolation silently vanished.
import '../i18n';

// ============================================================
// JSDoM polyfills — common browser APIs not implemented by JSDoM
// Universal baseline shared across seed/stem/niac test setups.
// ============================================================

// matchMedia: dark-mode detection, responsive hooks, prefers-reduced-motion
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // legacy API still used by some libs
    removeListener: vi.fn(), // legacy
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }),
});

// ResizeObserver: used by xyflow, codemirror, recharts, headlessui
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
})) as unknown as typeof ResizeObserver;

// IntersectionObserver: used by lazy loading, infinite scroll
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
  takeRecords: vi.fn(() => []),
  root: null,
  rootMargin: '',
  thresholds: [],
})) as unknown as typeof IntersectionObserver;

// EventSource: required by useSse (#302) SSE subscription hook. JSDoM does not
// implement EventSource. Without this polyfill, any component test that mounts
// a tree containing useSse() throws "ReferenceError: EventSource is not defined".
// Tests that need to assert on SSE behavior should mock this further per-suite.
const eventSourceMock = vi.fn().mockImplementation(() => ({
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  close: vi.fn(),
  dispatchEvent: vi.fn(() => true),
  url: '/api/v1/events',
  readyState: 1, // OPEN
  withCredentials: false,
  onerror: null,
  onmessage: null,
  onopen: null,
  CONNECTING: 0,
  OPEN: 1,
  CLOSED: 2,
}));
Object.assign(eventSourceMock, { CONNECTING: 0, OPEN: 1, CLOSED: 2 });
global.EventSource = eventSourceMock as unknown as typeof EventSource;

// ============================================================
// Mock localStorage
// ============================================================
export interface MockLocalStorage {
  getItem: ReturnType<typeof vi.fn>;
  setItem: ReturnType<typeof vi.fn>;
  removeItem: ReturnType<typeof vi.fn>;
  clear: () => void;
  _store: Record<string, string>;
}

export function createMockLocalStorage(): MockLocalStorage {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: () => {
      store = {};
    },
    get _store(): Record<string, string> {
      return store;
    },
  };
}

const mockLocalStorage: MockLocalStorage = createMockLocalStorage();
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

// Export for use in tests
export { mockLocalStorage };

// ============================================================
// Mock fetch
// ============================================================
export const mockFetch: ReturnType<typeof vi.fn> = vi.fn();
global.fetch = mockFetch;

// Helper to create standard API responses
export function createMockResponse<T>(
  data: T,
  ok = true,
  status = 200,
): Promise<{
  ok: boolean;
  status: number;
  json: () => Promise<T>;
  text: () => Promise<string>;
  headers: Headers;
}> {
  return Promise.resolve({
    ok,
    status,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
    headers: new Headers(),
  });
}

// Helper to create error responses
export function createMockErrorResponse(status = 500, message = 'Error'): Promise<Response> {
  return Promise.resolve({
    ok: false,
    status,
    json: () => Promise.resolve({ error: message }),
    text: () => Promise.resolve(message),
    headers: new Headers(),
  } as unknown as Response);
}

// ============================================================
// Mock window.location
// ============================================================
export function mockWindowLocation(overrides: Partial<Location> = {}): void {
  const defaultLocation = {
    protocol: 'https:',
    host: 'localhost:8444',
    hostname: 'localhost',
    port: '8444',
    pathname: '/',
    search: '',
    hash: '',
    href: 'https://localhost:8444/',
    origin: 'https://localhost:8444',
    ...overrides,
  };

  Object.defineProperty(window, 'location', {
    value: defaultLocation,
    writable: true,
  });
}

// ============================================================
// Test lifecycle hooks
// ============================================================
beforeEach(() => {
  vi.clearAllMocks();
  mockLocalStorage.clear();
  mockFetch.mockReset();
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ============================================================
// Common test data factories
// ============================================================

// Auth token factory
export function createMockAuthToken(expiresInSeconds = 3600): {
  token: string;
  expiry: number;
} {
  return {
    token: `test-token-${Date.now()}`,
    expiry: Math.floor(Date.now() / 1000) + expiresInSeconds,
  };
}
