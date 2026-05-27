# Shared UI shell — canonical source contract

Stem owns the canonical implementation of the three-product shared shell.
Seed and niac sync these files into their own repos via
`scripts/sync-shell.sh`; downstream edits are overwritten on next sync.

## Canonical files (owned here)

| File | Purpose |
|---|---|
| `ui/src/ui/Sidebar.tsx` | Persistent collapsible left navigation with mobile drawer, gradient active state, tri-color badges, hover prefetch, gradient page background. |
| `ui/src/ui/PageHeader.tsx` | Page title bar with optional breadcrumbs, action area, and slide-out help panel. |

Each canonical file carries a header banner identifying it as such.

## Files NOT shared

| File | Why |
|---|---|
| `HeaderBar.tsx` | Too much per-product variance (stem has interface picker; seed has ethernet/wifi split + recommended-star + logo-color-as-status; niac has no real HeaderBar). Each repo owns its own; structural pattern is the same. |

## Token contract

The shell expects each consuming repo's `ui/src/index.css` to define the
following tokens. Values are per-product; names are universal.

### Required color tokens
- `brand-primary`, `brand-accent`, `brand-gold`
- `surface-base`, `surface-raised`, `surface-border`, `surface-hover`,
  `surface-sunken`, `surface-deep`
- `text-primary`, `text-secondary`, `text-muted`, `text-accent`,
  `text-inverse`, `text-disabled`
- `status-success`, `status-warning`, `status-error`, `status-info`
- `log-trace`, `log-debug`, `log-info`, `log-warn`, `log-error`, `log-fatal`
- `scrim` (constant black, opacity-controlled via `bg-scrim/N`)
- `knob` (constant white, for toggle thumbs / text on saturated brand bg)

### Required typography classes (from `@layer components`)
- `heading-1`, `heading-2`, `heading-3`, `heading-4`, `section-title`
- `body-large`, `body`, `body-small`, `caption`, `label`, `code`

### Required spacing classes
- `stack-xs`, `stack-sm`, `stack`, `stack-lg`, `stack-xl`
- `gap-tight`, `gap-compact`, `gap-default`, `gap-comfortable`, `gap-spacious`
- `pad-xs`, `pad-sm`, `pad`, `pad-lg`, `pad-xl`
- `mb-tight`, `mb-heading`, `mb-content`, `mb-section`, `mb-section-lg`
- `mt-tight`, `mt-inline`, `mt-heading`, `mt-content`, `mt-section`
- `flex-center`, `flex-between`

### Required utility files
- `ui/src/utils/prefetch.ts` — exports `prefetchRoute(path: string): void`.
  Implementation may be a stub; the Sidebar calls it on nav-item hover.
- `ui/src/utils/storage.ts` — exports `safeGetItem(k)` and `safeSetItem(k, v)`.
- `ui/src/constants/sizes.ts` — exports `iconSizes` object with `xs/sm/md/lg`
  string properties mapping to Tailwind classes.

### Required dependencies
- React 19+
- `react-router-dom` (for `useLocation`, `useNavigate`, `Link`)
- `lucide-react` (icon library)

## How seed/niac consume

```bash
# In seed or niac:
make sync-shell    # copies canonical files from ../stem, formats, writes lock
make verify-shell  # checks lock matches current files (run in CI)
```

The synced files in seed/niac carry a banner identifying their stem source
SHA. Editing them locally produces a CI failure on `verify-shell` until
the changes either (a) are pushed upstream to stem and re-synced, or
(b) are reverted.

## Adding a new canonical file

1. Land the new file in stem with the `// CANONICAL SHELL` banner.
2. Update the table in this doc.
3. Update `scripts/sync-shell.sh` in seed and niac to include it.
4. Run `make sync-shell` downstream; commit the synced copy.
