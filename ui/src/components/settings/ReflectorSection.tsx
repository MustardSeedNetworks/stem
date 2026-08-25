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

/* The keys used to read `settings.reflector.netally` — a dot, which i18next
   parses as a path inside the default `common` namespace, not as the `settings`
   namespace. None of the eight resolved, so every label fell through to its
   hardcoded English default in both locales. They are nested under `.profiles`
   because `reflector.all` already means "All Traffic".

   The `t(key, default)` fallback that hid it is gone: that is the banned
   pattern, and semgrep misses this shape because the fallback was an object
   property rather than a literal argument.

   Keys are written out literally below rather than carried on the array, so
   the key checker can see them — the dynamic form is what let eight
   nonexistent keys sit here unnoticed. */
const REFLECTOR_PROFILE_IDS = ['netally', 'msn', 'all', 'custom'] as const;

export function ReflectorSection({
  profile,
  onProfileChange,
  className,
}: ReflectorSectionProps): React.JSX.Element {
  const { t } = useTranslation(['common', 'settings']);

  const profiles: Record<ReflectorProfile, { name: string; description: string }> = {
    netally: {
      name: t('settings:reflector.profiles.netally'),
      description: t('settings:reflector.profiles.netallyDesc'),
    },
    msn: {
      name: t('settings:reflector.profiles.msn'),
      description: t('settings:reflector.profiles.msnDesc'),
    },
    all: {
      name: t('settings:reflector.profiles.all'),
      description: t('settings:reflector.profiles.allDesc'),
    },
    custom: {
      name: t('settings:reflector.profiles.custom'),
      description: t('settings:reflector.profiles.customDesc'),
    },
  };

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Settings2 className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings:reflector.title')}</span>
        </div>
      }
      defaultOpen={true}
      className={className}
    >
      <div className="stack-sm">
        {REFLECTOR_PROFILE_IDS.map((id) => (
          <label
            key={id}
            className={cn(
              'flex items-center gap-default',
              spacing.pad.sm,
              radius.lg,
              'cursor-pointer hover:bg-surface-hover transition-colors',
            )}
          >
            <input
              type="radio"
              name="reflectorProfile"
              checked={profile === id}
              onChange={(): void => onProfileChange(id)}
              className="w-4 h-4 accent-brand-primary"
            />
            <div>
              <div className="body-small font-medium text-text-primary">{profiles[id].name}</div>
              <div className="caption text-text-muted">{profiles[id].description}</div>
            </div>
          </label>
        ))}
      </div>
    </CollapsibleSection>
  );
}

export default ReflectorSection;
