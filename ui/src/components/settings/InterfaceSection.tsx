/**
 * InterfaceSection Component
 *
 * Network interface selection with capability badges.
 * Shows XDP and physical interface indicators.
 */

import { Network } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, radius, spacing, status } from '../../styles/theme';
import { CollapsibleSection } from '../CollapsibleSection';
import type { InterfaceInfo, SettingsSectionProps } from './types';

interface InterfaceSectionProps extends SettingsSectionProps {
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  onInterfaceChange: (interfaceName: string) => void;
}

export function InterfaceSection({
  interfaces,
  selectedInterface,
  onInterfaceChange,
  className,
}: InterfaceSectionProps): React.JSX.Element {
  const { t } = useTranslation();

  const maxScore = useMemo(() => Math.max(...interfaces.map((i) => i.score), 0), [interfaces]);

  const selectedDetails = useMemo(
    () => interfaces.find((i) => i.name === selectedInterface),
    [interfaces, selectedInterface],
  );

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Network className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings.interface.title', 'Interface')}</span>
        </div>
      }
      defaultOpen={true}
      className={className}
    >
      <div className="space-y-3">
        <select
          value={selectedInterface}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
            onInterfaceChange(e.target.value)
          }
          className="w-full"
          aria-label={t('settings.interface.select', 'Select network interface')}
        >
          {interfaces.map((iface) => (
            <option key={iface.name} value={iface.name}>
              {iface.name} ({iface.speed}Mbps)
              {iface.score === maxScore && ' \u2605'}
            </option>
          ))}
        </select>

        {selectedDetails && (
          <div className="caption text-text-muted space-y-1">
            <div>
              {t('settings.interface.driver', 'Driver')}: {selectedDetails.driver}
            </div>
            <div className="flex gap-2">
              {selectedDetails.xdp === true && (
                <span
                  className={cn(
                    status.badge.success,
                    spacing.chip.sm,
                    radius.default,
                    'font-medium',
                  )}
                >
                  XDP
                </span>
              )}
              {selectedDetails.physical === true && (
                <span
                  className={cn(
                    status.badge.success,
                    spacing.chip.sm,
                    radius.default,
                    'font-medium',
                  )}
                >
                  {t('settings.interface.physical', 'Physical')}
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </CollapsibleSection>
  );
}

export default InterfaceSection;
