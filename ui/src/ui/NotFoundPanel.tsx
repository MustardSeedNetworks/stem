import { Compass } from 'lucide-react';
import { type ReactElement, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';
import { documentTitle } from '../shared/documentTitle';
import { EmptyState } from '../shared/EmptyState';
import { RoutedPageHeader } from './RoutedPageHeader';

/**
 * Shown for any path no page owns, in place of a silent redirect that left
 * the operator on a page they never asked for with no hint the address was wrong.
 */
export function NotFoundPanel(): ReactElement {
  const { t } = useTranslation('common');
  const product = t('app.title');
  const title = t('notFound.title');

  useEffect(() => {
    document.title = documentTitle(title, product);
  }, [title, product]);

  return (
    <section data-testid="not-found">
      <RoutedPageHeader kicker={t('notFound.kicker')} title={title} />
      <EmptyState
        icon={<Compass className="h-6 w-6" />}
        title={t('notFound.body', { product })}
        action={
          <Link
            to="/"
            data-testid="not-found-home"
            className="inline-flex min-h-11 items-center text-brand-primary hover:underline"
          >
            {t('notFound.home', { product })}
          </Link>
        }
      />
    </section>
  );
}
