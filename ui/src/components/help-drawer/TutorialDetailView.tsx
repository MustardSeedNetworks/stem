/**
 * TutorialDetailView
 *
 * Step-by-step walkthrough for a single selected tutorial. Extracted verbatim
 * from HelpDrawer. The back button and the "Expected"/"Tip" labels are
 * localized; tutorial step content remains data-driven.
 */

import { ChevronRight } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { Tutorial, TutorialStep } from '../../data/help-content';
import { cn, icon as iconTokens, layout, radius, spacing, status } from '../../styles/theme';
import { CopyCommandButton } from './CopyCommandButton';

interface TutorialDetailViewProps {
  tutorial: Tutorial;
  onBack: () => void;
  onCopy: (cmd: string) => void;
  copiedCommand: string | null;
}

export function TutorialDetailView({
  tutorial,
  onBack,
  onCopy,
  copiedCommand,
}: TutorialDetailViewProps): ReactElement {
  const { t } = useTranslation('help');
  return (
    <div className="section-gap">
      {/* Back Button */}
      <button
        type="button"
        onClick={onBack}
        className={cn(layout.inline.tight, 'body-small text-text-muted hover:text-text-primary')}
      >
        <ChevronRight className={cn(iconTokens.size.sm, 'rotate-180')} />
        {t('tutorials.backToTutorials')}
      </button>

      {/* Header */}
      <div>
        <h3 className="heading-3">{tutorial.title}</h3>
        <div className={cn(layout.inline.default, 'mt-tight')}>
          <span className="caption">{tutorial.duration}</span>
          <span className={cn('caption px-1.5 py-0.5 bg-surface-base', radius.default)}>
            {tutorial.level}
          </span>
        </div>
        <p className="body-small text-text-secondary mt-inline">{tutorial.description}</p>
      </div>

      {/* Steps */}
      <div className="stack-lg">
        {tutorial.steps.map((step: TutorialStep, stepIndex: number) => (
          <div key={step.title} className={cn('bg-surface-base', radius.lg, spacing.pad.default)}>
            <div className={cn(layout.inline.default, spacing.margin.bottom.inline)}>
              <span
                className={cn(
                  'w-6 h-6 text-xs font-medium',
                  radius.full,
                  'bg-brand-primary text-text-inverse',
                  layout.flex.center,
                )}
              >
                {stepIndex + 1}
              </span>
              <h4 className="font-medium body-small text-text-primary">{step.title}</h4>
            </div>
            <p
              className={cn(
                'body-small text-text-secondary whitespace-pre-line',
                spacing.margin.bottom.heading,
              )}
            >
              {step.content}
            </p>
            {step.command ? (
              <div
                className={cn(layout.flex.between, 'bg-surface-raised pad-xs mb-2', radius.default)}
              >
                <code className="caption font-mono text-brand-primary">$ {step.command}</code>
                <CopyCommandButton
                  command={step.command}
                  copiedCommand={copiedCommand}
                  onCopy={onCopy}
                />
              </div>
            ) : null}
            {step.expected ? (
              <div className={cn('caption bg-surface-raised pad-xs', radius.default)}>
                {t('tutorials.expected', { value: step.expected })}
              </div>
            ) : null}
            {step.tip ? (
              <div
                className={cn(cn('mt-inline caption pad-xs', status.badge.info), radius.default)}
              >
                {t('tutorials.tip', { value: step.tip })}
              </div>
            ) : null}
          </div>
        ))}
      </div>
    </div>
  );
}
