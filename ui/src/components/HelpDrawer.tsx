/**
 * @fileoverview The Stem - Help Drawer Component
 * @description Comprehensive help panel with tests, tutorials, and glossary.
 *              Supports both technical and layman-friendly explanations.
 *
 *              The drawer orchestrates tab/search/mode state and delegates
 *              rendering of each tab and detail view to focused components
 *              under ./help-drawer/. Framing/chrome strings are localized via
 *              the `help` i18n namespace; the help content corpus itself
 *              (per-test descriptions, glossary entries, tutorial bodies)
 *              stays English in the data/help corpus.
 */

import {
  Activity,
  Book,
  BookOpen,
  Clock,
  Cpu,
  GraduationCap,
  Radio,
  Search,
  Settings2,
  X,
  Zap,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { searchGlossary } from '../data/help/glossary';
import { searchTests } from '../data/help/tests';
import type { TestHelp, Tutorial } from '../data/help/types';
import { useBuildVersion } from '../hooks/useBuildVersion';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { cn, icon as iconTokens, layout, modal, radius, spacing, status } from '../styles/theme';
import { GlossaryTab } from './help-drawer/GlossaryTab';
import { TestDetailView } from './help-drawer/TestDetailView';
import { TestsTab } from './help-drawer/TestsTab';
import { TutorialDetailView } from './help-drawer/TutorialDetailView';
import { TutorialsTab } from './help-drawer/TutorialsTab';

type Tab = 'tests' | 'tutorials' | 'glossary';

interface HelpDrawerProps {
  isOpen: boolean;
  onClose: () => void;
}

export function HelpDrawer({ isOpen, onClose }: HelpDrawerProps): ReactElement | null {
  const { t } = useTranslation('help');
  const [activeTab, setActiveTab] = useState<Tab>('tests');
  const [searchQuery, setSearchQuery] = useState('');
  const [simpleMode, setSimpleMode] = useState(true);
  const [selectedTest, setSelectedTest] = useState<TestHelp | null>(null);
  const [selectedTutorial, setSelectedTutorial] = useState<Tutorial | null>(null);
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);
  const buildVersion = useBuildVersion();
  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: isOpen,
    onEscape: onClose,
  });

  if (!isOpen) {
    return null;
  }

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
  const filteredTests = searchQuery ? searchTests(searchQuery) : null;
  const filteredGlossary = searchQuery ? searchGlossary(searchQuery) : null;
  const searchTabLabel = t(`tabs.${activeTab}`);

  return (
    <>
      {/* Backdrop */}
      <div className={modal.backdrop} onClick={onClose} aria-hidden="true" />

      {/* Drawer */}
      <div
        ref={drawerRef}
        data-testid="help-drawer"
        role="dialog"
        aria-modal="true"
        aria-label={t('drawer.ariaLabel')}
        className={cn(
          'fixed right-0 top-0 h-full w-[520px] max-w-full z-50',
          layout.flex.col,
          'bg-surface-raised border-l border-surface-border shadow-xl',
        )}
      >
        {/* Header */}
        <div
          className={cn(
            'sticky top-0 bg-surface-raised border-b border-surface-border shrink-0',
            spacing.pad.default,
          )}
        >
          <div className={cn(layout.flex.between, spacing.margin.bottom.heading)}>
            <div className={cn(layout.inline.default)}>
              <BookOpen className={cn(iconTokens.size.md, 'text-brand-primary')} />
              <div>
                <h2 className="heading-3">{t('drawer.title')}</h2>
                <p
                  className="caption text-text-muted"
                  data-testid="help-drawer-version"
                  title={`commit ${buildVersion.commit} · built ${buildVersion.buildTime}`}
                >
                  Stem v{buildVersion.version}
                </p>
              </div>
            </div>
            <button
              type="button"
              data-testid="help-drawer-close"
              onClick={onClose}
              className={cn(
                'pad-xs text-text-muted hover:text-text-primary transition-colors',
                radius.lg,
                'hover:bg-surface-hover',
              )}
              title={t('drawer.closeTooltip')}
              aria-label={t('drawer.close')}
            >
              <X className={iconTokens.size.md} aria-hidden="true" />
            </button>
          </div>

          {/* Tabs */}
          <div className={cn('flex gap-tight bg-surface-base p-1', radius.lg)}>
            <button
              type="button"
              onClick={(): void => setActiveTab('tests')}
              title={t('tabs.testsTooltip')}
              data-testid="help-drawer-tab-tests"
              className={cn(
                'flex-1 px-3 py-row text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'tests'
                  ? 'bg-brand-primary text-on-brand'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <Book className={iconTokens.size.sm} />
              {t('tabs.tests')}
            </button>
            <button
              type="button"
              onClick={(): void => setActiveTab('tutorials')}
              title={t('tabs.tutorialsTooltip')}
              data-testid="help-drawer-tab-tutorials"
              className={cn(
                'flex-1 px-3 py-row text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'tutorials'
                  ? 'bg-brand-primary text-on-brand'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <GraduationCap className={iconTokens.size.sm} />
              {t('tabs.tutorials')}
            </button>
            <button
              type="button"
              onClick={(): void => setActiveTab('glossary')}
              title={t('tabs.glossaryTooltip')}
              data-testid="help-drawer-tab-glossary"
              className={cn(
                'flex-1 px-3 py-row text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'glossary'
                  ? 'bg-brand-primary text-on-brand'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <BookOpen className={iconTokens.size.sm} />
              {t('tabs.glossary')}
            </button>
          </div>

          {/* Search and Mode Toggle */}
          <div className={cn(layout.inline.default, spacing.margin.top.heading)}>
            <div className="flex-1 relative">
              <Search
                className={cn(
                  'absolute left-3 top-1/2 -translate-y-1/2',
                  iconTokens.size.sm,
                  'text-text-muted',
                )}
              />
              <input
                type="text"
                placeholder={t('drawer.searchPlaceholder', { tab: searchTabLabel })}
                value={searchQuery}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  setSearchQuery(e.target.value)
                }
                className={cn(
                  'w-full pl-9 pr-3 py-row body-small',
                  'bg-surface-base border border-surface-border text-text-primary',
                  'placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-brand-primary',
                  radius.lg,
                )}
              />
            </div>
            <button
              type="button"
              onClick={(): void => setSimpleMode(!simpleMode)}
              className={cn(
                'px-3 py-row text-xs font-medium transition-colors',
                radius.lg,
                simpleMode
                  ? cn(status.bg.info, 'text-text-inverse')
                  : 'bg-surface-base text-text-muted border border-surface-border hover:bg-surface-hover',
              )}
              title={simpleMode ? t('mode.toggleToTechnical') : t('mode.toggleToSimple')}
              aria-label={simpleMode ? t('mode.ariaToTechnical') : t('mode.ariaToSimple')}
            >
              {simpleMode ? t('mode.simple') : t('mode.technical')}
            </button>
          </div>
        </div>

        {/* Content */}
        <div className={cn('flex-1 overflow-y-auto', spacing.pad.default)}>
          {/* Test Detail View */}
          {selectedTest && activeTab === 'tests' && (
            <TestDetailView
              test={selectedTest}
              simpleMode={simpleMode}
              onBack={(): void => setSelectedTest(null)}
              onCopy={copyCommand}
              copiedCommand={copiedCommand}
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
        </div>
      </div>
    </>
  );
}

export default HelpDrawer;
