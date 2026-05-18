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
 * 1. CSS Variables (index.css) - Core color tokens via Tailwind v4 @theme directive
 * 2. This file (theme.ts) - Barrel that re-exports per-domain token modules
 *    plus the `cn` class-name composition helper
 * 3. Tailwind Classes - CSS-first configuration via @theme
 *
 * Token modules:
 *  - themeSpacing.ts    — margin / padding / gap / stack tokens
 *  - themeComponents.ts — button, input, card, badge, alert, modal variants
 *  - themeLayout.ts     — icon, radius, border, layout patterns
 *
 * BRAND COLORS:
 * - Primary: Stem Green (#2d7a3e / #81c784 dark)
 * - Accent: Lighter Stem Green (#4caf50 / #a5d6a7 dark)
 * - Gold: Mustard Gold (#d4a017 / #fbbf24 dark)
 *
 * STATUS COLORS (Industry Standard - DO NOT CHANGE):
 * - Success: Green (#28a745)
 * - Warning: Amber (#ffc107)
 * - Error: Red (#dc3545)
 * - Info: Cyan (#17a2b8)
 *
 * USAGE:
 * import { spacing, button, cn } from '../styles/theme';
 * <button className={cn(button.base, button.variant.primary)}>Action</button>
 *
 * =============================================================================
 */

// biome-ignore lint/performance/noBarrelFile: Design system barrel is intentional for API stability across ~100+ component import sites
export { status } from './themeColors';
export { alert, badge, button, card, input, modal } from './themeComponents';
export { border, icon, layout, radius } from './themeLayout';
export { spacing } from './themeSpacing';

/**
 * Combine class names with Tailwind class conflict resolution.
 */
export function cn(...classes: (string | boolean | undefined | null)[]): string {
  return twMerge(classes.filter(Boolean).join(' '));
}
