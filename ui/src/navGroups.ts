/**
 * Sidebar navigation groups for The Stem.
 *
 * Reflector and History are top-level items (no group header). Tests
 * remain grouped under a single 'Tests' header (one entry per module).
 * The previous singleton 'Mode' group was removed in #66; the role
 * selector now lives in the header RoleChip.
 */
import {
  Award,
  BarChart3,
  History,
  Repeat,
  Settings2,
  ShieldCheck,
  Waves,
  Zap,
} from 'lucide-react';
import type { SidebarNavGroup } from './ui/Sidebar';

export const navGroups: SidebarNavGroup[] = [
  {
    label: '',
    items: [{ path: '/reflector', label: 'Reflector', icon: Repeat }],
  },
  {
    label: 'Tests',
    items: [
      { path: '/tests/benchmark', label: 'Benchmark', icon: BarChart3 },
      { path: '/tests/servicetest', label: 'ServiceTest', icon: Settings2 },
      { path: '/tests/trafficgen', label: 'TrafficGen', icon: Zap },
      { path: '/tests/measure', label: 'Measure', icon: Waves },
      { path: '/tests/certify', label: 'Certify', icon: Award },
    ],
  },
  {
    label: '',
    items: [{ path: '/history', label: 'History', icon: History }],
  },
  {
    label: 'Account',
    items: [{ path: '/account/security', label: 'Security', icon: ShieldCheck }],
  },
];
