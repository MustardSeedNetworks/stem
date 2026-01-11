// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

import { twMerge } from 'tailwind-merge';

/**
 * =============================================================================
 * THE STEM DESIGN SYSTEM - Mustard Seed Networks
 * =============================================================================
 *
 * Centralized design tokens and utilities for consistent UI across the app.
 *
 * ARCHITECTURE:
 * 1. CSS Variables (index.css) - Core color tokens for light/dark modes
 * 2. This file (theme.ts) - TypeScript tokens and utility functions
 * 3. Tailwind Classes - CSS-first configuration using @theme directive
 *
 * BRAND COLORS:
 * - Primary: Stem Green (#2d7a3e / #81c784 dark) - Actions, links, focus states
 * - Accent: Lighter Stem Green (#4caf50 / #a5d6a7 dark) - Hover states
 * - Gold: Mustard Gold (#d4a017 / #fbbf24 dark) - Special highlights
 *
 * STATUS COLORS (Industry Standard - DO NOT CHANGE):
 * - Success: Green (#28a745) - Positive states
 * - Warning: Amber (#ffc107) - Caution states
 * - Error: Red (#dc3545) - Error/danger states
 * - Info: Cyan (#17a2b8) - Informational states
 *
 * MODULE COLORS (The Stem specific):
 * - Reflector: Cyan - Loopback/Echo
 * - Benchmark: Red - RFC 2544
 * - ServiceTest: Orange - Y.1564/MEF
 * - TrafficGen: Yellow - Traffic Generation
 * - Measure: Blue - Y.1731 OAM
 * - Certify: Green - RFC 2889/6349/TSN
 *
 * USAGE:
 * import { spacing, button, cn, moduleColor } from '../styles/theme';
 * <button className={cn(button.base, button.variant.primary)}>Action</button>
 *
 * =============================================================================
 */

// ============================================================================
// SPACING SCALE
// ============================================================================

/**
 * Spacing scale - based on 4px grid
 * Use these semantic spacing utilities for consistency.
 */
export const spacing = {
  // Semantic CSS utility classes
  stack: {
    xs: 'stack-xs', // 4px vertical
    sm: 'stack-sm', // 8px vertical
    default: 'stack', // 12px vertical
    lg: 'stack-lg', // 16px vertical
    xl: 'stack-xl', // 24px vertical
  },

  section: {
    default: 'section-gap', // 24px between sections
  },

  gap: {
    tight: 'gap-tight', // 4px
    compact: 'gap-compact', // 8px
    default: 'gap-default', // 12px
    comfortable: 'gap-comfortable', // 16px
    spacious: 'gap-spacious', // 24px
  },

  pad: {
    xs: 'pad-xs', // 8px
    sm: 'pad-sm', // 12px
    default: 'pad', // 16px
    lg: 'pad-lg', // 24px
    xl: 'pad-xl', // 32px
  },

  // Chip/pill padding
  chip: {
    sm: 'px-3 py-1',
    md: 'px-3 py-1.5',
    lg: 'px-3 py-2',
  },

  // Tab button padding
  tab: 'py-2.5 px-3',

  inline: {
    xs: 'inline-gap-xs', // 4px
    sm: 'inline-gap-sm', // 6px
    default: 'inline-gap', // 8px
    lg: 'inline-gap-lg', // 12px
  },

  margin: {
    bottom: {
      section: 'mb-section', // 24px
      sectionLg: 'mb-section-lg', // 32px
      heading: 'mb-heading', // 12px
      content: 'mb-content', // 16px
      inline: 'mb-2', // 8px
      tight: 'mb-tight', // 4px
    },
    top: {
      section: 'mt-section', // 32px
      content: 'mt-content', // 16px
      heading: 'mt-heading', // 12px
      inline: 'mt-inline', // 8px
      tight: 'mt-tight', // 4px
    },
    left: {
      tight: 'ml-tight', // 4px
      inline: 'ml-inline', // 8px
      content: 'ml-content', // 16px
      spacious: 'ml-spacious', // 24px
    },
  },

  padding: {
    top: {
      heading: 'pt-heading', // 12px
      section: 'pt-section', // 16px
      tight: 'pt-tight', // 4px
    },
    bottom: {
      inline: 'pb-inline', // 8px
      tight: 'pb-tight', // 4px
    },
    right: {
      icon: 'pr-icon', // 40px
      tight: 'pr-tight', // 32px
    },
  },

  centered: 'py-centered', // 48px vertical

  badge: {
    xs: 'p-badge-xs', // 2px
    sm: 'p-badge-sm', // 4px
    padXs: 'badge-pad-xs', // px-2 py-0.5
  },

  compact: {
    py: 'py-compact', // 4px
    pyMd: 'py-compact-md', // 6px
  },

  row: {
    py: 'py-row', // 8px
    pyLg: 'py-row-lg', // 12px
  },

  iconBtn: {
    sm: 'p-icon-btn', // 4px
    md: 'p-icon-btn-md', // 6px
  },

  mainPadding: {
    y: 'main-padding-y', // py-4 sm:py-6
    x: 'content-padding-x', // px-4 sm:px-6 lg:px-8
  },

  drawerPad: 'drawer-content-pad', // px-4 sm:px-5 pb-10 pt-4

  // Indentation for nested content
  indent: 'pl-6', // 24px left indent for collapsible content
} as const;

