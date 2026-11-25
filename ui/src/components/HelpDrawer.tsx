/**
 * @fileoverview Stem - Help Drawer Component
 * @description Comprehensive help panel with tests, tutorials, and glossary.
 *              Supports both technical and layman-friendly explanations.
 *
 *              The frame is the shared RightDrawer and the tabs are the shared
 *              SettingsTabs. The drawer orchestrates tab/search/mode state and
 *              delegates rendering of each tab and detail view to focused
 *              components under ./help-drawer/. Framing/chrome strings are
 *              localized via the `help` i18n namespace. Short usage and
 *              glossary copy are translated; technical descriptions and
 *              tutorials remain English.
 */

import {
  Activity,
  Book,
  BookOpen,
  Clock,
  Cpu,
  GraduationCap,
  Radio,
  Settings2,
  Zap,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { getGlossary } from '../data/help/glossary';
import { searchTests, tests } from '../data/help/tests';
import type { GlossaryId, TestHelp, TestHelpId, Tutorial } from '../data/help/types';
import { useBuildVersion } from '../hooks/useBuildVersion';
import { RightDrawer } from '../shared/RightDrawer';
import { SearchInput } from '../shared/SearchInput';
import { SegmentedControl } from '../shared/SegmentedControl';
import { SettingsTabs } from '../shared/SettingsTabs';
import { Tooltip } from '../shared/Tooltip';
import { GlossaryTab } from './help-drawer/GlossaryTab';
import { TestDetailView } from './help-drawer/TestDetailView';
import { TestsTab } from './help-drawer/TestsTab';
import { TutorialDetailView } from './help-drawer/TutorialDetailView';
import { TutorialsTab } from './help-drawer/TutorialsTab';

export type HelpTab = 'tests' | 'tutorials' | 'glossary';

interface HelpDrawerProps {
  isOpen: boolean;
  initialTab?: HelpTab;
  initialTestId?: TestHelpId;
  initialGlossaryId?: GlossaryId;
  onClose: () => void;
}

export function HelpDrawer({
  isOpen,
  onClose,
  initialTab = 'tests',
  initialTestId,
  initialGlossaryId,
}: HelpDrawerProps): ReactElement | null {
  const { t } = useTranslation('help');
  const [activeTab, setActiveTab] = useState<HelpTab>(initialTab);
  const [searchQuery, setSearchQuery] = useState<string>(() =>
    initialGlossaryId ? t(`glossary.entries.${initialGlossaryId}.term`) : '',
  );
  const [simpleMode, setSimpleMode] = useState(true);
  const [selectedTest, setSelectedTest] = useState<TestHelp | null>(() =>
    initialTestId ? (tests[initialTestId] ?? null) : null,
  );
  const [selectedTutorial, setSelectedTutorial] = useState<Tutorial | null>(null);
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);
  const buildVersion = useBuildVersion();
  const searchRef = useRef<HTMLInputElement>(null);
  // Runs after the drawer's own focus-in, so the search box wins: help is
  // looked up, not read, and typing should start at once.
  useEffect(() => {
    if (isOpen) searchRef.current?.focus();
  }, [isOpen]);

  const copyCommand = (command: string): void => {
    navigator.clipboard.writeText(command);
    setCopiedCommand(command);
    setTimeout((): void => setCopiedCommand(null), 2000);
  };

  const getCategoryIcon = (categoryId: string): ReactElement => {
    switch (categoryId) {
      case 'rfc2544':
        return <Zap className="w-4 h-4" />;
      case 'y1564':
        return <Activity className="w-4 h-4" />;
      case 'rfc2889':
        return <Cpu className="w-4 h-4" />;
      case 'rfc6349':
        return <Activity className="w-4 h-4" />;
      case 'y1731':
        return <Radio className="w-4 h-4" />;
      case 'mef':
        return <Settings2 className="w-4 h-4" />;
      case 'tsn':
        return <Clock className="w-4 h-4" />;
      default:
        return <Book className="w-4 h-4" />;
    }
  };

  // Filter tests and glossary based on search
  const matchingTests = searchQuery ? new Set(searchTests(searchQuery)) : null;
  const filteredTests = searchQuery
    ? Object.values(tests).filter(
        (test) =>
          matchingTests?.has(test) ||
          t(`tests.${test.id}.whenToUse`)
            .toLocaleLowerCase()
            .includes(searchQuery.toLocaleLowerCase()),
      )
    : null;
  const glossaryEntries = getGlossary(t);
  const filteredGlossary = searchQuery
    ? glossaryEntries.filter((entry) =>
        [entry.term, entry.fullName, entry.techDef, entry.laymanDef].some((value) =>
          value.toLocaleLowerCase().includes(searchQuery.toLocaleLowerCase()),
        ),
      )
    : glossaryEntries;
  const searchTabLabel = t(`tabs.${activeTab}`);

  const tabs = [
    { id: 'tests', label: t('tabs.tests'), icon: <Book className="h-4 w-4" /> },
    { id: 'tutorials', label: t('tabs.tutorials'), icon: <GraduationCap className="h-4 w-4" /> },
    { id: 'glossary', label: t('tabs.glossary'), icon: <BookOpen className="h-4 w-4" /> },
  ];

  return (
    <RightDrawer
      open={isOpen}
      onClose={onClose}
      title={t('drawer.title')}
      closeLabel={t('drawer.close')}
      data-testid="help-drawer"
      footer={
        <Tooltip
          text={t('drawer.buildInfo', {
            commit: buildVersion.commit,
            time: buildVersion.buildTime,
          })}
        >
          <button
            type="button"
            className="text-xs text-text-muted"
            data-testid="help-drawer-version"
          >
            {t('drawer.version', { version: buildVersion.version.replace(/^v/, '') })}
          </button>
        </Tooltip>
      }
    >
      <SettingsTabs
        tabs={tabs}
        value={activeTab}
        onChange={(id): void => setActiveTab(id as HelpTab)}
        label={t('drawer.ariaLabel')}
      >
        <div className="mb-4 flex flex-wrap items-center gap-2">
          <SearchInput
            ref={searchRef}
            value={searchQuery}
            onValueChange={setSearchQuery}
            label={t('drawer.searchPlaceholder', { tab: searchTabLabel })}
            clearLabel={t('drawer.clearSearch')}
            placeholder={t('drawer.searchPlaceholder', { tab: searchTabLabel })}
            className="min-w-0 flex-1"
          />
          <SegmentedControl
            label={t('mode.label')}
            value={simpleMode ? 'simple' : 'technical'}
            onChange={(value): void => setSimpleMode(value === 'simple')}
            options={[
              { value: 'simple', label: t('mode.simple') },
              { value: 'technical', label: t('mode.technical') },
            ]}
          />
        </div>

        {/* Test Detail View */}
        {selectedTest && activeTab === 'tests' && (
          <TestDetailView
            test={selectedTest}
            simpleMode={simpleMode}
            onBack={(): void => setSelectedTest(null)}
            onCopy={copyCommand}
            copiedCommand={copiedCommand}
            onSelectTest={setSelectedTest}
          />
        )}

        {/* Tutorial Detail View */}
        {selectedTutorial && activeTab === 'tutorials' && (
          <TutorialDetailView
            tutorial={selectedTutorial}
            onBack={(): void => setSelectedTutorial(null)}
            onCopy={copyCommand}
            copiedCommand={copiedCommand}
          />
        )}

        {/* Tests Tab */}
        {activeTab === 'tests' && !selectedTest && (
          <TestsTab
            filteredTests={filteredTests}
            simpleMode={simpleMode}
            onSelectTest={setSelectedTest}
            getCategoryIcon={getCategoryIcon}
          />
        )}

        {/* Tutorials Tab */}
        {activeTab === 'tutorials' && !selectedTutorial && (
          <TutorialsTab searchQuery={searchQuery} onSelectTutorial={setSelectedTutorial} />
        )}

        {/* Glossary Tab */}
        {activeTab === 'glossary' && (
          <GlossaryTab
            searchQuery={searchQuery}
            filteredGlossary={filteredGlossary}
            simpleMode={simpleMode}
          />
        )}
      </SettingsTabs>
    </RightDrawer>
  );
}

export default HelpDrawer;
