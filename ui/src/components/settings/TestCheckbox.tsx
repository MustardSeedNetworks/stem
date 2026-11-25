/**
 * TestCheckbox Component
 *
 * Reusable checkbox for test selection, with the test's help beside it.
 * Used across all test section components.
 */

import { Checkbox } from '../../shared/Checkbox';
import { HelpIcon } from '../HelpIcon';
import type { TestDefinition } from './types';

interface TestCheckboxProps {
  test: TestDefinition;
  checked: boolean;
  onChange: () => void;
}

export function TestCheckbox({ test, checked, onChange }: TestCheckboxProps): React.JSX.Element {
  return (
    <div className="flex items-start gap-1 rounded-lg p-2 hover:bg-surface-hover">
      <Checkbox
        className="min-w-0 flex-1"
        checked={checked}
        onChange={onChange}
        label={test.name}
        description={test.desc}
      />
      <HelpIcon tooltip={test.tooltip} />
    </div>
  );
}

export default TestCheckbox;
