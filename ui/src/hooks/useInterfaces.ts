/**
 * @fileoverview Interfaces Hook
 * @description Manages network interface discovery and selection.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { useCallback, useEffect, useState } from 'react';
import { type InterfaceInfo, isValidInterfaceArray } from '../types/api';

/** Select the best interface by score */
function selectBestInterface(data: InterfaceInfo[]): string {
  const best = data.reduce((a: InterfaceInfo, b: InterfaceInfo) => (a.score > b.score ? a : b));
  return best.name;
}

/** Check if error is unauthorized */
function isUnauthorizedError(error: unknown): boolean {
  return error instanceof Error && error.message === 'Unauthorized';
}

interface UseInterfacesOptions {
  /** Authenticated fetch function */
  authFetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>;
  /** Whether authenticated */
  isAuthenticated: boolean;
  /** Set connected state */
  setConnected: (connected: boolean) => void;
}

interface UseInterfacesResult {
  /** List of available network interfaces */
  interfaces: InterfaceInfo[];
  /** Currently selected interface name */
  selectedInterface: string;
  /** Update selected interface */
  setSelectedInterface: (name: string) => void;
  /** Refresh interface list */
  fetchInterfaces: () => Promise<void>;
  /** Get the currently selected interface object */
  selectedInterfaceInfo: InterfaceInfo | undefined;
}

/**
 * Hook for managing network interface discovery and selection.
 * Automatically fetches interfaces on mount and auto-selects the best one.
 */
export function useInterfaces({
  authFetch,
  isAuthenticated,
  setConnected,
}: UseInterfacesOptions): UseInterfacesResult {
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([]);
  const [selectedInterface, setSelectedInterface] = useState<string>('');

  // Process and store interface data
  const processInterfaceData = useCallback(
    (data: InterfaceInfo[]): void => {
      setInterfaces(data);
      if (data.length > 0) {
        setSelectedInterface((prev) => prev || selectBestInterface(data));
      }
      setConnected(true);
    },
    [setConnected],
  );

  // Fetch available interfaces
  const fetchInterfaces = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }
    try {
      const response = await authFetch('/api/v1/interfaces');
      if (!response.ok) {
        throw new Error('Failed to load interfaces');
      }
      const data: unknown = await (response.json() as Promise<unknown>);
      if (!isValidInterfaceArray(data)) {
        throw new Error('Invalid interface data received from server');
      }
      processInterfaceData(data);
    } catch (error) {
      if (!isUnauthorizedError(error)) {
        setConnected(false);
      }
    }
  }, [authFetch, isAuthenticated, processInterfaceData, setConnected]);

  // Fetch interfaces on mount
  useEffect(() => {
    fetchInterfaces().catch(() => {
      // Errors already handled inside fetchInterfaces
    });
  }, [fetchInterfaces]);

  // Get selected interface object
  const selectedInterfaceInfo = interfaces.find((i) => i.name === selectedInterface);

  return {
    interfaces,
    selectedInterface,
    setSelectedInterface,
    fetchInterfaces,
    selectedInterfaceInfo,
  };
}
