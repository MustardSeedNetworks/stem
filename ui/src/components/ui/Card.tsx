/**
 * Card Component
 *
 * Base card container used throughout the application for displaying information.
 *
 * Features:
 * - Status badge with color coding (success, warning, error, loading, unknown)
 * - Header with title, optional subtitle, icon, and custom actions
 * - Keyboard accessibility (Enter/Space to activate)
 * - Optional click handler for interactive cards
 * - Responsive padding (mobile and desktop)
 * - Smooth transitions and hover effects
 * - Focus ring for keyboard navigation
 */

import type React from 'react';
import type { ReactNode } from 'react';
import { card, cn, icon as iconTokens, layout, spacing } from '../../styles/theme';
import { StatusBadge } from './StatusBadge';
import { getStatusConfig, type Status } from './StatusConfig';

// Re-export Status type for convenience
export type { Status };

// Type-safe size class getter
function getSizeClass(size: 'sm' | 'md' | 'lg'): string {
  switch (size) {
    case 'sm':
      return 'body-small';
    case 'md':
      return 'body font-medium leading-snug';
    case 'lg':
      return 'body-large font-semibold leading-snug';
    default:
      return 'body font-medium leading-snug';
  }
}

/**
 * Props for the Card component
 */
interface CardProps {
  /** Card title (required) */
  title: string;
  /** Optional subtitle or description */
  subtitle?: string;
  /** Status indicator color (success, warning, error, loading, unknown) */
  status: Status;
  /** Card content */
  children: ReactNode;
  /** Additional CSS classes */
  className?: string;
  /** Callback when card is clicked */
  onClick?: () => void;
  /** Optional icon displayed in header */
  icon?: ReactNode;
  /** Optional action element in header (right side) */
  headerAction?: ReactNode;
  /** Enable aria-live region for dynamic content updates */
  enableLiveRegion?: boolean;
  /** ARIA label for the card */
  ariaLabel?: string;
}

/**
 * Card component - displays information in a styled container with status badge.
 */
export function Card({
  title,
  subtitle,
  status,
  children,
  className = '',
  onClick,
  icon,
  headerAction,
  enableLiveRegion = false,
  ariaLabel,
}: CardProps): React.JSX.Element {
  // Card is interactive if click handler is provided
  const isInteractive = typeof onClick === 'function';

  /**
   * Handle keyboard activation (Enter/Space) for interactive cards.
   */
  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>): void => {
    if (!isInteractive) {
      return;
    }
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick?.();
    }
  };

  // Generate title ID for accessibility
  const titleId = `card-title-${title.replace(/\s+/g, '-').toLowerCase()}`;

  // Add aria-label and aria-live for accessibility
  const computedAriaLabel = ariaLabel || (subtitle ? `${title} - ${subtitle}` : title);
  const ariaProps = {
    'aria-label': computedAriaLabel,
    ...(enableLiveRegion && {
      'aria-live': 'polite' as const,
      'aria-atomic': 'true' as const,
    }),
  };

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: role is dynamically set based on isInteractive
    <div
      className={cn(
        card.base,
        card.variant.default,
        spacing.pad.sm,
        'sm:pad',
        // Fixed width for consistent card grid layout
        'w-full max-w-sm',
        'transition-all hover:border-brand-primary/40 touch-manipulation focus-visible:ring-2 focus-visible:ring-brand-primary focus-visible:ring-offset-2 focus-visible:ring-offset-surface-base outline-none',
        isInteractive && 'cursor-pointer active:scale-[0.98]',
        className,
      )}
      onClick={onClick}
      onKeyDown={handleKeyDown}
      role={isInteractive ? 'button' : undefined}
      tabIndex={isInteractive ? 0 : undefined}
      {...ariaProps}
    >
      <div className={layout.flex.between}>
        <div className={layout.inline.default}>
          {icon ? (
            <span className={cn('text-text-muted shrink-0', iconTokens.size.md)} aria-hidden="true">
              {icon}
            </span>
          ) : null}
          <div className={layout.flex.col}>
            <h3 className="heading-4 font-display" id={titleId}>
              {title}
            </h3>
            {subtitle ? <p className="caption leading-tight">{subtitle}</p> : null}
          </div>
        </div>
        <div className={layout.inline.default}>
          {headerAction}
          <StatusBadge status={status} size="md" />
        </div>
      </div>
      <div className={cn(spacing.margin.top.inline, 'sm:mt-content')} aria-describedby={titleId}>
        {children}
      </div>
    </div>
  );
}

