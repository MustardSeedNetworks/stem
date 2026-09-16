import { HelpCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { Tooltip } from './ui/Tooltip';

interface HelpIconProps {
  tooltip: string;
  onClick?: () => void;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

const sizeClasses = { sm: 'w-3.5 h-3.5', md: 'w-4 h-4', lg: 'w-5 h-5' };

export function HelpIcon({
  tooltip,
  onClick,
  className = '',
  size = 'sm',
}: HelpIconProps): ReactElement {
  const { t } = useTranslation('common');
  return (
    <Tooltip
      text={
        <>
          {tooltip}
          {onClick ? (
            <span className="block text-brand-primary mt-tight">{t('help.clickForDetails')}</span>
          ) : null}
        </>
      }
    >
      <button
        type="button"
        data-testid="help-icon"
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onClick?.();
        }}
        className={`p-0.5 rounded-full hover:bg-surface-hover transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary ${className}`}
        aria-label={t('help.label', { topic: tooltip })}
      >
        <HelpCircle className={`${sizeClasses[size]} text-text-muted`} aria-hidden="true" />
      </button>
    </Tooltip>
  );
}

export default HelpIcon;
