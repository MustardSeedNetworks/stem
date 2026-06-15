/**
 * Shell Store
 *
 * Zustand store for app-shell UI state — the modal/drawer and command-palette
 * open flags that previously lived as local `useState` in App.tsx (stage 1 of
 * the App.tsx decomposition). Deliberately NOT persisted: drawers and the
 * palette must never reopen across reloads. Connection/auth/test state are
 * extracted in later slices, not here.
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

interface ShellState {
  /** Settings drawer visibility. */
  settingsOpen: boolean;
  /** Help drawer visibility. */
  helpOpen: boolean;
  /** History drawer visibility. */
  historyOpen: boolean;
  /** Command palette (⌘K / Ctrl+K) visibility. */
  paletteOpen: boolean;
}

interface ShellActions {
  setSettingsOpen: (open: boolean) => void;
  setHelpOpen: (open: boolean) => void;
  setHistoryOpen: (open: boolean) => void;
  setPaletteOpen: (open: boolean) => void;
}

export type ShellStore = ShellState & ShellActions;

export const useShellStore = create<ShellStore>()(
  devtools(
    (set) => ({
      settingsOpen: false,
      helpOpen: false,
      historyOpen: false,
      paletteOpen: false,
      setSettingsOpen: (open) => set({ settingsOpen: open }, false, 'setSettingsOpen'),
      setHelpOpen: (open) => set({ helpOpen: open }, false, 'setHelpOpen'),
      setHistoryOpen: (open) => set({ historyOpen: open }, false, 'setHistoryOpen'),
      setPaletteOpen: (open) => set({ paletteOpen: open }, false, 'setPaletteOpen'),
    }),
    { name: 'shell-store' },
  ),
);
