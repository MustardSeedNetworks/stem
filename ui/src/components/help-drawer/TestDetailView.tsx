/**
 * TestDetailView
 *
 * Full documentation view for a single selected test: description, usage,
 * parameters, metrics, examples, tips and related tests. Extracted verbatim
 * from HelpDrawer. Section headings and other framing strings are localized;
 * the test content itself remains data-driven (English in the data file).
 */

import { ChevronRight } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { TestHelp } from '../../data/help-content';
import { cn, icon as iconTokens, layout, radius, spacing, status } from '../../styles/theme';
import { CopyCommandButton } from './CopyCommandButton';

interface TestDetailViewProps {
  test: TestHelp;
  simpleMode: boolean;
  onBack: () => void;
  onCopy: (cmd: string) => void;
  copiedCommand: string | null;
}

export function TestDetailView({
  test,
  simpleMode,
  onBack,
  onCopy,
  copiedCommand,
}: TestDetailViewProps): ReactElement {
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
        {t('detail.backToTests')}
      </button>

      {/* Header */}
      <div>
        <h3 className="heading-3">{test.name}</h3>
        <p className="caption">{test.standard}</p>
      </div>

      {/* Description */}
      <div className={cn('bg-surface-base', radius.lg, spacing.pad.default)}>
        <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
          {simpleMode ? t('detail.descriptionSimple') : t('detail.descriptionTechnical')}
        </h4>
        <p className="body-small text-text-primary whitespace-pre-line">
          {simpleMode ? test.laymanDesc : test.techDesc}
        </p>
      </div>

      {/* When to Use */}
      <div className={cn(status.bg.successSoft, radius.lg, spacing.pad.default)}>
        <h4 className={cn('section-title', status.text.success, spacing.margin.bottom.inline)}>
          {t('detail.whenToUse')}
        </h4>
        <p className="body-small text-text-primary whitespace-pre-line">{test.whenToUse}</p>
      </div>

      {/* When NOT to Use */}
      {test.whenNotToUse ? (
        <div className={cn(status.bg.warningSoft, radius.lg, spacing.pad.default)}>
          <h4 className={cn('section-title', status.text.warning, spacing.margin.bottom.inline)}>
            {t('detail.whenNotToUse')}
          </h4>
          <p className="body-small text-text-primary whitespace-pre-line">{test.whenNotToUse}</p>
        </div>
      ) : null}

      {/* Parameters */}
      {test.parameters.length > 0 ? (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
            {t('detail.parameters')}
          </h4>
          <div className="stack-sm">
            {test.parameters.map((param) => (
              <div key={param.flag} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.inline.default}>
                  <code className={cn('code', 'bg-surface-raised px-1.5 py-0.5', radius.default)}>
                    {param.flag}
                  </code>
                  {param.required ? (
                    <span className={cn('caption', status.text.warning)}>
                      {t('detail.required')}
                    </span>
                  ) : null}
                </div>
                <p className="caption mt-tight">
                  {t('detail.parameterMeta', { type: param.type, default: param.defaultValue })}
                </p>
                <p className="body-small text-text-secondary mt-tight">
                  {simpleMode ? param.laymanDesc : param.techDesc}
                </p>
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {/* Metrics */}
      {test.metrics.length > 0 ? (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
            {t('detail.metrics')}
          </h4>
          <div className="stack-sm">
            {test.metrics.map((metric) => (
              <div key={metric.name} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.flex.between}>
                  <span className="font-medium body-small">{metric.name}</span>
                  <span className="caption">{metric.unit}</span>
                </div>
                <p className={cn('caption mt-tight', status.text.success)}>
                  {t('detail.metricGood', { value: metric.goodRange })}
                </p>
                <p className={cn('caption mt-0.5', status.text.warning)}>
                  {t('detail.metricBad', { value: metric.badMeaning })}
                </p>
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {/* Examples */}
      {test.examples.length > 0 ? (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
            {t('detail.examples')}
          </h4>
          <div className="stack-sm">
            {test.examples.map((example) => (
              <div
                key={example.command}
                className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}
              >
                <p className="caption mb-tight">{example.desc}</p>
                <div
                  className={cn(layout.flex.between, 'bg-surface-raised pad-xs', radius.default)}
                >
                  <code className="caption font-mono text-brand-primary">$ {example.command}</code>
                  <CopyCommandButton
                    command={example.command}
                    copiedCommand={copiedCommand}
                    onCopy={onCopy}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {/* Tips */}
      {test.tips.length > 0 ? (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>{t('detail.tips')}</h4>
          <ul className="stack-xs">
            {test.tips.map((tip) => (
              <li
                key={tip}
                className={cn(layout.inline.default, 'body-small text-text-secondary items-start')}
              >
                <span className="text-brand-primary">-</span>
                {tip}
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {/* See Also */}
      {test.seeAlso.length > 0 ? (
        <div className="pt-section border-t border-surface-border">
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
            {t('detail.relatedTests')}
          </h4>
          <div className={layout.inline.wrap}>
            {test.seeAlso.map((related: string) => (
              <span
                key={related}
                className={cn('caption bg-surface-base px-cell py-compact', radius.default)}
              >
                {related}
              </span>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