interface CardValueProps {
  label?: string;
  value: string | number;
  unit?: string;
  size?: 'sm' | 'md' | 'lg';
  status?: Status;
  mono?: boolean;
  allowWrap?: boolean;
}

/**
 * Displays a prominent value with optional label, unit, and status indicator.
 */
export function CardValue({
  label,
  value,
  unit,
  size = 'md',
  status,
  mono = false,
  allowWrap = false,
}: CardValueProps): React.JSX.Element {
  const statusColorClass = status ? getStatusConfig(status).color : 'text-text-primary';
  const textMods = [
    statusColorClass,
    mono ? 'font-mono tabular-nums' : '',
    allowWrap ? 'break-all whitespace-pre-wrap' : '',
  ]
    .filter(Boolean)
    .join(' ');

  const statusIcon = status ? getStatusConfig(status).icon : null;

  return (
    <div>
      {label ? <p className={cn('caption', spacing.margin.bottom.tight)}>{label}</p> : null}
      <p className={cn(getSizeClass(size), textMods, layout.inline.tight)} data-testid="card-value">
        {statusIcon ? (
          <span className={cn(layout.flex.center, iconTokens.size.xs, 'shrink-0 text-current')}>
            {statusIcon}
          </span>
        ) : null}
        <span className={cn(layout.inline.tight, 'items-baseline')}>
          <span>{value}</span>
          {unit ? <span className="body-small font-normal text-text-muted">{unit}</span> : null}
        </span>
      </p>
    </div>
  );
}

interface CardRowProps {
  label: string;
  value: string | number;
  status?: Status;
  wrap?: boolean;
  mono?: boolean;
  align?: 'left' | 'right';
}

/**
 * Displays a label-value pair in a horizontal row with optional status indicator.
 */
export function CardRow({
  label,
  value,
  status,
  wrap = false,
  mono = false,
  align = 'right',
}: CardRowProps): React.JSX.Element {
  const resolvedStatus = status ? getStatusConfig(status) : null;
  const statusIcon = resolvedStatus?.icon ?? null;
  const justifyClass = align === 'right' ? 'justify-end' : 'justify-start';

  return (
    <div
      className={cn(
        'flex justify-between',
        spacing.compact.py,
        layout.inline.default,
        wrap ? 'items-start' : 'items-center',
      )}
    >
      <span className="body-small shrink-0">{label}</span>
      <span
        className={cn(
          'body-small font-medium',
          layout.inline.tight,
          justifyClass,
          align === 'right' ? 'text-right' : 'text-left',
          wrap ? 'break-all whitespace-pre-wrap' : 'truncate',
          mono && 'font-mono tabular-nums',
          resolvedStatus?.color ?? 'text-text-primary',
        )}
        title={String(value)}
        data-testid="card-row-value"
      >
        {statusIcon ? (
          <span className={cn(iconTokens.size.xs, 'shrink-0 text-current')}>{statusIcon}</span>
        ) : null}
        <span>{value}</span>
      </span>
    </div>
  );
}

interface CardDividerProps {
  className?: string;
}

/**
 * Horizontal divider line for separating card sections.
 */
export function CardDivider({ className = '' }: CardDividerProps): React.JSX.Element {
  return <hr className={cn('border-surface-border', spacing.margin.top.content, className)} />;
}
