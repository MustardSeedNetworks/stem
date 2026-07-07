/**
 * GlossaryTab
 *
 * Glossary entries grouped by category, optionally filtered by search.
 * Extracted verbatim from HelpDrawer. The intro line, result count and
 * "Related" label are localized; glossary terms and definitions remain
 * data-driven.
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { glossary } from '../../data/help/glossary';
import type { GlossaryEntry } from '../../data/help/types';
import { cn, layout, radius, spacing } from '../../styles/theme';
import { CollapsibleSection } from '../CollapsibleSection';

interface GlossaryTabProps {
  searchQuery: string;
  filteredGlossary: GlossaryEntry[] | null;
  simpleMode: boolean;
}

export function GlossaryTab({
  searchQuery,
  filteredGlossary,
  simpleMode,
}: GlossaryTabProps): ReactElement {
  const { t } = useTranslation(['help', 'common']);
  const glossaryEntries = filteredGlossary || Object.values(glossary);

  // Group by category
  const byCategory = glossaryEntries.reduce(
    (acc, entry) => {
      if (!acc[entry.category]) {
        acc[entry.category] = [];
      }
      acc[entry.category].push(entry);
      return acc;
    },
    {} as Record<string, GlossaryEntry[]>,
  );

  const categoryNames = Object.keys(byCategory).sort();

  return (
    <div className="section-gap">
      <p className="body-small text-text-muted">
        {simpleMode ? t('glossary.introSimple') : t('glossary.introTechnical')}
      </p>
      {searchQuery ? (
        <p className="caption">
          {t('plurals.glossaryEntryCount', { ns: 'common', count: glossaryEntries.length })}
        </p>
      ) : null}
      {categoryNames.map((category) => (
        <CollapsibleSection
          key={category}
          title={<span>{category}</span>}
          defaultOpen={searchQuery.length > 0}
        >
          <div className="stack-sm">
            {byCategory[category].map((entry: GlossaryEntry) => (
              <div key={entry.term} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.inline.default}>
                  <span className="font-medium body-small text-text-primary">{entry.term}</span>
                  <span className="caption">{entry.fullName}</span>
                </div>
                <p className="body-small text-text-secondary mt-tight">
                  {simpleMode ? entry.laymanDef : entry.techDef}
                </p>
                {entry.related.length > 0 && (
                  <div className={cn(layout.inline.tight, 'mt-inline')}>
                    <span className="caption">{t('glossary.related')}</span>
                    {entry.related.map((r: string) => (
                      <span key={r} className="caption text-brand-primary">
                        {r}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </CollapsibleSection>
      ))}
    </div>
  );
}
