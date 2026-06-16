# State management — what goes where

The Stem UI uses four state mechanisms. Picking the wrong one is the most common
source of re-render bugs, stale data, and "why are there two sources of truth"
confusion. Use this decision order; when two fit, prefer the one higher in the list.

## Decision order

1. **Is it server data** (fetched from `/api/**`, owned by the backend)?
   → **React Query** (`@tanstack/react-query`). It owns caching, dedup, background
   refetch, and staleness. Never copy server data into Zustand/Context "to keep it
   handy" — that creates a second source of truth that goes stale.
   - Singleton + provider: `lib/queryClient` (`getQueryClient`) wired via
     `QueryClientProvider` in `main.tsx`.
   - Examples: the `['interfaces']` and `['stats']` reads in `useTestExecution`; all
     server reads go through the module-level `authFetch` (see `stores/auth-store`).

2. **Is it ephemeral global app/UI state** that many unrelated components read or
   drive, but is NOT server-owned and need NOT persist across reloads?
   → **Zustand store** (`stores/*`). Atomic updates, no provider nesting, no
   re-render cascade, trivially unit-testable.
   - `stores/shell-store` — drawers (settings/help/history) + command palette open
     state. The reference example for "ephemeral global UI".
   - `stores/test-store` — test-execution config + run state (selected tests, the
     per-RFC config objects, isStarting/isStopping). Setters are useState-compatible.
   - `stores/auth-store` — authentication state + flows (login/MFA/logout/setup/
     recovery) and the module-level `authFetch` primitive. Security-sensitive; the
     localStorage auth flag is the only persisted bit. NOT a React Query cache.
   - `stores/profile-store` — test-configuration profiles (CRUD, active profile,
     backend-defaults fallback); persists via `zustand/persist`.

3. **Is it a cross-cutting capability** scoped to the app — identity/role, or a
   small derived bundle handed to a subtree?
   → **React Context**. Use for stable, low-frequency values; avoid high-churn data
   in Context (no selector → every consumer re-renders).
   - `contexts/RoleContext` — the Stem instance role (test_master / reflector),
     persisted; drives the header RoleChip + per-page RoleGuard.
   - `contexts/AppContext` — the per-render bundle handed to the routed test pages
     (configs, interfaces, stats, reflector start/stop). Assembled in `App.tsx`.
   - `contexts/ModuleSettingsContext` — per-module settings consumed by test pages.

4. **Is it local to one component / subtree** (open/closed, hovered, form draft)?
   → **`useState` / `useReducer`**. Default for anything not shared. Lift only when
   a second component genuinely needs it.
   - Examples that stay local in `useTestExecution`: `connected` (reconciled off the
     interfaces query + auth flag), `testResult`, `selectedInterface`.

## Anti-patterns

- **Duplicating server data** into Zustand/Context. React Query is the cache.
- **Global store for one component's flag.** Keep it local.
- **High-churn values in Context.** Context has no selector; every consumer
  re-renders on any change. Use Zustand (selectors) or local state.
- **A new Context provider per feature.** Prefer a Zustand store unless the value is
  genuinely a tree-scoped capability (role).
- **Copying tokens into JS.** Auth tokens live in httpOnly cookies; the store tracks
  only a boolean flag. Never read/store tokens client-side.
