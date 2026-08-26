/**
 * Page registry — declarative route table for The Stem.
 *
 * Reflector is the default landing route. Heavy test config forms are
 * lazy-loaded so the initial chunk only carries the reflector view.
 *
 * The header a page wears is rendered centrally by AppShell from this
 * table, not by the page itself — one edit per route, and page bodies
 * can't drift from the nav label the way hand-rolled headers did.
 */
import {
  Award,
  BarChart3,
  History,
  type LucideIcon,
  Repeat,
  Settings2,
  ShieldCheck,
  Waves,
  Zap,
} from 'lucide-react';
import { type FC, lazy } from 'react';
import { useTranslation } from 'react-i18next';

// Eager — default landing.
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
 * via usePages() — label/title/description/eyebrow are translations of
 * the corresponding pages.{i18nKey}.* keys.
 */
export interface PageConfig {
  path: string;
  label: string;
  /** Kicker above the title naming the product domain. */
  eyebrow?: string;
  title: string;
  description: string;
  icon: LucideIcon;
  iconColorClass?: string;
  component: FC;
}

/**
 * PageI18nKey is the closed set of pages.* namespaces that carry a
 * matching {label,title,description} triple. Kept strict so adding a
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
 * translation lives at pages.{i18nKey}.{label,title,description} in
 * ui/locales/{en,es}/pages.json.
 */
interface PageDef {
  path: string;
  i18nKey: PageI18nKey;
  icon: LucideIcon;
  iconColorClass?: string;
  component: FC;
}

const staticPages: PageDef[] = [
  {
    path: '/reflector',
    i18nKey: 'reflector',
    icon: Repeat,
    iconColorClass: 'text-module-reflector',
    component: ReflectorPage,
  },
  {
    path: '/tests/benchmark',
    i18nKey: 'benchmark',
    icon: BarChart3,
    iconColorClass: 'text-module-benchmark',
    component: BenchmarkPage,
  },
  {
    path: '/tests/servicetest',
    i18nKey: 'serviceTest',
    icon: Settings2,
    iconColorClass: 'text-module-servicetest',
    component: ServiceTestPage,
  },
  {
    path: '/tests/trafficgen',
    i18nKey: 'trafficGen',
    icon: Zap,
    iconColorClass: 'text-module-trafficgen',
    component: TrafficGenPage,
  },
  {
    path: '/tests/measure',
    i18nKey: 'measure',
    icon: Waves,
    iconColorClass: 'text-module-measure',
    component: MeasurePage,
  },
  {
    path: '/tests/certify',
    i18nKey: 'certify',
    icon: Award,
    iconColorClass: 'text-module-certify',
    component: CertifyPage,
  },
  { path: '/history', i18nKey: 'history', icon: History, component: HistoryPage },
  {
    path: '/account/security',
    i18nKey: 'accountSecurity',
    icon: ShieldCheck,
    component: SecurityPage,
  },
];

/**
 * usePages returns the route table with label/title/description/eyebrow
 * resolved against the active locale. A hook rather than a const so
 * react-i18next's languageChanged event re-renders consumers.
 */
export function usePages(): PageConfig[] {
  const { t } = useTranslation('pages');
  return staticPages.map((p) => ({
    path: p.path,
    label: t(`${p.i18nKey}.label`),
    // A page has an eyebrow when its locale namespace declares one, so the
    // copy lives in one place instead of being mirrored by a flag here.
    // Pages still awaiting their archetype pass have none.
    eyebrow: t(`${p.i18nKey}.eyebrow`, { defaultValue: '' }) || undefined,
    title: t(`${p.i18nKey}.title`),
    description: t(`${p.i18nKey}.description`),
    icon: p.icon,
    iconColorClass: p.iconColorClass,
    component: p.component,
  }));
}
