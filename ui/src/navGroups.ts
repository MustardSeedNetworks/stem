/**
 * Sidebar navigation for The Stem.
 *
 * Reflector and History are top-level items (no group header). Tests
 * remain grouped under a single 'Tests' header (one entry per module).
 * The previous singleton 'Mode' group was removed in #66; the role
 * selector now lives in the header RoleChip.
 *
 * Item labels resolve from the same pages.{i18nKey}.label keys the page
 * registry uses, so the rail and the page header cannot disagree and a
 * translator sees one canonical label per route. Most stem modules are
 * glossary terms and read the same in every locale (Reflector, Benchmark,
 * ServiceTest, TrafficGen, Measure, Certify); History and Security are
 * ordinary words and do translate — which is why they were the two the
 * rail used to get wrong.
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
import { useTranslation } from 'react-i18next';
import type { SidebarNavGroup } from './ui/Sidebar';

export function useNavGroups(): SidebarNavGroup[] {
  const { t } = useTranslation('pages');
  const { t: tCommon } = useTranslation('common');
  return [
    {
      label: '',
      items: [{ path: '/reflector', label: t('reflector.label'), icon: Repeat }],
    },
    {
      label: tCommon('sections.modules'),
      items: [
        { path: '/tests/benchmark', label: t('benchmark.label'), icon: BarChart3 },
        { path: '/tests/servicetest', label: t('serviceTest.label'), icon: Settings2 },
        { path: '/tests/trafficgen', label: t('trafficGen.label'), icon: Zap },
        { path: '/tests/measure', label: t('measure.label'), icon: Waves },
        { path: '/tests/certify', label: t('certify.label'), icon: Award },
      ],
    },
    {
      label: '',
      items: [{ path: '/history', label: t('history.label'), icon: History }],
    },
    {
      label: tCommon('sections.account'),
      items: [{ path: '/account/security', label: t('accountSecurity.label'), icon: ShieldCheck }],
    },
  ];
}
