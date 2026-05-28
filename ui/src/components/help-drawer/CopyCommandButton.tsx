/**
 * CopyCommandButton
 *
 * Copy-to-clipboard control rendered beside example/tutorial commands.
 * Markup extracted verbatim from the inline copy buttons in HelpDrawer so
 * the detail views can share a single implementation. Chrome strings
 * (tooltip + aria-label) are localized.
 */

import { Check, Copy } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, icon as iconTokens, radius, status } from '../../styles/theme';

interface CopyCommandButtonProps {
  command: string;
  copiedCommand: string | null;
  onCopy: (cmd: string) => void;
}

export function CopyCommandButton({
  command,
  copiedCommand,
  onCopy,
}: CopyCommandButtonProps): ReactElement {
  const { t } = useTranslation('help');
  return (
    <button
      type="button"
      onClick={(): void => onCopy(command)}
      className={cn('p-1 hover:bg-surface-hover', radius.default)}
      title={t('detail.copyCommand')}
      aria-label={t('detail.copyCommandAria')}
    >
      {copiedCommand === command ? (
        <Check className={cn(iconTokens.size.xs, status.text.success)} />
      ) : (
        <Copy className={cn(iconTokens.size.xs, 'text-text-muted')} />
      )}
    </button>
  );
}
