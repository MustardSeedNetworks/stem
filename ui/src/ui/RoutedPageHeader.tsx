/**
 * RoutedPageHeader — the shared PageHeader wired to stem's router, copy and
 * help drawer. Every routed page, and the not-found page, opens with this: the
 * trail is the product, the nav group (when the page has one) and the page.
 */
import type { ReactElement, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';
import type { PageConfig } from '../pageRegistry';
import type { BreadcrumbItem, BreadcrumbLinkProps } from '../shared/Breadcrumbs';
import { PageHeader } from '../shared/PageHeader';

function renderLink({ href, className, children }: BreadcrumbLinkProps): ReactNode {
  return (
    <Link to={href} className={className}>
      {children}
    </Link>
  );
}

interface RoutedPageHeaderProps {
  kicker: string;
  title: string;
  section?: PageConfig['section'];
  subtitle?: string;
  /** Opens the help drawer; omit on a page with no help content. */
  onHelp?: () => void;
}

export function RoutedPageHeader({
  kicker,
  title,
  section,
  subtitle,
  onHelp,
}: RoutedPageHeaderProps): ReactElement {
  const { t } = useTranslation('common');
  const { t: tHelp } = useTranslation('help');
  const trail: BreadcrumbItem[] = [{ label: t('app.title'), href: '/' }];
  if (section) {
    trail.push({ label: t(`sections.${section}`) });
  }
  trail.push({ label: title });
  // WebKit does not focus a button on click, and the drawer returns focus to
  // whatever was focused when it opened. Focus the help button first so Escape
  // lands back on it in every engine.
  const openHelp = onHelp
    ? (): void => {
        document.querySelector<HTMLElement>('[data-testid="page-header-help-button"]')?.focus();
        onHelp();
      }
    : undefined;
  return (
    <PageHeader
      kicker={kicker}
      title={title}
      breadcrumbs={trail}
      breadcrumbsLabel={t('accessibility.breadcrumb')}
      renderBreadcrumbLink={renderLink}
      {...(subtitle ? { subtitle } : {})}
      {...(onHelp
        ? { help: { label: tHelp('drawer.pageHelp', { page: title }), onClick: onHelp } }
        : {})}
    />
  );
}
