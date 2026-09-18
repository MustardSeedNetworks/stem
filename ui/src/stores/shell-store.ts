/**
 * Shell Store
 *
 * Zustand store for app-shell UI state — the modal/drawer and command-palette
 * open state that previously lived as local `useState` in App.tsx (stage 1 of
 * the App.tsx decomposition). Deliberately NOT persisted: drawers and the
 * palette must never reopen across reloads. Connection/auth/test state are
 * extracted in later slices, not here.
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface ShellState {
  /** Settings drawer visibility. */
  settingsOpen: boolean;
  /**
   * The route the help drawer was opened for, or null when it is closed.
   * A route rather than a flag so the drawer cannot be opened for one page
   * and rendered on another (#1305).
   */
  helpRoute: string | null;
  /** Command palette (⌘K / Ctrl+K) visibility. */
  paletteOpen: boolean;
}

interface ShellActions {
  setSettingsOpen: (open: boolean) => void;
  setHelpRoute: (route: string | null) => void;
  setPaletteOpen: (open: boolean) => void;
}

export type ShellStore = ShellState & ShellActions;

export const useShellStore = create<ShellStore>()(
  devtools(
    (set) => ({
      settingsOpen: false,
      helpRoute: null,
      paletteOpen: false,
      setSettingsOpen: (open) => set({ settingsOpen: open }, false, 'setSettingsOpen'),
      setHelpRoute: (route) => set({ helpRoute: route }, false, 'setHelpRoute'),
      setPaletteOpen: (open) => set({ paletteOpen: open }, false, 'setPaletteOpen'),
    }),
    { name: 'shell-store' },
  ),
);
