// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * Profile Store
 *
 * Zustand store for managing test configuration profiles.
 * Follows Seed patterns for state management with React Query integration.
 *
 * Features:
 * - Profile CRUD operations
 * - Active profile management
 * - Backend defaults fallback
 * - Settings update by category
 * - Import/Export functionality
 */

import { create } from 'zustand';
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { profileApi } from '../api/profiles';
import type {
  CreateProfileRequest,
  DisplaySettings,
  GeneralSettings,
  InterfaceSettings,
  Profile,
  ProfileConfig,
  ProfileExport,
  ProfileImportRequest,
  ProfileImportResult,
  ProfileStoreActions,
  ProfileStoreState,
  TestConfigs,
  ThresholdSettings,
  UpdateProfileRequest,
} from '../types/profile';
import { DEFAULT_PROFILE_CONFIG } from '../types/profile';

/**
 * Combined store type.
 */
type ProfileStore = ProfileStoreState & ProfileStoreActions;

/**
 * Merge settings with defaults.
 * Priority: Profile settings → Backend defaults → Hardcoded defaults
 */
function mergeWithDefaults(
  config: ProfileConfig | undefined,
  backendDefaults: ProfileConfig | null,
): ProfileConfig {
  const hardcodedDefaults = DEFAULT_PROFILE_CONFIG;

  return {
    general: {
      ...(hardcodedDefaults.general ?? {}),
      ...(backendDefaults?.general ?? {}),
      ...(config?.general ?? {}),
    } as GeneralSettings,
    tests: {
      ...(hardcodedDefaults.tests ?? {}),
      ...(backendDefaults?.tests ?? {}),
      ...(config?.tests ?? {}),
    },
    interfaces: {
      ...(hardcodedDefaults.interfaces ?? {}),
      ...(backendDefaults?.interfaces ?? {}),
      ...(config?.interfaces ?? {}),
    },
    thresholds: {
      ...(hardcodedDefaults.thresholds ?? {}),
      ...(backendDefaults?.thresholds ?? {}),
      ...(config?.thresholds ?? {}),
    } as ThresholdSettings,
    display: {
      ...(hardcodedDefaults.display ?? {}),
      ...(backendDefaults?.display ?? {}),
      ...(config?.display ?? {}),
    } as DisplaySettings,
  };
}

/**
 * Profile store with Zustand.
 * Uses immer for immutable updates and subscribeWithSelector for optimized subscriptions.
 */
export const useProfileStore = create<ProfileStore>()(
  devtools(
    persist(
      subscribeWithSelector(
        immer((set, get) => ({
          // Initial state
          profiles: [],
          activeProfile: null,
          backendDefaults: null,
          isLoading: false,
          error: null,

          // Actions
          loadProfiles: async () => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const [profiles, backendDefaults] = await Promise.all([
                profileApi.list(),
                profileApi.getDefaults().catch(() => null),
              ]);

              set((state) => {
                state.profiles = profiles as unknown as Profile[];
                state.backendDefaults = backendDefaults;
                state.isLoading = false;
              });
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to load profiles';
                state.isLoading = false;
              });
            }
          },

          loadActiveProfile: async () => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const activeProfile = await profileApi.getActive();
              const { backendDefaults } = get();

              // Merge with defaults
              const mergedConfig = mergeWithDefaults(activeProfile.config, backendDefaults);

              set((state) => {
                state.activeProfile = { ...activeProfile, config: mergedConfig };
                state.isLoading = false;
              });
            } catch (error) {
              set((state) => {
                state.error =
                  error instanceof Error ? error.message : 'Failed to load active profile';
                state.isLoading = false;
              });
            }
          },

          switchProfile: async (profileId: string) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const profile = await profileApi.switchTo(profileId);
              const { backendDefaults } = get();

              // Merge with defaults
              const mergedConfig = mergeWithDefaults(profile.config, backendDefaults);

              set((state) => {
                state.activeProfile = { ...profile, config: mergedConfig };
                state.isLoading = false;
              });
              return true;
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to switch profile';
                state.isLoading = false;
              });
              return false;
            }
          },

          createProfile: async (request: CreateProfileRequest) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const profile = await profileApi.create(request);

              set((state) => {
                state.profiles.push(profile);
                state.isLoading = false;
              });
              return profile;
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to create profile';
                state.isLoading = false;
              });
              throw error;
            }
          },

          updateProfile: async (id: string, request: UpdateProfileRequest) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const profile = await profileApi.update(id, request);
              const { backendDefaults } = get();

              set((state) => {
                // Update in profiles list
                const index = state.profiles.findIndex((p) => p.id === id);
                if (index !== -1) {
                  state.profiles[index] = profile;
                }

                // Update active profile if it's the one being updated
                if (state.activeProfile?.id === id) {
                  const mergedConfig = mergeWithDefaults(profile.config, backendDefaults);
                  state.activeProfile = { ...profile, config: mergedConfig };
                }

                state.isLoading = false;
              });
              return profile;
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to update profile';
                state.isLoading = false;
              });
              throw error;
            }
          },

          deleteProfile: async (id: string) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              await profileApi.delete(id);

              set((state) => {
                state.profiles = state.profiles.filter((p) => p.id !== id);
                // Clear active profile if it was deleted
                if (state.activeProfile?.id === id) {
                  state.activeProfile = null;
                }
                state.isLoading = false;
              });
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to delete profile';
                state.isLoading = false;
              });
              throw error;
            }
          },

          duplicateProfile: async (id: string, newName: string) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const profile = await profileApi.duplicate(id, newName);

              set((state) => {
                state.profiles.push(profile);
                state.isLoading = false;
              });
              return profile;
            } catch (error) {
              set((state) => {
                state.error =
                  error instanceof Error ? error.message : 'Failed to duplicate profile';
                state.isLoading = false;
              });
              throw error;
            }
          },

          setDefaultProfile: async (id: string) => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              await profileApi.setDefault(id);

              set((state) => {
                // Update isDefault flag in profiles
                for (const profile of state.profiles) {
                  profile.isDefault = profile.id === id;
                }
                state.isLoading = false;
              });
            } catch (error) {
              set((state) => {
                state.error =
                  error instanceof Error ? error.message : 'Failed to set default profile';
                state.isLoading = false;
              });
              throw error;
            }
          },

          updateSettings: async <K extends keyof ProfileConfig>(
            category: K,
            settings: ProfileConfig[K],
          ) => {
            const { activeProfile } = get();
            if (!activeProfile) {
              set((state) => {
                state.error = 'No active profile';
              });
              return;
            }

            // Optimistic update with immer
            set((state) => {
              if (state.activeProfile) {
                const currentSettings = state.activeProfile.config[category] ?? {};
                state.activeProfile.config[category] = {
                  ...currentSettings,
                  ...settings,
                } as ProfileConfig[K];
              }
            });

            try {
              await profileApi.update(activeProfile.id, {
                config: { [category]: settings },
              });
            } catch (error) {
              // Revert on error
              set((state) => {
                state.activeProfile = activeProfile;
                state.error = error instanceof Error ? error.message : 'Failed to update settings';
              });
            }
          },

          exportProfiles: async (): Promise<ProfileExport> => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const data = await profileApi.exportAll();
              set((state) => {
                state.isLoading = false;
              });
              return data;
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to export profiles';
                state.isLoading = false;
              });
              throw error;
            }
          },

          importProfiles: async (data: ProfileImportRequest): Promise<ProfileImportResult> => {
            set((state) => {
              state.isLoading = true;
              state.error = null;
            });
            try {
              const result = await profileApi.import(data);

              // Reload profiles after import
              await get().loadProfiles();

              set((state) => {
                state.isLoading = false;
              });
              return result;
            } catch (error) {
              set((state) => {
                state.error = error instanceof Error ? error.message : 'Failed to import profiles';
                state.isLoading = false;
              });
              throw error;
            }
          },
        })),
      ),
      {
        name: 'stem-profile-store',
        // Only persist the active profile ID, not the full data
        partialize: (state) => ({
          activeProfileId: state.activeProfile?.id,
        }),
      },
    ),
    { name: 'ProfileStore' },
  ),
);

