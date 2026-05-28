/**
 * TestCard
 *
 * Compact, clickable summary row for a single test. Rendered in the Tests
 * tab list and category groups. Extracted verbatim from HelpDrawer.
 */

import { ChevronRight } from 'lucide-react';
import type { ReactElement } from 'react';
import type { TestHelp } from '../../data/help-content';
import { cn, icon as iconTokens, layout, radius, spacing } from '../../styles/theme';

interface TestCardProps {
  test: TestHelp;
  simpleMode: boolean;
  onClick: () => void;
}

export function TestCard({ test, simpleMode, onClick }: TestCardProps): ReactElement {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'w-full text-left border transition-colors',
        spacing.pad.sm,
        radius.lg,
        'border-surface-border hover:border-brand-primary hover:bg-surface-hover',
      )}
    >
      <div className={cn(layout.flex.between, 'items-start')}>
        <div className="flex-1">
          <div className="font-medium body-small text-text-primary">{test.name}</div>
          <div className="caption mt-0.5">{test.standard}</div>
          <p className="caption text-text-secondary mt-tight line-clamp-2">
            {simpleMode ? test.laymanDesc.split('\n')[0] : test.summary}
          </p>
        </div>
        <ChevronRight
          className={cn(iconTokens.size.sm, 'text-text-muted flex-shrink-0 mt-tight')}
        />
      </div>
    </button>
  );
}
