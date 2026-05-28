/**
 * TestsTab
 *
 * Lists supported tests grouped by standard, or a flat search-result list.
 * Extracted verbatim from HelpDrawer; the framing intro string is localized
 * while the per-category names/summaries remain data-driven.
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { categories, getTestsByCategory, type TestHelp } from '../../data/help-content';
import { CollapsibleSection } from '../CollapsibleSection';
import { TestCard } from './TestCard';

interface TestsTabProps {
  filteredTests: TestHelp[] | null;
  simpleMode: boolean;
  onSelectTest: (test: TestHelp) => void;
  getCategoryIcon: (id: string) => ReactElement;
}

const categoryOrder = ['rfc2544', 'y1564', 'rfc2889', 'rfc6349', 'y1731', 'mef', 'tsn'];

export function TestsTab({
  filteredTests,
  simpleMode,
  onSelectTest,
  getCategoryIcon,
}: TestsTabProps): ReactElement {
  const { t } = useTranslation(['help', 'common']);

  if (filteredTests) {
    return (
      <div className="stack-sm">
        <p className="text-xs text-text-muted mb-heading">
          {t('plurals.testCount', { ns: 'common', count: filteredTests.length })}
        </p>
        {filteredTests.map((test) => (
          <TestCard
            key={test.id}
            test={test}
            simpleMode={simpleMode}
            onClick={() => onSelectTest(test)}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="stack-lg">
      <p className="text-sm text-text-muted">{t('tests.intro')}</p>
      {categoryOrder.map((catId) => {
        const category = categories[catId];
        if (!category) {
          return null;
        }
        const catTests = getTestsByCategory(catId);

        return (
          <CollapsibleSection
            key={catId}
            title={
              <div className="flex items-center gap-compact">
                {getCategoryIcon(catId)}
                <span>{category.name}</span>
                <span className="text-xs text-text-muted">({catTests.length})</span>
              </div>
            }
            defaultOpen={catId === 'rfc2544'}
          >
            <div className="stack-sm">
              <p className="text-xs text-text-muted mb-2">{category.summary}</p>
              {catTests.map((test) => (
                <TestCard
                  key={test.id}
                  test={test}
                  simpleMode={simpleMode}
                  onClick={() => onSelectTest(test)}
                />
              ))}
            </div>
          </CollapsibleSection>
        );
      })}
    </div>
  );
}
