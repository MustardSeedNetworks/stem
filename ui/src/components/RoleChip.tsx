/**
 * RoleChip — persistent header control for switching the stem role.
 *
 * Segmented control with two values: Reflector / Test Master. Clicking
 * the inactive value pops a ConfirmModal explaining the consequences
 * (active reflector or in-progress test will be cancelled). On confirm
 * the role flips via RoleContext, which POSTs /api/v1/mode and
 * updates local state only after the backend accepts the change.
 *
 * While the POST is in flight the chip shows an inline spinner. If
 * the POST fails, an inline error tag renders below the chip with a
 * dismiss control. Translation keys live under role.switchError.*.
 */
import { Loader2, Repeat, Target, X } from 'lucide-react';
import { type FC, useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { type StemRole, useRole } from '../contexts/RoleContext';
import { ConfirmModal } from './ui/ConfirmModal';

interface RoleChipProps {
  className?: string;
}

interface RoleOption {
  id: StemRole;
  labelKey: string;
  labelDefault: string;
  icon: typeof Repeat;
}

const ROLE_OPTIONS: readonly RoleOption[] = [
  {
    id: 'reflector',
    labelKey: 'role.reflector',
    labelDefault: 'Reflector',
    icon: Repeat,
  },
  {
    id: 'test_master',
    labelKey: 'role.testMaster',
    labelDefault: 'Test Master',
    icon: Target,
  },
];

export const RoleChip: FC<RoleChipProps> = ({ className = '' }) => {
  const { t } = useTranslation();
  const { role, setRole, isSwitchingRole, roleSwitchError, clearRoleSwitchError } = useRole();
  const [pendingRole, setPendingRole] = useState<StemRole | null>(null);

  const handleClick = useCallback(
    (next: StemRole): void => {
      if (next === role) {
        return;
      }
      // Block new switches while one is in flight. Clicking the
      // other option mid-spin is a no-op rather than queueing.
      if (isSwitchingRole) {
        return;
      }
      setPendingRole(next);
    },
    [role, isSwitchingRole],
  );

  const handleConfirm = useCallback((): void => {
    if (pendingRole) {
      setRole(pendingRole);
    }
    setPendingRole(null);
  }, [pendingRole, setRole]);

  const handleCancel = useCallback((): void => {
    setPendingRole(null);
  }, []);

  const confirmTitleKey =
    pendingRole === 'test_master'
      ? 'role.confirm.toTestMaster.title'
      : 'role.confirm.toReflector.title';
  const confirmTitleDefault =
    pendingRole === 'test_master' ? 'Switch to Test Master?' : 'Switch to Reflector?';
  const confirmMessageKey =
    pendingRole === 'test_master'
      ? 'role.confirm.toTestMaster.message'
      : 'role.confirm.toReflector.message';
  const confirmMessageDefault =
    pendingRole === 'test_master'
      ? 'The current Reflector will stop. Any in-progress tests will be cancelled.'
      : 'Any in-progress test will be cancelled. This stem will start as a Reflector.';

  return (
    <div className={`inline-flex flex-col items-start gap-tight ${className}`}>
      <fieldset
        className="inline-flex items-center gap-0 rounded-lg border border-surface-border bg-surface-raised p-0.5"
        aria-label={t('role.label', 'Stem role')}
        aria-busy={isSwitchingRole}
      >
        <legend className="sr-only">{t('role.label', 'Stem role')}</legend>
        {ROLE_OPTIONS.map((option) => {
          const Icon = option.icon;
          const active = role === option.id;
          return (
            <button
              key={option.id}
              type="button"
              onClick={() => handleClick(option.id)}
              disabled={isSwitchingRole}
              className={`inline-flex items-center gap-1.5 px-2.5 py-compact rounded-md text-xs font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary disabled:opacity-60 disabled:cursor-not-allowed ${
                active
                  ? 'bg-brand-primary text-on-brand shadow-sm'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover'
              }`}
              aria-pressed={active}
              data-testid={`role-chip-${option.id}`}
            >
              {active && isSwitchingRole ? (
                <Loader2
                  className="h-3.5 w-3.5 animate-spin"
                  aria-hidden="true"
                  data-testid="role-chip-spinner"
                />
              ) : (
                <Icon className="h-3.5 w-3.5" aria-hidden="true" />
              )}
              <span>{t(option.labelKey, option.labelDefault)}</span>
            </button>
          );
        })}
      </fieldset>

      {roleSwitchError !== null && roleSwitchError.length > 0 ? (
        <div
          role="alert"
          data-testid="role-chip-error"
          className="inline-flex items-center gap-compact rounded-md border border-status-error/40 bg-status-error/10 px-cell py-compact text-xs text-status-error"
        >
          <span className="font-medium">{t('role.switchError.label', 'Role switch failed:')}</span>
          <span className="font-normal text-text-primary">{roleSwitchError}</span>
          <button
            type="button"
            onClick={clearRoleSwitchError}
            aria-label={t('role.switchError.dismiss', 'Dismiss')}
            className="inline-flex h-4 w-4 items-center justify-center rounded text-status-error/80 hover:bg-status-error/20 hover:text-status-error focus:outline-none focus-visible:ring-2 focus-visible:ring-status-error"
            data-testid="role-chip-error-dismiss"
          >
            <X className="h-3 w-3" aria-hidden="true" />
          </button>
        </div>
      ) : null}

      <ConfirmModal
        isOpen={pendingRole !== null}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
        title={t(confirmTitleKey, confirmTitleDefault)}
        message={t(confirmMessageKey, confirmMessageDefault)}
        confirmLabel={t('role.confirm.confirmLabel', 'Switch role')}
        cancelLabel={t('buttons.cancel', 'Cancel')}
        confirmTone="violet"
      />
    </div>
  );
};

export default RoleChip;
