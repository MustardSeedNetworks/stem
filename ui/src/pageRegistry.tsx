/**
 * Page registry — declarative route table for Stem.
 *
 * Reflector is the default landing route. Heavy test config forms are
 * lazy-loaded so the initial chunk only carries the reflector view.
 *
 * The header a page wears is rendered centrally by AppShell from this
 * table, not by the page itself — one edit per route, and page bodies
 * can't drift from the nav label the way hand-rolled headers did. The
 * title is the nav label, so there is no separate title copy to drift.
 */
import { type ComponentType, lazy } from 'react';
import { useTranslation } from 'react-i18next';
import type { HelpTab } from './components/HelpDrawer';
// Eager — default landing.
import type { GlossaryId, TestHelpId } from './data/help/types';
import { ReflectorPage } from './pages/ReflectorPage';

const BenchmarkPage = lazy(() =>
  import('./pages/BenchmarkPage').then((m) => ({ default: m.BenchmarkPage })),
);
const ServiceTestPage = lazy(() =>
  import('./pages/ServiceTestPage').then((m) => ({ default: m.ServiceTestPage })),
);
const TrafficGenPage = lazy(() =>
  import('./pages/TrafficGenPage').then((m) => ({ default: m.TrafficGenPage })),
);
const MeasurePage = lazy(() =>
  import('./pages/MeasurePage').then((m) => ({ default: m.MeasurePage })),
);
const CertifyPage = lazy(() =>
  import('./pages/CertifyPage').then((m) => ({ default: m.CertifyPage })),
);
const HistoryPage = lazy(() =>
  import('./pages/HistoryPage').then((m) => ({ default: m.HistoryPage })),
);
const SecurityPage = lazy(() =>
  import('./pages/account/security/SecurityPage').then((m) => ({ default: m.SecurityPage })),
);

/**
 * PageConfig is one entry in the route table, resolved at render time
 * via usePages() — label/description/eyebrow are translations of
 * the corresponding pages.{i18nKey}.* keys.
 */
export interface PageConfig {
  path: string;
  help: { initialTab: HelpTab; initialTestId?: TestHelpId; initialGlossaryId?: GlossaryId };
  /** The nav label, and the page title: the same string the user clicked. */
  label: string;
  /** Kicker above the title naming the product domain. */
  eyebrow: string;
  /** The nav group the page sits under, when it has one: the middle breadcrumb. */
  section?: 'modules' | 'account';
  description: string;
  component: ComponentType;
  /** Page runs a test, so it carries the run controls under its header. */
  runControls?: boolean;
}

/**
 * PageI18nKey is the closed set of pages.* namespaces that carry a
 * matching {label,eyebrow,description} triple. Kept strict so adding a
 * new route forces a corresponding locale entry.
 */
type PageI18nKey =
  | 'reflector'
  | 'benchmark'
  | 'serviceTest'
  | 'trafficGen'
  | 'measure'
  | 'certify'
  | 'history'
  | 'accountSecurity';

/**
 * PageDef is the static, language-agnostic definition. The matching
 * translation lives at pages.{i18nKey}.{label,eyebrow,description} in
 * internal/i18n/locales/{en,es}/pages.json.
 */
interface PageDef {
  path: string;
  help: PageConfig['help'];
  i18nKey: PageI18nKey;
  section?: PageConfig['section'];
  component: ComponentType;
  runControls?: boolean;
}

const staticPages: PageDef[] = [
  {
    path: '/reflector',
    help: { initialTab: 'glossary', initialGlossaryId: 'reflector' },
    i18nKey: 'reflector',
    component: ReflectorPage,
  },
  {
    path: '/tests/benchmark',
    section: 'modules',
    help: { initialTab: 'tests', initialTestId: 'rfc2544_throughput' },
    i18nKey: 'benchmark',
    component: BenchmarkPage,
    runControls: true,
  },
  {
    path: '/tests/servicetest',
    section: 'modules',
    help: { initialTab: 'tests', initialTestId: 'y1564_config' },
    i18nKey: 'serviceTest',
    component: ServiceTestPage,
    runControls: true,
  },
  {
    path: '/tests/trafficgen',
    section: 'modules',
    help: { initialTab: 'tests', initialTestId: 'custom_stream' },
    i18nKey: 'trafficGen',
    component: TrafficGenPage,
    runControls: true,
  },
  {
    path: '/tests/measure',
    section: 'modules',
    help: { initialTab: 'tests', initialTestId: 'y1731_delay' },
    i18nKey: 'measure',
    component: MeasurePage,
    runControls: true,
  },
  {
    path: '/tests/certify',
    section: 'modules',
    help: { initialTab: 'tests', initialTestId: 'tsn' },
    i18nKey: 'certify',
    component: CertifyPage,
    runControls: true,
  },
  {
    path: '/history',
    help: { initialTab: 'glossary', initialGlossaryId: 'history' },
    i18nKey: 'history',
    component: HistoryPage,
  },
  {
    path: '/account/security',
    section: 'account',
    help: { initialTab: 'glossary', initialGlossaryId: 'security' },
    i18nKey: 'accountSecurity',
    component: SecurityPage,
  },
];

/**
 * usePages returns the route table with label/description/eyebrow
 * resolved against the active locale. A hook rather than a const so
 * react-i18next's languageChanged event re-renders consumers.
 */
export function usePages(): PageConfig[] {
  const { t } = useTranslation('pages');
  return staticPages.map((p) => ({
    path: p.path,
    help: p.help,
    label: t(`${p.i18nKey}.label`),
    eyebrow: t(`${p.i18nKey}.eyebrow`),
    section: p.section,
    description: t(`${p.i18nKey}.description`),
    component: p.component,
    runControls: p.runControls,
  }));
}
