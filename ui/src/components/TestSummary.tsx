/**
 * TestSummary — the recap block that closes every test-configuration form.
 *
 * TSNConfigForm extracted this and the other six inlined the same JSX, which
 * is also how the heading came to read "Traffic Summary" in TrafficGen and
 * "Test Summary" in the other six. The divergence carried no meaning, so it
 * is resolved to one heading rather than parameterised — a `label` prop would
 * just be somewhere for it to drift back to.
 *
 * Rows are the caller's: what a summary should recap is per-test, and the
 * shared part is the frame, the heading and the row rhythm.
 */
import { Info } from 'lucide-react';
import type { ReactElement, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

interface TestSummaryProps {
  children: ReactNode;
}

export function TestSummary({ children }: TestSummaryProps): ReactElement {
  const { t } = useTranslation('settings');

  return (
    <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
      <div className="flex items-center gap-compact label mb-2">
        <Info className="w-4 h-4" />
        {t('testConfig.common.summary')}
      </div>
      <div className="text-xs text-text-muted stack-xs">{children}</div>
    </div>
  );
}
