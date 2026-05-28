# Shared UI shell — design conventions

The three products (seed, stem, niac) share the same shell *look and
behavior*, but **each repo owns its own implementation** — there is no
master repo and no sync script. When you change one of these files, keep
the others consistent by following the conventions here; ideally the
implementations stay near-identical, differing only where the product
genuinely differs (nav items, brand tokens, product content).

## Shared shell files (each repo owns its own copy)

| File | Purpose |
|---|---|
| `ui/src/ui/Sidebar.tsx` | Persistent collapsible left navigation with mobile drawer, gradient active state, tri-color badges, hover prefetch, gradient page background. |
| `ui/src/ui/PageHeader.tsx` | Page title bar with optional breadcrumbs, action area, and slide-out help panel. |

## Files NOT shared

| File | Why |
|---|---|
| `HeaderBar.tsx` | Too much per-product variance. Each repo owns its own; **the SHAPE is conventional** (see below). |

## HeaderBar shape convention (per-product)

Every product's `HeaderBar.tsx` follows the same three-slot layout, even
though slot fills differ:

```
┌─────────────────────────────────────────────────────────────┐
│ [logo]  [product name]  [connection-status]   …   [right] │
└─────────────────────────────────────────────────────────────┘
   ◄────── LEFT ──────►   ◄ CENTER ►    ◄──── RIGHT ────►
```

- **LEFT**: brand mark (logo) + product name + live connection / session
  status indicator. Always visible. The logo button doubles as a
  reconnect trigger when the connection is down.
- **CENTER**: empty today; reserved for breadcrumbs / page-level chrome.
- **RIGHT**: per-product context selectors followed by the theme toggle.
  Order: context first (left to right), theme toggle last.

### What belongs in the right slot (per product)
- **stem**: interface picker, profile dropdown (logout lives inside its
  menu), theme toggle.
- **seed**: ethernet picker, wifi picker (with recommended-interface
  marker), profile dropdown (logout inside), theme toggle.
- **niac**: theme toggle only (no profiles or interfaces in niac).

### What does NOT belong in the header (universal)
- **Settings** — lives in the sidebar footer.
- **Help** — lives in the sidebar footer.
- **Logout** — lives inside the profile dropdown menu (when a product
  has profiles); otherwise can live in sidebar footer or a user menu.
- **Refresh** — page-specific action; mount on the relevant page.
- **History** — sidebar nav item (or page-level drawer trigger);
  not a header concern.
- **Marketing taglines / company name** — header is for app chrome only.

### Banned vocabulary in headers
Same as the global product rules: no "AI", no "Premium" tier, no
deprecated tier names. See `CLAUDE.md` for the full list.

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

## Keeping the three repos consistent

There is no sync script and no master repo — each repo owns its copy of
these files outright. When you change a shared shell file in one product:

1. Apply the equivalent change to the other two repos (seed, stem, niac),
   keeping the implementations near-identical.
2. Preserve the conventions in this doc (HeaderBar shape, sidebar
   structure, theme-token usage).
3. Keep per-product differences only where they must exist — nav items,
   brand tokens (`index.css`), and product-specific content.

If you add a new shared shell file, document it in the table above and add
the equivalent file to all three repos.
