import { useTranslation } from 'react-i18next';
import type { DataTableLabels } from '../shared/DataTable';

/**
 * The DataTable's translated chrome (pager, select-all, loading), so a page
 * passes `useTableLabels()` and names only its own rows.
 */
export function useTableLabels<T>(rowName?: (row: T) => string): DataTableLabels<T> {
  const { t } = useTranslation('common');
  return {
    selectAll: t('table.selectAll'),
    selectRow: (row) => t('table.selectRow', { name: rowName ? rowName(row) : '' }),
    actions: t('table.actions'),
    pagination: t('table.pagination'),
    previousPage: t('table.previousPage'),
    nextPage: t('table.nextPage'),
    loading: t('table.loading'),
    pageStatus: (page, pageCount, total) => t('table.pageStatus', { page, pageCount, total }),
  };
}
