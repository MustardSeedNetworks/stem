/**
 * themeComponents.ts — button, input, card, badge, alert, modal variants.
 * Re-exported through theme.ts.
 */

/**
 * Input variants - consistent form input styling
 */
export const input = {
  base: 'w-full rounded border bg-surface-raised text-text-primary transition-colors focus:outline-none focus:ring-2 focus:ring-brand-primary disabled:opacity-50 disabled:cursor-not-allowed',

  state: {
    default: 'border-surface-border',
    error: 'border-status-error',
    success: 'border-status-success',
  },

  size: {
    sm: 'px-2 py-1.5 text-sm',
    md: 'px-2.5 py-2 text-sm',
    lg: 'px-3 py-2.5 text-base',
  },
} as const;

/**
 * Button variants
 */
export const button = {
  base: 'inline-flex items-center justify-center gap-2 rounded font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary disabled:opacity-50 disabled:cursor-not-allowed',

  variant: {
    // Stem anchor is stem-500 (#1976d2) — needs WHITE text. text-text-inverse
    // flips to dark in dark mode and fails AA against the constant blue
    // anchor. Opacity hover avoids the lighten-to-accent trap (stem-300
    // fails contrast against white text). Per Phase 7 of the 2026-05-22 audit.
    primary: 'bg-brand-primary text-on-brand hover:bg-brand-primary/90',
    secondary: 'border border-surface-border bg-surface-raised hover:bg-surface-hover',
    ghost: 'hover:bg-surface-hover',
    // Status danger/success buttons: align with brand-primary fix (constant
    // foreground rather than text-inverse mode-flip).
    danger: 'bg-status-error text-on-danger hover:bg-status-error/90',
    success: 'bg-status-success text-zinc-900 hover:bg-status-success/90',
  },

  size: {
    xs: 'px-2 py-1 text-xs',
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg',
  },
} as const;

/**
 * Card variants
 */
export const card = {
  base: 'rounded-lg border bg-surface-raised',

  variant: {
    default: 'border-surface-border',
    elevated: 'border-surface-border shadow-lg',
    interactive:
      'border-surface-border hover:border-brand-primary cursor-pointer transition-colors',
  },

  padding: {
    none: '',
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6',
  },
} as const;

/**
 * Badge variants
 */
export const badge = {
  base: 'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium',

  variant: {
    default: 'bg-surface-hover text-text-primary',
    success: 'bg-status-success/10 text-status-success',
    warning: 'bg-status-warning/10 text-status-warning',
    error: 'bg-status-error/10 text-status-error',
    info: 'bg-status-info/10 text-status-info',
    primary: 'bg-brand-primary/10 text-brand-primary',
  },
} as const;

/**
 * Alert/Banner variants
 */
export const alert = {
  base: 'px-4 py-3 rounded-lg border',

  variant: {
    error: 'bg-status-error/10 border-status-error/20 text-status-error',
    warning: 'bg-status-warning/10 border-status-warning/20 text-status-warning',
    success: 'bg-status-success/10 border-status-success/20 text-status-success',
    info: 'bg-status-info/10 border-status-info/20 text-status-info',
  },
} as const;

/**
 * Modal/Dialog variants
 */
export const modal = {
  overlay: 'fixed inset-0 z-50 flex items-center justify-center p-4',
  backdrop: 'absolute inset-0 bg-black/50 backdrop-blur-sm',
  content:
    'bg-surface-raised border border-surface-border rounded-lg shadow-xl max-h-[85vh] overflow-y-auto',

  size: {
    sm: 'max-w-md w-full',
    md: 'max-w-2xl w-full',
    lg: 'max-w-4xl w-full',
    xl: 'max-w-6xl w-full',
    full: 'max-w-7xl w-full',
  },

  padding: {
    sm: 'pad',
    md: 'pad-lg',
    lg: 'pad-xl',
  },
} as const;
