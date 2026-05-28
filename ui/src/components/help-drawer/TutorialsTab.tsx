/**
 * TutorialsTab
 *
 * Search-filterable list of tutorial summaries. Extracted verbatim from
 * HelpDrawer. The framing intro string is localized; tutorial titles,
 * durations, levels and descriptions remain data-driven.
 */

import { ChevronRight } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { type Tutorial, tutorials } from '../../data/help-content';
import { cn, icon as iconTokens, layout, radius, spacing } from '../../styles/theme';

interface TutorialsTabProps {
  searchQuery: string;
  onSelectTutorial: (tutorial: Tutorial) => void;
}

export function TutorialsTab({ searchQuery, onSelectTutorial }: TutorialsTabProps): ReactElement {
  const { t } = useTranslation('help');
  const tutorialList: Tutorial[] = Object.values(tutorials);
  const filtered = searchQuery
    ? tutorialList.filter(
        (tutorial: Tutorial) =>
          tutorial.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
          tutorial.description.toLowerCase().includes(searchQuery.toLowerCase()),
      )
    : tutorialList;

  return (
    <div className="section-gap">
      <p className="body-small text-text-muted">{t('tutorials.intro')}</p>
      <div className="stack-sm">
        {filtered.map((tutorial: Tutorial) => (
          <button
            type="button"
            key={tutorial.id}
            onClick={(): void => onSelectTutorial(tutorial)}
            className={cn(
              'w-full text-left border transition-colors',
              spacing.pad.default,
              radius.lg,
              'border-surface-border hover:border-brand-primary hover:bg-surface-hover',
            )}
          >
            <div className={cn(layout.flex.between, 'items-start')}>
              <div className="flex-1">
                <div className="font-medium body-small text-text-primary">{tutorial.title}</div>
                <div className={cn(layout.inline.default, 'mt-tight')}>
                  <span className="caption">{tutorial.duration}</span>
                  <span className={cn('caption px-1.5 py-0.5 bg-surface-base', radius.default)}>
                    {tutorial.level}
                  </span>
                </div>
                <p className="caption text-text-secondary mt-inline">{tutorial.description}</p>
              </div>
              <ChevronRight className={cn(iconTokens.size.sm, 'text-text-muted flex-shrink-0')} />
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
