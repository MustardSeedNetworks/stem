/**
 * React Query Client Configuration
 *
 * Centralized QueryClient setup with sensible defaults for a
 * network-focused application. Mirrors seed's lib/queryClient.ts
 * (harmonization-by-convention; no shared package).
 *
 * Stem drives real-time data with short-interval polling (per-query
 * `refetchInterval`), so the global defaults stay conservative and let
 * individual queries opt into faster cadences.
 */

import { QueryClient } from '@tanstack/react-query';

/**
 * Create a QueryClient with application-specific defaults.
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // Data considered fresh for 30s; aggressive readers override via refetchInterval.
        staleTime: 30 * 1000,
        // Keep cache for 5 minutes after a query goes unused.
        gcTime: 5 * 60 * 1000,
        // Conservative retry — pollers refetch on their own interval anyway.
        retry: 1,
        // The app owns its own polling/refresh; don't double up on focus/reconnect.
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
      },
      mutations: {
        retry: 1,
      },
    },
  });
}

// Singleton QueryClient for the application.
let queryClient: QueryClient | null = null;

/**
 * Get or create the singleton QueryClient.
 * Use this to access the client outside of React components.
 */
export function getQueryClient(): QueryClient {
  if (!queryClient) {
    queryClient = createQueryClient();
  }
  return queryClient;
}
