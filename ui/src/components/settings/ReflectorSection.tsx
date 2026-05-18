/**
 * ReflectorSection Component
 *
 * Reflector profile selection for packet reflection mode.
 * Allows selection of signature types to reflect.
 */

import { Settings2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn, radius, spacing } from '../../styles/theme';
import { CollapsibleSection } from '../CollapsibleSection';
import type { ReflectorProfile, SettingsSectionProps } from './types';

interface ReflectorSectionProps extends SettingsSectionProps {
  profile: ReflectorProfile;
  onProfileChange: (profile: ReflectorProfile) => void;
}

const REFLECTOR_PROFILES = [
  {
    id: 'netally' as const,
    nameKey: 'settings.reflector.netally',
    nameDefault: 'NetAlly',
    descKey: 'settings.reflector.netallyDesc',
    descDefault: 'ITO signatures only',
  },
  {
    id: 'msn' as const,
    nameKey: 'settings.reflector.msn',
    nameDefault: 'MSN',
    descKey: 'settings.reflector.msnDesc',
    descDefault: 'Mustard Seed signatures',
  },
  {
    id: 'all' as const,
    nameKey: 'settings.reflector.all',
    nameDefault: 'All',
    descKey: 'settings.reflector.allDesc',
    descDefault: 'All signature types',
  },
  {
    id: 'custom' as const,
    nameKey: 'settings.reflector.custom',
    nameDefault: 'Custom',
    descKey: 'settings.reflector.customDesc',
    descDefault: 'Manual configuration',
  },
] as const;

export function ReflectorSection({
  profile,
  onProfileChange,
  className,
}: ReflectorSectionProps): React.JSX.Element {
  const { t } = useTranslation();

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Settings2 className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings.reflector.title', 'Reflector Profile')}</span>
        </div>
      }
      defaultOpen={true}
      className={className}
    >
      <div className="space-y-2">
        {REFLECTOR_PROFILES.map((p) => (
          <label
            key={p.id}
            className={cn(
              'flex items-center gap-3',
              spacing.pad.sm,
              radius.lg,
              'cursor-pointer hover:bg-surface-hover transition-colors',
            )}
          >
            <input
              type="radio"
              name="reflectorProfile"
              checked={profile === p.id}
              onChange={(): void => onProfileChange(p.id)}
              className="w-4 h-4 accent-brand-primary"
            />
            <div>
              <div className="body-small font-medium text-text-primary">
                {t(p.nameKey, p.nameDefault)}
              </div>
              <div className="caption text-text-muted">{t(p.descKey, p.descDefault)}</div>
            </div>
          </label>
        ))}
      </div>
    </CollapsibleSection>
  );
}

export default ReflectorSection;
