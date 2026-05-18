/**
 * themeColors.ts — semantic color tokens for status states.
 * Re-exported through theme.ts.
 *
 * Status colors follow industry conventions (do not change):
 *   - success: green   (#28a745)
 *   - warning: amber   (#ffc107)
 *   - error:   red     (#dc3545)
 *   - info:    cyan    (#17a2b8)
 *
 * Underlying values are defined via @theme in src/index.css.
 *
 * Four composition surfaces, picked by what the call site needs:
 *   - text.*    — text color only
 *   - bg.*      — background only, with intensity variants (solid/strong/soft/subtle)
 *   - border.*  — border color (soft = 20% alpha)
 *   - badge.*   — compound bg+text for chip/pill/banner patterns
 *   - hover.*   — hover-state background tints
 */

const statusText = {
  success: 'text-status-success',
  warning: 'text-status-warning',
  error: 'text-status-error',
  info: 'text-status-info',
} as const;

const statusBg = {
  // 100% — pulse dot, prominent fill
  success: 'bg-status-success',
  warning: 'bg-status-warning',
  error: 'bg-status-error',
  info: 'bg-status-info',

  // 20% — emphasized chip / active toggle
  successStrong: 'bg-status-success/20',
  warningStrong: 'bg-status-warning/20',
  errorStrong: 'bg-status-error/20',
  infoStrong: 'bg-status-info/20',

  // 10% — standard chip backdrop
  successSoft: 'bg-status-success/10',
  warningSoft: 'bg-status-warning/10',
  errorSoft: 'bg-status-error/10',
  infoSoft: 'bg-status-info/10',

  // 5% — subtle row highlight
  successSubtle: 'bg-status-success/5',
  warningSubtle: 'bg-status-warning/5',
  errorSubtle: 'bg-status-error/5',
  infoSubtle: 'bg-status-info/5',
} as const;

const statusBorder = {
  // 20% — soft border accompanying a soft bg
  successSoft: 'border-status-success/20',
  warningSoft: 'border-status-warning/20',
  errorSoft: 'border-status-error/20',
  infoSoft: 'border-status-info/20',
} as const;

const statusBadge = {
  // 10% bg + matching text — standard chip
  success: 'bg-status-success/10 text-status-success',
  warning: 'bg-status-warning/10 text-status-warning',
  error: 'bg-status-error/10 text-status-error',
  info: 'bg-status-info/10 text-status-info',

  // 20% bg + matching text — emphasized chip / active state
  successStrong: 'bg-status-success/20 text-status-success',
  warningStrong: 'bg-status-warning/20 text-status-warning',
  errorStrong: 'bg-status-error/20 text-status-error',
  infoStrong: 'bg-status-info/20 text-status-info',
} as const;

const statusHover = {
  successStrong: 'hover:bg-status-success/20',
  warningStrong: 'hover:bg-status-warning/20',
  errorStrong: 'hover:bg-status-error/20',
  infoStrong: 'hover:bg-status-info/20',
} as const;

/**
 * Small circular indicator base classes — combine with statusBg.<color>
 * and 'animate-pulse' for a live indicator.
 */
const statusDot = 'inline-block w-2 h-2 rounded-full';

export const status = {
  text: statusText,
  bg: statusBg,
  border: statusBorder,
  badge: statusBadge,
  hover: statusHover,
  dot: statusDot,
} as const;
