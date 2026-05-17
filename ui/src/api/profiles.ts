// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * Profile API Client
 *
 * HTTP client for profile CRUD operations.
 * Follows Seed patterns for API integration.
 */

import type {
    CreateProfileRequest,
    Profile,
    ProfileConfig,
    ProfileExport,
    ProfileImportRequest,
    ProfileImportResult,
    ProfileListItem,
    UpdateProfileRequest,
} from '../types/profile';

const API_BASE = '/api/v1';

/**
 * API error with status code and message.
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/**
 * Fetch wrapper with error handling.
 */
async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
<<<<<<< Updated upstream
    const errorText = await response.text();
||||||| Stash base
    // biome-ignore lint/nursery/useAwaitThenable: Response.text() returns Promise<string>
    const errorText = await response.text();
=======
    const errorText = response.text();
>>>>>>> Stashed changes
    throw new ApiError(response.status, errorText || `HTTP ${response.status}`);
  }

  return response.json();
}

/**
 * Profile API endpoints.
 */
export const profileApi = {
  /**
   * List all profiles (minimal data for listings).
   */
  list(): Promise<ProfileListItem[]> {
    return fetchJson<ProfileListItem[]>(`${API_BASE}/profiles`);
  },

  /**
   * Get a single profile by ID.
   */
  get(id: string): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/${id}`);
  },

  /**
   * Get the currently active profile.
   */
  getActive(): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/active`);
  },

  /**
   * Get backend default settings.
   */
  getDefaults(): Promise<ProfileConfig> {
    return fetchJson<ProfileConfig>(`${API_BASE}/profiles/defaults`);
  },

  /**
   * Create a new profile.
   */
  create(request: CreateProfileRequest): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  },

  /**
   * Update an existing profile.
   */
  update(id: string, request: UpdateProfileRequest): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(request),
    });
  },

  /**
   * Delete a profile.
   */
  async delete(id: string): Promise<void> {
    await fetch(`${API_BASE}/profiles/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
  },

  /**
   * Switch to a different profile.
   */
  switchTo(id: string): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/${id}/activate`, {
      method: 'POST',
    });
  },

  /**
   * Set a profile as the default.
   */
  setDefault(id: string): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/${id}/default`, {
      method: 'POST',
    });
  },

  /**
   * Duplicate a profile.
   */
  duplicate(id: string, newName: string): Promise<Profile> {
    return fetchJson<Profile>(`${API_BASE}/profiles/${id}/duplicate`, {
      method: 'POST',
      body: JSON.stringify({ name: newName }),
    });
  },

  /**
   * Export all profiles.
   */
  exportAll(): Promise<ProfileExport> {
    return fetchJson<ProfileExport>(`${API_BASE}/profiles/export`);
  },

  /**
   * Import profiles.
   */
  import(request: ProfileImportRequest): Promise<ProfileImportResult> {
    return fetchJson<ProfileImportResult>(`${API_BASE}/profiles/import`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  },
};

export default profileApi;
