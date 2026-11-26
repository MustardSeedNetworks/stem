/**
 * useRoleSwitch — the rail's one product control: Reflector or Test Master.
 *
 * It returns the props the shared RailFooter's product slot takes (a menu of
 * the two roles) and the overlay that goes with it. Choosing the other role
 * opens a ConfirmDialog explaining the consequences (an active reflector or
 * in-progress test is cancelled). On confirm the role flips via RoleContext,
 * which POSTs /api/v1/mode and updates local state only after the backend
 * accepts the change. If the POST fails, the overlay shows an Alert with a
 * dismiss control. Translation keys live under role.switchError.*.
 */
import { Repeat, Target } from 'lucide-react';
import { type ReactElement, useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { type StemRole, useRole } from '../contexts/RoleContext';
import { Alert } from '../shared/Alert';
import { ConfirmDialog } from '../shared/ConfirmDialog';
import type { RailProductSlotProps } from '../shared/RailProductSlot';

export function useRoleSwitch(): { productSlot: RailProductSlotProps; overlay: ReactElement } {
  const { t } = useTranslation();
  const { role, setRole, isSwitchingRole, roleSwitchError, clearRoleSwitchError } = useRole();
  const [pendingRole, setPendingRole] = useState<StemRole | null>(null);

  const labels: Record<StemRole, string> = {
    reflector: t('role.reflector'),
    test_master: t('role.testMaster'),
  };

  const handleSelect = useCallback(
    (next: string): void => {
      // Block new switches while one is in flight: choosing mid-switch is a
      // no-op rather than a queue.
      if (next === role || isSwitchingRole) {
        return;
      }
      setPendingRole(next === 'test_master' ? 'test_master' : 'reflector');
    },
    [role, isSwitchingRole],
  );

  const handleConfirm = useCallback((): void => {
    if (pendingRole) {
      setRole(pendingRole);
    }
    setPendingRole(null);
  }, [pendingRole, setRole]);

  const toTestMaster = pendingRole === 'test_master';

  return {
    productSlot: {
      label: t('role.label'),
      value: labels[role],
      icon: role === 'reflector' ? <Repeat /> : <Target />,
      options: [
        { value: 'reflector', label: labels.reflector },
        { value: 'test_master', label: labels.test_master },
      ],
      selected: role,
      onSelect: handleSelect,
    },
    overlay: (
      <>
        {roleSwitchError !== null ? (
          <div data-testid="role-switch-error" className="mb-4">
            <Alert
              tone="error"
              title={t('role.switchError.label')}
              onDismiss={clearRoleSwitchError}
              dismissLabel={t('role.switchError.dismiss')}
            >
              {roleSwitchError.kind === 'failed'
                ? roleSwitchError.detail
                : t('role.switchError.unexpectedResponse')}
            </Alert>
          </div>
        ) : null}
        <ConfirmDialog
          open={pendingRole !== null}
          onConfirm={handleConfirm}
          onCancel={() => setPendingRole(null)}
          title={
            toTestMaster
              ? t('role.confirm.toTestMaster.title')
              : t('role.confirm.toReflector.title')
          }
          message={
            toTestMaster
              ? t('role.confirm.toTestMaster.message')
              : t('role.confirm.toReflector.message')
          }
          confirmLabel={t('role.confirm.confirmLabel')}
          cancelLabel={t('buttons.cancel')}
          closeLabel={t('buttons.close')}
          variant="primary"
        />
      </>
    ),
  };
}