// ============================================================================
// Derived Selectors (memoized automatically by Zustand)
// ============================================================================

/**
 * Helper to merge settings with defaults.
 * Priority: Profile value → Backend default → Hardcoded default
 */
function mergeSettingsWithDefaults<T>(
  profileValue: T | undefined,
  backendDefault: T | undefined,
  hardcodedDefault: T,
): T {
  return profileValue ?? backendDefault ?? hardcodedDefault;
}

/**
 * Selector for getting effective settings (merged with defaults).
 */
export function useEffectiveSettings(): ProfileConfig {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeWithDefaults(activeProfile?.config, backendDefaults);
}

/**
 * Selector for specific settings category.
 */
export function useSettingsCategory<K extends keyof ProfileConfig>(category: K): ProfileConfig[K] {
  const settings = useEffectiveSettings();
  return settings[category];
}

// ============================================================================
// Typed Settings Selectors (like Seed's pattern)
// ============================================================================

/**
 * Get general settings with defaults applied.
 */
export function useGeneralSettings(): GeneralSettings {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeSettingsWithDefaults(
    activeProfile?.config?.general,
    backendDefaults?.general,
    DEFAULT_PROFILE_CONFIG.general as GeneralSettings,
  );
}

/**
 * Get tests settings with defaults applied.
 */
export function useTestsSettings(): TestConfigs {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeSettingsWithDefaults(
    activeProfile?.config?.tests,
    backendDefaults?.tests,
    DEFAULT_PROFILE_CONFIG.tests ?? {},
  );
}

/**
 * Get interfaces settings with defaults applied.
 */
export function useInterfacesSettings(): InterfaceSettings {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeSettingsWithDefaults(
    activeProfile?.config?.interfaces,
    backendDefaults?.interfaces,
    DEFAULT_PROFILE_CONFIG.interfaces ?? {},
  );
}

/**
 * Get thresholds settings with defaults applied.
 */
export function useThresholdsSettings(): ThresholdSettings {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeSettingsWithDefaults(
    activeProfile?.config?.thresholds,
    backendDefaults?.thresholds,
    DEFAULT_PROFILE_CONFIG.thresholds as ThresholdSettings,
  );
}

/**
 * Get display settings with defaults applied.
 */
export function useDisplaySettings(): DisplaySettings {
  const activeProfile = useProfileStore((s) => s.activeProfile);
  const backendDefaults = useProfileStore((s) => s.backendDefaults);
  return mergeSettingsWithDefaults(
    activeProfile?.config?.display,
    backendDefaults?.display,
    DEFAULT_PROFILE_CONFIG.display as DisplaySettings,
  );
}

export default useProfileStore;
