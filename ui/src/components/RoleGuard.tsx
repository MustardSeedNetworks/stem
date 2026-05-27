/**
 * RoleGuard — page-level banner that surfaces a role mismatch.
 *
 * Used at the top of each module page. If the current stem role does
 * not match `requires`, the children are still rendered (we don't hide
 * page content) but a banner is shown explaining the mismatch with an
 * inline action to switch roles. Clicking the action pops the same
 * ConfirmModal used by the header RoleChip.
 */
import { AlertTriangle } from 'lucide-react';
import { type FC, type ReactNode, useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { type StemRole, useRole } from '../contexts/RoleContext';
import { ConfirmModal } from './ui/ConfirmModal';

interface RoleGuardProps {
  requires: StemRole;
  /** Module display name interpolated into the banner message. */
  moduleName?: string;
  children: ReactNode;
}

export const RoleGuard: FC<RoleGuardProps> = ({ requires, moduleName, children }) => {
  const { t } = useTranslation();
  const { role, setRole } = useRole();
  const [confirmOpen, setConfirmOpen] = useState(false);

  const handleSwitchClick = useCallback((): void => {
    setConfirmOpen(true);
  }, []);

  const handleConfirm = useCallback((): void => {
    setRole(requires);
    setConfirmOpen(false);
  }, [requires, setRole]);

  const handleCancel = useCallback((): void => {
    setConfirmOpen(false);
  }, []);

  if (role === requires) {
    return <>{children}</>;
  }

  const bannerMessage =
    requires === 'test_master'
      ? t('role.guard.needTestMaster', {
          module: moduleName ?? t('role.testMaster', 'Test Master'),
          defaultValue: `This stem is currently configured as Reflector. Switch to Test Master to run ${moduleName ?? 'this module'}.`,
        })
      : t(
          'role.guard.needReflector',
          'This stem is currently configured as Test Master. Switch to Reflector to use the loopback reflector.',
        );

  const confirmTitleKey =
    requires === 'test_master'
      ? 'role.confirm.toTestMaster.title'
      : 'role.confirm.toReflector.title';
  const confirmTitleDefault =
    requires === 'test_master' ? 'Switch to Test Master?' : 'Switch to Reflector?';
  const confirmMessageKey =
    requires === 'test_master'
      ? 'role.confirm.toTestMaster.message'
      : 'role.confirm.toReflector.message';
  const confirmMessageDefault =
    requires === 'test_master'
      ? 'The current Reflector will stop. Any in-progress tests will be cancelled.'
      : 'Any in-progress test will be cancelled. This stem will start as a Reflector.';

  return (
    <>
      <div
        role="status"
        className="flex flex-wrap items-center gap-default rounded-lg border border-status-warning/30 bg-status-warning/10 px-3 py-row text-sm text-status-warning"
      >
        <AlertTriangle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
        <span className="flex-1 text-text-primary">{bannerMessage}</span>
        <button
          type="button"
          onClick={handleSwitchClick}
          className="inline-flex items-center gap-tight rounded-md border border-status-warning/40 bg-surface-raised px-2.5 py-compact text-xs font-medium text-text-primary hover:bg-surface-hover transition-colors"
        >
          {t('role.guard.switchAction', 'Switch role')}
        </button>
      </div>

      <ConfirmModal
        isOpen={confirmOpen}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
        title={t(confirmTitleKey, confirmTitleDefault)}
        message={t(confirmMessageKey, confirmMessageDefault)}
        confirmLabel={t('role.confirm.confirmLabel', 'Switch role')}
        cancelLabel={t('buttons.cancel', 'Cancel')}
        confirmTone="violet"
      />

      {children}
    </>
  );
};

export default RoleGuard;
