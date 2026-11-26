/**
 * ReflectorSection Component
 *
 * Reflector profile selection for packet reflection mode.
 * Allows selection of signature types to reflect.
 */

import { useTranslation } from 'react-i18next';
import { Card } from '../../shared/Card';
import { RadioGroup } from '../../shared/RadioGroup';
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
    <Card {...(className ? { className } : {})}>
      <RadioGroup
        label={t('settings:reflector.title')}
        name="reflectorProfile"
        value={profile}
        onValueChange={(value) => {
          const next = REFLECTOR_PROFILE_IDS.find((id) => id === value);
          if (next) {
            onProfileChange(next);
          }
        }}
        options={REFLECTOR_PROFILE_IDS.map((id) => ({
          value: id,
          label: profiles[id].name,
          description: profiles[id].description,
        }))}
      />
    </Card>
  );
}

export default ReflectorSection;
