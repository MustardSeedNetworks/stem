/**
 * PageActions — a page puts its own actions in the shared PageHeader.
 *
 * The header is rendered once by the shell, but what a page can do (refresh the
 * data it loaded) is the page's business. The shell holds the header's actions
 * element and provides it here; a page renders `<PageActions>` and its children
 * appear in the header, in the page's own React tree, with its own state.
 */

import { RefreshCw } from 'lucide-react';
import { createContext, type ReactNode, useContext } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import { Button } from '../shared/Button';

export const PageActionsSlotContext = createContext<HTMLElement | null>(null);

export function PageActions({ children }: { children: ReactNode }) {
  const slot = useContext(PageActionsSlotContext);
  return slot ? createPortal(children, slot) : null;
}

/** The header's Refresh: reload what this page loaded. */
export function RefreshAction({ onRefresh, title }: { onRefresh: () => void; title?: string }) {
  const { t } = useTranslation();
  return (
    <PageActions>
      <Button
        variant="secondary"
        size="sm"
        icon={<RefreshCw className="h-4 w-4" aria-hidden="true" />}
        onClick={onRefresh}
        data-testid="page-refresh"
        {...(title ? { title } : {})}
      >
        {t('labels.refresh')}
      </Button>
    </PageActions>
  );
}
