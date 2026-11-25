/**
 * ModuleEmptyState — what a module page shows when none of its tests are
 * selected.
 *
 * `selectedTests` starts as the four RFC 2544 ids, so every other module page
 * opened with each of its ConfigForms returning null and the operator facing a
 * title and nothing else (#1257). The page says which module it is and offers
 * the one action that resolves it; opening the drawer goes through the shell
 * store so the lazily loaded drawer from #1150 stays lazy.
 */
import { SlidersHorizontal } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../shared/Button';
import { EmptyState } from '../shared/EmptyState';
import { useShellStore } from '../stores/shell-store';

export interface ModuleEmptyStateProps {
  /** Module display name, as the sidebar and RoleGuard name it. */
  moduleName: string;
}

export function ModuleEmptyState({ moduleName }: ModuleEmptyStateProps): ReactElement {
  const { t } = useTranslation('common');
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);

  return (
    <div data-testid="module-empty-state">
      <EmptyState
        icon={<SlidersHorizontal className="h-6 w-6" />}
        title={t('moduleEmptyState.title', { module: moduleName })}
        hint={t('moduleEmptyState.body')}
        action={
          <Button
            variant="secondary"
            data-testid="module-empty-state-open-settings"
            onClick={() => setSettingsOpen(true)}
          >
            {t('moduleEmptyState.openSettings')}
          </Button>
        }
      />
    </div>
  );
}
