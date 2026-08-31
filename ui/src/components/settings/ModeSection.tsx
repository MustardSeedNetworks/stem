/**
 * ModeSection Component
 *
 * Operating mode selection between Reflector and Test Master modes.
 * Uses theme tokens for consistent styling.
 */

import { Monitor } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn, radius, spacing } from '../../styles/theme';
import { CollapsibleSection } from '../CollapsibleSection';
import type { OperatingMode, SettingsSectionProps } from './types';

interface ModeSectionProps extends SettingsSectionProps {
  mode: OperatingMode;
  onModeChange: (mode: OperatingMode) => void;
}

// Namespace-qualified with a colon, like the title above. A dot makes
// i18next look in the DEFAULT namespace (common) for a key literally named
// "settings.mode.reflector", which does not exist -- so every option fell
// through to its hardcoded English default and Spanish rendered "Reflector
// Mode" under a translated heading. The defaults are gone with it: a fallback
// that fires silently is what hid this, and one that cannot fire is dead.
const MODES = [
  {
    id: 'reflector' as const,
    nameKey: 'settings:mode.reflector',
    descKey: 'settings:mode.reflectorDesc',
  },
  {
    id: 'test_master' as const,
    nameKey: 'settings:mode.testMaster',
    descKey: 'settings:mode.testMasterDesc',
  },
] as const;

export function ModeSection({
  mode,
  onModeChange,
  className,
}: ModeSectionProps): React.JSX.Element {
  const { t } = useTranslation(['common', 'settings']);

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Monitor className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings:mode.title')}</span>
        </div>
      }
      defaultOpen={true}
      className={className}
    >
      <div className="stack-sm">
        {MODES.map((modeOption) => (
          <label
            key={modeOption.id}
            className={cn(
              'flex items-center gap-default',
              spacing.pad.sm,
              radius.lg,
              'cursor-pointer hover:bg-surface-hover transition-colors',
            )}
          >
            <input
              type="radio"
              name="operatingMode"
              checked={mode === modeOption.id}
              onChange={(): void => onModeChange(modeOption.id)}
              className="w-4 h-4 accent-brand-primary"
            />
            <div>
              <div className="body-small font-medium text-text-primary">
                {t(modeOption.nameKey)}
              </div>
              <div className="caption text-text-muted">{t(modeOption.descKey)}</div>
            </div>
          </label>
        ))}
      </div>
    </CollapsibleSection>
  );
}

export default ModeSection;
