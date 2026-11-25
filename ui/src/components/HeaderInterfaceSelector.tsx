/**
 * HeaderInterfaceSelector — compact interface picker for the top bar.
 *
 * Defaults to filtering for usable interfaces (`usable === true` from
 * /api/v1/interfaces). A small "Show all interfaces" toggle below
 * lets power users see virtual / down / unaddressed interfaces too.
 */
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Checkbox } from '../shared/Checkbox';
import { FormField } from '../shared/FormField';
import { Select } from '../shared/Select';
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
  const { t } = useTranslation(['common', 'settings']);
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
    <div className={`flex w-52 max-w-full flex-col gap-2 ${className}`}>
      <FormField label={t('settings:interface.select')}>
        {(control) => (
          <Select
            {...control}
            value={selectedInterface}
            onValueChange={onSelectInterface}
            placeholder={t('settings:interface.select')}
            options={visibleInterfaces.map((iface) => ({
              value: iface.name,
              label: `${iface.name} (${iface.speed}Mbps)`,
            }))}
          />
        )}
      </FormField>
      <Checkbox checked={showAll} onChange={setShowAll} label={t('settings:interface.showAll')} />
    </div>
  );
}

export default HeaderInterfaceSelector;
