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
import { useShellStore } from '../stores/shell-store';

export interface ModuleEmptyStateProps {
  /** Module display name, as the sidebar and RoleGuard name it. */
  moduleName: string;
}

export function ModuleEmptyState({ moduleName }: ModuleEmptyStateProps): ReactElement {
  const { t } = useTranslation('common');
  const setSettingsOpen = useShellStore((s) => s.setSettingsOpen);

  return (
    <div className="card" data-testid="module-empty-state">
      <div className="flex flex-col items-center gap-default py-centered text-center">
        <SlidersHorizontal className="h-6 w-6 text-text-muted" aria-hidden="true" />
        <p className="text-sm font-medium text-text-primary">
          {t('moduleEmptyState.title', { module: moduleName })}
        </p>
        <p className="text-sm text-text-muted">{t('moduleEmptyState.body')}</p>
        <button
          type="button"
          data-testid="module-empty-state-open-settings"
          onClick={() => setSettingsOpen(true)}
          className="inline-flex items-center gap-tight rounded-md border border-border-muted bg-surface-raised px-3 py-compact text-sm font-medium text-text-primary hover:bg-surface-hover transition-colors"
        >
          {t('moduleEmptyState.openSettings')}
        </button>
      </div>
    </div>
  );
}