// ============================================================================
// TYPOGRAPHY
// ============================================================================

export const typography = {
  heading: {
    h1: 'heading-1',
    h2: 'heading-2',
    h3: 'heading-3',
    h4: 'heading-4',
    section: 'section-title',
  },

  body: {
    large: 'body-large',
    default: 'body',
    small: 'body-small',
    caption: 'caption',
  },

  label: 'label',
  code: 'code',

  size: {
    xs: 'text-xs',
    sm: 'text-sm',
    base: 'text-base',
    lg: 'text-lg',
    xl: 'text-xl',
  },

  weight: {
    normal: 'font-normal',
    medium: 'font-medium',
    semibold: 'font-semibold',
    bold: 'font-bold',
  },
} as const;

// ============================================================================
// COMPONENT VARIANTS
// ============================================================================

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
    primary: 'bg-brand-primary text-text-inverse hover:bg-brand-accent',
    secondary: 'border border-surface-border bg-surface-raised hover:bg-surface-hover',
    ghost: 'hover:bg-surface-hover',
    danger: 'bg-status-error text-text-inverse hover:opacity-90',
    success: 'bg-status-success text-text-inverse hover:opacity-90',
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

/**
 * Status indicator variants - for connection status, health, etc.
 */
export const status = {
  dot: 'inline-block w-2 h-2 rounded-full',

  color: {
    success: 'bg-status-success',
    warning: 'bg-status-warning',
    error: 'bg-status-error',
    info: 'bg-status-info',
    inactive: 'bg-surface-border',
  },

  withLabel: 'inline-flex items-center gap-2',
} as const;

/**
 * Sizing tokens - for consistent heights/widths
 */
export const sizing = {
  height: {
    modal: 'max-h-modal', // 85vh
    drawer: 'h-full',
    panel: 'max-h-[70vh]',
  },

  width: {
    drawer: 'w-80', // 320px
    drawerWide: 'w-96', // 384px
    panel: 'w-72', // 288px
    dropdown: 'w-64', // 256px
  },

  minHeight: {
    card: 'min-h-[120px]',
    section: 'min-h-[200px]',
  },
} as const;

// ============================================================================
// ICON SIZING
// ============================================================================

export const icon = {
  size: {
    xs: 'w-3 h-3',
    sm: 'w-4 h-4',
    md: 'w-5 h-5',
    lg: 'w-6 h-6',
    xl: 'w-8 h-8',
  },

  inline: 'inline-flex items-center gap-1.5',
  button: 'inline-flex items-center gap-2',
  leading: 'flex items-center gap-2',
} as const;

// ============================================================================
// BORDER & RADIUS
// ============================================================================

export const radius = {
  none: 'rounded-none',
  sm: 'rounded-sm',
  default: 'rounded',
  md: 'rounded-md',
  lg: 'rounded-lg',
  xl: 'rounded-xl',
  full: 'rounded-full',
} as const;

/**
 * Border tokens - consistent border styling
 */
export const border = {
  width: {
    none: 'border-0',
    default: 'border',
    thick: 'border-2',
  },

  color: {
    default: 'border-surface-border',
    focus: 'border-brand-primary',
    error: 'border-status-error',
    success: 'border-status-success',
    warning: 'border-status-warning',
  },

  card: 'border border-surface-border',
  input: 'border border-surface-border focus:border-brand-primary',
  divider: 'border-t border-surface-border',
} as const;

