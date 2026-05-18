/**
 * @fileoverview The Stem - Help Icon Component
 * @description A (?) icon that shows a tooltip on hover and opens help on click.
 */

import { HelpCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useEffect, useRef, useState } from 'react';

interface HelpIconProps {
  tooltip: string;
  onClick?: () => void;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function HelpIcon({
  tooltip,
  onClick,
  className = '',
  size = 'sm',
}: HelpIconProps): ReactElement {
  const [showTooltip, setShowTooltip] = useState(false);
  const [tooltipPosition, setTooltipPosition] = useState<'top' | 'bottom'>('top');
  const iconRef = useRef<HTMLButtonElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  const sizeClasses = {
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  // Calculate tooltip position based on available space
  useEffect(() => {
    if (showTooltip && iconRef.current && tooltipRef.current) {
      const iconRect = iconRef.current.getBoundingClientRect();
      const tooltipRect = tooltipRef.current.getBoundingClientRect();

      // Check if there's room above
      if (iconRect.top - tooltipRect.height - 8 < 0) {
        setTooltipPosition('bottom');
      } else {
        setTooltipPosition('top');
      }
    }
  }, [showTooltip]);

  return (
    <div className={`relative inline-block ${className}`}>
      <button
        ref={iconRef}
        type="button"
        onClick={(e: React.MouseEvent<HTMLButtonElement>): void => {
          e.stopPropagation();
          onClick?.();
        }}
        onMouseEnter={(): void => setShowTooltip(true)}
        onMouseLeave={(): void => setShowTooltip(false)}
        onFocus={(): void => setShowTooltip(true)}
        onBlur={(): void => setShowTooltip(false)}
        className="p-0.5 rounded-full hover:bg-[var(--color-surface-hover)] transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/50"
        title={onClick ? `${tooltip} (click for details)` : tooltip}
        aria-label={`Help: ${tooltip}`}
      >
        <HelpCircle
          className={`${sizeClasses[size]} text-[var(--color-text-muted)] hover:text-[var(--color-brand-primary)] transition-colors`}
        />
      </button>

      {/* Tooltip */}
      {showTooltip ? (
        <div
          ref={tooltipRef}
          role="tooltip"
          className={`absolute z-50 px-2 py-1.5 text-xs bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] rounded-lg shadow-lg max-w-xs whitespace-normal text-[var(--color-text-secondary)] ${
            tooltipPosition === 'top'
              ? 'bottom-full mb-2 left-1/2 -translate-x-1/2'
              : 'top-full mt-2 left-1/2 -translate-x-1/2'
          }`}
        >
          {tooltip}
          {onClick ? (
            <span className="block text-[var(--color-brand-primary)] mt-1 text-[10px]">
              Click for more details
            </span>
          ) : null}
          {/* Tooltip Arrow */}
          <div
            className={`absolute w-2 h-2 bg-[var(--color-surface-raised)] border-[var(--color-surface-border)] rotate-45 ${
              tooltipPosition === 'top'
                ? 'top-full -mt-1 left-1/2 -translate-x-1/2 border-r border-b'
                : 'bottom-full -mb-1 left-1/2 -translate-x-1/2 border-l border-t'
            }`}
          />
        </div>
      ) : null}
    </div>
  );
}

export default HelpIcon;
