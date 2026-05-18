// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * TestCheckbox Component
 *
 * Reusable checkbox for test selection with tooltip support.
 * Used across all test section components.
 */

import { cn, radius, spacing } from '../../styles/theme';
import { HelpIcon } from '../HelpIcon';
import type { TestDefinition } from './types';

interface TestCheckboxProps {
  test: TestDefinition;
  checked: boolean;
  onChange: () => void;
}

export function TestCheckbox({ test, checked, onChange }: TestCheckboxProps): React.JSX.Element {
  return (
    <label
      title={test.tooltip}
      className={cn(
        'flex items-start gap-3',
        spacing.pad.sm,
        radius.lg,
        'cursor-pointer hover:bg-surface-hover transition-colors',
      )}
    >
      <input
        type="checkbox"
        checked={checked}
        onChange={onChange}
        aria-label={`Toggle ${test.name}`}
        className="mt-0.5 w-4 h-4 accent-brand-primary"
      />
      <div className="flex-1">
        <div className="body-small font-medium text-text-primary flex items-center gap-1">
          {test.name}
          <HelpIcon tooltip={test.tooltip} />
        </div>
        <div className="caption text-text-muted">{test.desc}</div>
      </div>
    </label>
  );
}

export default TestCheckbox;