// ============================================================================
// LAYOUT PATTERNS
// ============================================================================

export const layout = {
  flex: {
    center: 'flex items-center justify-center',
    between: 'flex items-center justify-between',
    start: 'flex items-center justify-start',
    end: 'flex items-center justify-end',
    col: 'flex flex-col',
    colCenter: 'flex flex-col items-center justify-center',
    wrap: 'flex flex-wrap',
  },

  grid: {
    cards: 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6',
    form2col: 'grid grid-cols-2 gap-2',
    data2col: 'grid grid-cols-2 gap-x-4 gap-y-2',
  },

  inline: {
    tight: 'flex items-center gap-1',
    default: 'flex items-center gap-2',
    comfortable: 'flex items-center gap-3',
    spacious: 'flex items-center gap-4',
    wrap: 'flex flex-wrap items-center gap-2',
  },

  stack: {
    tight: 'flex flex-col gap-1',
    default: 'flex flex-col gap-2',
    comfortable: 'flex flex-col gap-3',
    spacious: 'flex flex-col gap-4',
  },
} as const;

// ============================================================================
// MODULE COLORS - The Stem specific
// ============================================================================

/**
 * Module colors - accent colors for The Stem's test modules
 *
 * IMPORTANT: Use these for icons and small badges only, NOT for card backgrounds.
 * Cards should remain consistent (surface-raised) across all modules.
 */
export const moduleColor = {
  reflector: {
    icon: 'text-module-reflector',
    badge: 'bg-module-reflector/20 text-module-reflector',
    border: 'border-module-reflector/30',
  },
  benchmark: {
    icon: 'text-module-benchmark',
    badge: 'bg-module-benchmark/20 text-module-benchmark',
    border: 'border-module-benchmark/30',
  },
  servicetest: {
    icon: 'text-module-servicetest',
    badge: 'bg-module-servicetest/20 text-module-servicetest',
    border: 'border-module-servicetest/30',
  },
  trafficgen: {
    icon: 'text-module-trafficgen',
    badge: 'bg-module-trafficgen/20 text-module-trafficgen',
    border: 'border-module-trafficgen/30',
  },
  measure: {
    icon: 'text-module-measure',
    badge: 'bg-module-measure/20 text-module-measure',
    border: 'border-module-measure/30',
  },
  certify: {
    icon: 'text-module-certify',
    badge: 'bg-module-certify/20 text-module-certify',
    border: 'border-module-certify/30',
  },
} as const;

/**
 * Brand colors - for special brand elements
 */
export const brand = {
  gold: {
    text: 'text-brand-gold',
    bg: 'bg-brand-gold',
    badge: 'bg-brand-gold/20 text-brand-gold',
    border: 'border-brand-gold/30',
  },
} as const;

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

/**
 * Combine class names with Tailwind class conflict resolution.
 */
export function cn(...classes: (string | boolean | undefined | null)[]): string {
  return twMerge(classes.filter(Boolean).join(' '));
}

/**
 * Build a button class string
 */
export function buttonClass(
  variant: keyof typeof button.variant = 'primary',
  size: keyof typeof button.size = 'md',
  className?: string,
): string {
  return cn(button.base, button.variant[variant], button.size[size], className);
}

/**
 * Build a card class string
 */
export function cardClass(
  variant: keyof typeof card.variant = 'default',
  padding: keyof typeof card.padding = 'md',
  className?: string,
): string {
  return cn(card.base, card.variant[variant], card.padding[padding], className);
}

/**
 * Build a badge class string
 */
export function badgeClass(
  variant: keyof typeof badge.variant = 'default',
  className?: string,
): string {
  return cn(badge.base, badge.variant[variant], className);
}

/**
 * Build an input class string
 */
export function inputClass(
  state: keyof typeof input.state = 'default',
  size: keyof typeof input.size = 'md',
  className?: string,
): string {
  return cn(input.base, input.state[state], input.size[size], className);
}

/**
 * Build a modal class string
 */
export function modalClass(
  size: keyof typeof modal.size = 'md',
  padding: keyof typeof modal.padding = 'md',
  className?: string,
): string {
  return cn(modal.content, modal.size[size], modal.padding[padding], className);
}
