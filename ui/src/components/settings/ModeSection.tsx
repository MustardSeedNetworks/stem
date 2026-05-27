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

const MODES = [
  {
    id: 'reflector' as const,
    nameKey: 'settings.mode.reflector',
    nameDefault: 'Reflector Mode',
    descKey: 'settings.mode.reflectorDesc',
    descDefault: 'Packet reflection (Tier 1)',
  },
  {
    id: 'test_master' as const,
    nameKey: 'settings.mode.testMaster',
    nameDefault: 'Test Master Mode',
    descKey: 'settings.mode.testMasterDesc',
    descDefault: 'Network testing (Tier 2)',
  },
] as const;

export function ModeSection({
  mode,
  onModeChange,
  className,
}: ModeSectionProps): React.JSX.Element {
  const { t } = useTranslation();

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Monitor className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings:mode.title', 'Mode')}</span>
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
                {t(modeOption.nameKey, modeOption.nameDefault)}
              </div>
              <div className="caption text-text-muted">
                {t(modeOption.descKey, modeOption.descDefault)}
              </div>
            </div>
          </label>
        ))}
      </div>
    </CollapsibleSection>
  );
}

export default ModeSection;
