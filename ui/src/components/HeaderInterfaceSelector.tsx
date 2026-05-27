/**
 * HeaderInterfaceSelector — compact interface picker for the top bar.
 *
 * Defaults to filtering for usable interfaces (`usable === true` from
 * /api/v1/interfaces). A small "Show all interfaces" toggle below
 * lets power users see virtual / down / unaddressed interfaces too.
 */
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { InterfaceInfo } from '../types/api';

interface HeaderInterfaceSelectorProps {
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  onSelectInterface: (name: string) => void;
  className?: string;
}

export function HeaderInterfaceSelector({
  interfaces,
  selectedInterface,
  onSelectInterface,
  className = '',
}: HeaderInterfaceSelectorProps): React.ReactElement {
  const { t } = useTranslation();
  const [showAll, setShowAll] = useState(false);

  const visibleInterfaces = useMemo(() => {
    if (showAll) {
      return interfaces;
    }
    const filtered = interfaces.filter((iface) => iface.usable === true);
    // If filtering yields nothing (e.g. dev box with only virtuals), fall
    // back to the full list so the dropdown is never empty.
    return filtered.length > 0 ? filtered : interfaces;
  }, [interfaces, showAll]);

  return (
    <div className={`flex flex-col gap-1 ${className}`}>
      <select
        value={selectedInterface}
        onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
          onSelectInterface(e.target.value)
        }
        className="w-48"
        aria-label={t('settings:interface.select', 'Select network interface')}
      >
        <option value="">{t('settings:interface.select', 'Select Interface')}</option>
        {visibleInterfaces.map((iface) => (
          <option key={iface.name} value={iface.name}>
            {iface.name} ({iface.speed}Mbps)
          </option>
        ))}
      </select>
      <label className="flex items-center gap-1.5 text-xs text-text-muted cursor-pointer select-none">
        <input
          type="checkbox"
          checked={showAll}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void => setShowAll(e.target.checked)}
          className="w-3 h-3 accent-brand-primary"
        />
        <span>{t('settings:interface.showAll', 'Show all interfaces')}</span>
      </label>
    </div>
  );
}

export default HeaderInterfaceSelector;
