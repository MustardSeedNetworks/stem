/**
 * TestsTab
 *
 * Lists supported tests grouped by standard, or a flat search-result list.
 * Extracted verbatim from HelpDrawer; the framing intro string is localized
 * while the per-category names/summaries remain data-driven.
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { categories } from '../../data/help/categories';
import { getTestsByCategory } from '../../data/help/tests';
import type { TestHelp } from '../../data/help/types';
import { CollapsibleSection } from '../../shared/CollapsibleSection';
import { TestCard } from './TestCard';

interface TestsTabProps {
  filteredTests: TestHelp[] | null;
  simpleMode: boolean;
  onSelectTest: (test: TestHelp) => void;
}

const categoryOrder = [
  'rfc2544',
  'y1564',
  'rfc2889',
  'rfc6349',
  'y1731',
  'mef',
  'tsn',
  'trafficgen',
];

export function TestsTab({ filteredTests, simpleMode, onSelectTest }: TestsTabProps): ReactElement {
  const { t } = useTranslation(['help', 'common']);

  if (filteredTests) {
    return (
      <div className="stack-sm" data-testid="help-drawer-tests">
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
    // Named so a spec can assert on what the tab lists without matching loose
    // text across the whole drawer (#941).
    <div className="stack-lg" data-testid="help-drawer-tests">
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
            title={`${category.name} (${catTests.length})`}
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
