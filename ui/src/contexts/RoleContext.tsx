/**
 * RoleContext — global stem role (Reflector vs Test Master).
 *
 * A single stem instance runs in exactly one role at a time. The role
 * is selected on first boot (Setup Wizard) and can be switched at any
 * time via the header RoleChip. The selection persists to localStorage
 * under the key `stem-role` so it survives reload.
 *
 * setRole now wires through to the backend POST /api/v1/mode endpoint.
 * On success local state and localStorage are updated; on failure
 * local state is unchanged and an inline error is surfaced via
 * roleSwitchError so the chip can render a dismissable error tag. An
 * in-flight switch sets isSwitchingRole so the chip can show a
 * spinner.
 *
 * The signature of setRole stays `(role: StemRole) => void` rather
 * than returning a Promise: callers (RoleChip, RoleGuard, SetupWizard,
 * ReflectorPage) fire-and-forget the switch and react to state
 * changes, which avoids forcing every consumer to thread async.
 */
import {
  createContext,
  type FC,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';

export type StemRole = 'reflector' | 'test_master';

export const ROLE_STORAGE_KEY = 'stem-role';
export const DEFAULT_ROLE: StemRole = 'reflector';
export const ROLE_ENDPOINT = '/api/v1/mode';

function readPersistedRole(): StemRole {
  if (typeof window === 'undefined') {
    return DEFAULT_ROLE;
  }
  const raw = window.localStorage.getItem(ROLE_STORAGE_KEY);
  if (raw === 'reflector' || raw === 'test_master') {
    return raw;
  }
  return DEFAULT_ROLE;
}

/**
 * Shape of the backend POST /api/v1/mode response. Mirrors the Go
 * ModeUpdateResponse in internal/api/types.go. Only `mode` is required
 * for the frontend to update local state.
 */
interface ModeUpdateResponse {
  status: string;
  mode: StemRole;
  previous: StemRole;
}

/**
 * Backend error response shape (mirrors HTTPErrorResponse in
 * internal/api/errors.go). Both message and error are optional from
 * the caller's point of view because some fallback paths only set one.
 */
interface ApiErrorBody {
  message?: string;
  error?: string;
}

function isStemRole(value: unknown): value is StemRole {
  return value === 'reflector' || value === 'test_master';
}

function isModeUpdateResponse(value: unknown): value is ModeUpdateResponse {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  return (
    typeof candidate.status === 'string' &&
    isStemRole(candidate.mode) &&
    isStemRole(candidate.previous)
  );
}

async function extractErrorMessage(response: Response, fallback: string): Promise<string> {
  try {
    const body: unknown = await response.json();
    if (typeof body === 'object' && body !== null) {
      const candidate = body as ApiErrorBody;
      if (typeof candidate.message === 'string' && candidate.message.length > 0) {
        return candidate.message;
      }
      if (typeof candidate.error === 'string' && candidate.error.length > 0) {
        return candidate.error;
      }
    }
  } catch {
    // Body was not JSON — fall through to the fallback.
  }
  return fallback;
}

/**
 * Result of a single mode-switch attempt. The provider translates this
 * back into context state (`role`, `roleSwitchError`).
 */
type SwitchResult =
  | { kind: 'ok'; mode: StemRole }
  | { kind: 'error'; message: string }
  | { kind: 'cancelled' };

async function requestModeSwitch(next: StemRole): Promise<SwitchResult> {
  let response: Response;
  try {
    response = await fetch(ROLE_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ mode: next }),
    });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    return { kind: 'error', message: `Role switch failed: ${message}` };
  }

  if (!response.ok) {
    const message = await extractErrorMessage(
      response,
      `Role switch failed (HTTP ${response.status})`,
    );
    return { kind: 'error', message };
  }

  let body: unknown;
  try {
    body = await response.json();
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err);
    return { kind: 'error', message: `Role switch failed: ${message}` };
  }

  if (!isModeUpdateResponse(body)) {
    return { kind: 'error', message: 'Role switch failed: unexpected server response' };
  }

  // Trust the server's echoed mode rather than our local request —
  // handles the rare race where the server normalizes the value.
  return { kind: 'ok', mode: body.mode };
}

export interface RoleContextValue {
  role: StemRole;
  setRole: (role: StemRole) => void;
  isSwitchingRole: boolean;
  roleSwitchError: string | null;
  clearRoleSwitchError: () => void;
}

const RoleContext = createContext<RoleContextValue | null>(null);

interface RoleProviderProps {
  children: ReactNode;
}

export const RoleProvider: FC<RoleProviderProps> = ({ children }) => {
  const [role, setRoleState] = useState<StemRole>(() => readPersistedRole());
  const [isSwitchingRole, setIsSwitchingRole] = useState<boolean>(false);
  const [roleSwitchError, setRoleSwitchError] = useState<string | null>(null);

  // Track the latest in-flight switch so an older response cannot
  // overwrite a newer one if the user clicks quickly.
  const inflightToken = useRef<number>(0);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    window.localStorage.setItem(ROLE_STORAGE_KEY, role);
  }, [role]);

  const applySwitchResult = useCallback((result: SwitchResult, token: number): void => {
    if (token !== inflightToken.current) {
      return;
    }
    if (result.kind === 'ok') {
      setRoleState(result.mode);
      setRoleSwitchError(null);
    } else if (result.kind === 'error') {
      setRoleSwitchError(result.message);
    }
    setIsSwitchingRole(false);
  }, []);

  const setRole = useCallback(
    (next: StemRole): void => {
      // Clear any previous error so the spinner replaces it visually
      // and the new attempt has a clean slate.
      setRoleSwitchError(null);
      setIsSwitchingRole(true);
      inflightToken.current += 1;
      const token = inflightToken.current;
      // Fire-and-forget — state updates happen inside applySwitchResult.
      // Using .then/.catch (instead of `void promise`) satisfies biome's
      // noVoid + noFloatingPromises rules together.
      requestModeSwitch(next)
        .then((result) => {
          applySwitchResult(result, token);
        })
        .catch((err: unknown) => {
          const message = err instanceof Error ? err.message : String(err);
          applySwitchResult({ kind: 'error', message: `Role switch failed: ${message}` }, token);
        });
    },
    [applySwitchResult],
  );

  const clearRoleSwitchError = useCallback((): void => {
    setRoleSwitchError(null);
  }, []);

  const value = useMemo<RoleContextValue>(
    () => ({
      role,
      setRole,
      isSwitchingRole,
      roleSwitchError,
      clearRoleSwitchError,
    }),
    [role, setRole, isSwitchingRole, roleSwitchError, clearRoleSwitchError],
  );

  return <RoleContext.Provider value={value}>{children}</RoleContext.Provider>;
};

export function useRole(): RoleContextValue {
  const ctx = useContext(RoleContext);
  if (!ctx) {
    throw new Error('useRole must be used inside <RoleProvider>');
  }
  return ctx;
}
