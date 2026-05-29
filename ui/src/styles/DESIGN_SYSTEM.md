# The Stem Design System

This design system keeps styling consistent and theme-aware. Use the centralized
tokens and component classes instead of scattered raw values.

## Token Architecture (read first)

Three tiers, **one** source of truth for values, **one** derivation direction:

```
Primitive   Tailwind's built-in palette (blue-700 = #1976d2)     ← never referenced directly in app code
   ↓ alias
Semantic    index.css @theme + :root/.dark                       ← THE source of truth for VALUES
            brand-*, status-*, surface-*, text-*, module-*,
            log-*, scrim, knob, z-overlay/z-max
   ↓ alias
Component   index.css @layer components (.btn-*, .card, .badge,  ← consume semantic tokens
            .alert, .table, .module-badge …) + the TS class-token
            objects in styles/ (status.*, layout.* …)
```

**Two invariants (enforced by `scripts/check-token-discipline.sh`):**

1. **Values flow one direction** — defined once in `index.css`, everything else
   references them. Never hand-copy a hex sideways into a `.ts`/`.tsx` file.
2. **App code names only semantic / component tokens** — never a primitive
   palette utility (`bg-blue-500`) and never a raw hex.

**Charts / SVG / inline styles:** use the CSS variables directly —
`fill="var(--color-module-benchmark)"`, `style={{ color: 'var(--color-module-reflector)' }}`.
They resolve through the cascade and flip light↔dark automatically. Stem has no
`<canvas>` drawing, so (unlike seed) it needs no JS token-reader; if a `<canvas>`
visualization is ever added, read the vars via `getComputedStyle` in a `tokens.ts`
helper, since a raw canvas can't resolve CSS variables.

**Brand:** Stem's anchor is **blue** `#1976d2` (stem-500) — the universal
test/measurement color (Spirent, Viavi, Keysight, Fluke). The six test modules
have their own accents (`--color-module-{reflector,benchmark,servicetest,trafficgen,measure,certify}`).
