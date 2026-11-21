import { Compass } from 'lucide-react';
import { type ReactElement, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';
import { iconSizes } from '../constants/sizes';

/**
 * Shown for any path no page owns, in place of a silent redirect that left
 * the operator on a page they never asked for with no hint the address was wrong.
 */
export function NotFoundPanel(): ReactElement {
  const { t } = useTranslation('common');
  const product = t('app.title');
  const title = t('notFound.title');

  useEffect(() => {
    document.title = `${title} | ${product}`;
  }, [title, product]);

  return (
    <section className="stack-lg" data-testid="not-found">
      <Compass className={`${iconSizes.lg} text-text-muted`} aria-hidden="true" />
      <h1 className="heading-2 text-text-primary">{title}</h1>
      <p className="text-text-muted">{t('notFound.body', { product })}</p>
      <Link
        to="/"
        data-testid="not-found-home"
        className="inline-flex min-h-11 items-center text-brand-primary hover:underline"
      >
        {t('notFound.home', { product })}
      </Link>
    </section>
  );
}
