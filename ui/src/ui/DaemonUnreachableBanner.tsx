import { WifiOff } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { iconSizes } from '../constants/sizes';
import type { ConnectionState } from '../hooks/useConnectionStatus';

/**
 * Banner shown while the daemon does not answer. It renders nothing in any
 * other state, including the first check, so a slow start does not flash it.
 */
export function DaemonUnreachableBanner({
  status,
}: {
  status: ConnectionState;
}): ReactElement | null {
  const { t } = useTranslation('common');
  if (status !== 'disconnected') {
    return null;
  }
  return (
    <div
      role="alert"
      data-testid="daemon-unreachable-banner"
      className="mb-4 flex items-center gap-compact rounded-lg border border-status-error/30 bg-status-error/10 pad-sm text-status-error-strong"
    >
      <WifiOff className={`${iconSizes.md} flex-shrink-0`} aria-hidden="true" />
      <span>{t('connection.unreachable')}</span>
    </div>
  );
}
