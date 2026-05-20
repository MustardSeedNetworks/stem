/**
 * Page registry — declarative route table for The Stem.
 *
 * Reflector is the default landing route. Heavy test config forms are
 * lazy-loaded so the initial chunk only carries the reflector view.
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

export interface PageConfig {
  path: string;
  label: string;
  title: string;
  description: string;
  icon: LucideIcon;
  component: FC;
}

export const pages: PageConfig[] = [
  {
    path: '/reflector',
    label: 'Reflector',
    title: 'Reflector',
    description: 'Loopback reflector — bounces frames back for end-to-end measurement.',
    icon: Repeat,
    component: ReflectorPage,
  },
  {
    path: '/tests/benchmark',
    label: 'Benchmark',
    title: 'Benchmark',
    description: 'RFC 2544 throughput, latency, frame-loss, and back-to-back tests.',
    icon: BarChart3,
    component: BenchmarkPage,
  },
  {
    path: '/tests/servicetest',
    label: 'ServiceTest',
    title: 'ServiceTest',
    description: 'Y.1564 / MEF service activation and performance verification.',
    icon: Settings2,
    component: ServiceTestPage,
  },
  {
    path: '/tests/trafficgen',
    label: 'TrafficGen',
    title: 'TrafficGen',
    description: 'Custom traffic generation — shape streams and drive load.',
    icon: Zap,
    component: TrafficGenPage,
  },
  {
    path: '/tests/measure',
    label: 'Measure',
    title: 'Measure',
    description: 'Y.1731 OAM delay, loss, and synthetic loss measurement.',
    icon: Waves,
    component: MeasurePage,
  },
  {
    path: '/tests/certify',
    label: 'Certify',
    title: 'Certify',
    description: 'RFC 2889, RFC 6349, and TSN certification workflows.',
    icon: Award,
    component: CertifyPage,
  },
  {
    path: '/history',
    label: 'History',
    title: 'History',
    description: 'Latest test result snapshot.',
    icon: History,
    component: HistoryPage,
  },
  {
    path: '/account/security',
    label: 'Security',
    title: 'Account Security',
    description: 'Manage two-factor authentication and passkeys.',
    icon: ShieldCheck,
    component: SecurityPage,
  },
];
