import { CircleHelp } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { Tooltip } from '../shared/Tooltip';

interface HelpIconProps {
  tooltip: string;
  className?: string;
}

/**
 * A (?) beside a heading or checkbox, showing its help on hover and focus.
 * Labelled by the help text itself so a screen reader hears it too. Field help
 * is a FormField hint instead; this is for the places a hint has no home.
 */
export function HelpIcon({ tooltip, className = '' }: HelpIconProps): ReactElement {
  const { t } = useTranslation('common');
  return (
    <Tooltip text={tooltip}>
      <button
        type="button"
        data-testid="help-icon"
        onClick={(event) => {
          // Inside a label this must not toggle the control it sits beside.
          event.preventDefault();
          event.stopPropagation();
        }}
        className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-text-muted hover:bg-surface-hover hover:text-text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand-primary ${className}`}
        aria-label={t('help.label', { topic: tooltip })}
      >
        <CircleHelp className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
    </Tooltip>
  );
}

export default HelpIcon;
