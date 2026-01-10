/**
 * @fileoverview The Stem - Help Drawer Component
 * @description Comprehensive help panel with tests, tutorials, and glossary.
 *              Supports both technical and layman-friendly explanations.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  Activity,
  Book,
  BookOpen,
  Check,
  ChevronRight,
  Clock,
  Copy,
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
import {
  categories,
  type GlossaryEntry,
  getTestsByCategory,
  glossary,
  searchGlossary,
  searchTests,
  type TestHelp,
  type Tutorial,
  tutorials,
} from '../data/helpContent';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { cn, icon as iconTokens, layout, modal, radius, spacing } from '../styles/theme';
import { CollapsibleSection } from './CollapsibleSection';

type Tab = 'tests' | 'tutorials' | 'glossary';

interface HelpDrawerProps {
  isOpen: boolean;
  onClose: () => void;
}

export function HelpDrawer({ isOpen, onClose }: HelpDrawerProps) {
  const [activeTab, setActiveTab] = useState<Tab>('tests');
  const [searchQuery, setSearchQuery] = useState('');
  const [simpleMode, setSimpleMode] = useState(true);
  const [selectedTest, setSelectedTest] = useState<TestHelp | null>(null);
  const [selectedTutorial, setSelectedTutorial] = useState<Tutorial | null>(null);
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);
  const drawerRef = useFocusTrap<HTMLDivElement>({
    isActive: isOpen,
    onEscape: onClose,
  });

  if (!isOpen) return null;

  const copyCommand = (command: string) => {
    navigator.clipboard.writeText(command);
    setCopiedCommand(command);
    setTimeout(() => setCopiedCommand(null), 2000);
  };

  const getCategoryIcon = (categoryId: string) => {
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

  return (
    <>
      {/* Backdrop */}
      <div className={modal.backdrop} onClick={onClose} aria-hidden="true" />

      {/* Drawer */}
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Help & Documentation"
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
              <h2 className="heading-3">Help & Documentation</h2>
            </div>
            <button
              type="button"
              onClick={onClose}
              className={cn(
                'p-2 text-text-muted hover:text-text-primary transition-colors',
                radius.lg,
                'hover:bg-surface-hover',
              )}
            >
              <X className={iconTokens.size.md} />
            </button>
          </div>

          {/* Tabs */}
          <div className={cn('flex gap-1 bg-surface-base p-1', radius.lg)}>
            <button
              type="button"
              onClick={() => setActiveTab('tests')}
              className={cn(
                'flex-1 px-3 py-2 text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'tests'
                  ? 'bg-brand-primary text-text-inverse'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <Book className={iconTokens.size.sm} />
              Tests
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('tutorials')}
              className={cn(
                'flex-1 px-3 py-2 text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'tutorials'
                  ? 'bg-brand-primary text-text-inverse'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <GraduationCap className={iconTokens.size.sm} />
              Tutorials
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('glossary')}
              className={cn(
                'flex-1 px-3 py-2 text-sm font-medium transition-colors',
                radius.md,
                layout.inline.default,
                'justify-center',
                activeTab === 'glossary'
                  ? 'bg-brand-primary text-text-inverse'
                  : 'text-text-muted hover:text-text-primary hover:bg-surface-hover',
              )}
            >
              <BookOpen className={iconTokens.size.sm} />
              Glossary
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
                placeholder={`Search ${activeTab}...`}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className={cn(
                  'w-full pl-9 pr-3 py-2 body-small',
                  'bg-surface-base border border-surface-border text-text-primary',
                  'placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-brand-primary',
                  radius.lg,
                )}
              />
            </div>
            <button
              type="button"
              onClick={() => setSimpleMode(!simpleMode)}
              className={cn(
                'px-3 py-2 text-xs font-medium transition-colors',
                radius.lg,
                simpleMode
                  ? 'bg-status-info text-text-inverse'
                  : 'bg-surface-base text-text-muted border border-surface-border hover:bg-surface-hover',
              )}
              title={simpleMode ? 'Showing simple explanations' : 'Showing technical explanations'}
            >
              {simpleMode ? 'Simple' : 'Technical'}
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
              onBack={() => setSelectedTest(null)}
              onCopy={copyCommand}
              copiedCommand={copiedCommand}
            />
          )}

          {/* Tutorial Detail View */}
          {selectedTutorial && activeTab === 'tutorials' && (
            <TutorialDetailView
              tutorial={selectedTutorial}
              onBack={() => setSelectedTutorial(null)}
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

// Tests Tab Component
function TestsTab({
  filteredTests,
  simpleMode,
  onSelectTest,
  getCategoryIcon,
}: {
  filteredTests: TestHelp[] | null;
  simpleMode: boolean;
  onSelectTest: (test: TestHelp) => void;
  getCategoryIcon: (id: string) => ReactElement;
}) {
  const categoryOrder = ['rfc2544', 'y1564', 'rfc2889', 'rfc6349', 'y1731', 'mef', 'tsn'];

  if (filteredTests) {
    return (
      <div className="space-y-2">
        <p className="text-xs text-[var(--color-text-muted)] mb-3">
          Found {filteredTests.length} test{filteredTests.length !== 1 ? 's' : ''}
        </p>
        {filteredTests.map((test) => (
          <TestCard
            key={test.id}
            test={test}
            simpleMode={simpleMode}
            onClick={() => onSelectTest(test)}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-[var(--color-text-muted)]">
        The Stem supports 27 tests across 7 standards. Click any test for detailed documentation.
      </p>
      {categoryOrder.map((catId) => {
        const category = categories[catId];
        if (!category) return null;
        const catTests = getTestsByCategory(catId);

        return (
          <CollapsibleSection
            key={catId}
            title={
              <div className="flex items-center gap-2">
                {getCategoryIcon(catId)}
                <span>{category.name}</span>
                <span className="text-xs text-[var(--color-text-muted)]">({catTests.length})</span>
              </div>
            }
            defaultOpen={catId === 'rfc2544'}
          >
            <div className="space-y-2">
              <p className="text-xs text-[var(--color-text-muted)] mb-2">{category.summary}</p>
              {catTests.map((test) => (
                <TestCard
                  key={test.id}
                  test={test}
                  simpleMode={simpleMode}
                  onClick={() => onSelectTest(test)}
                />
              ))}
            </div>
          </CollapsibleSection>
        );
      })}
    </div>
  );
}

// Test Card Component
function TestCard({
  test,
  simpleMode,
  onClick,
}: {
  test: TestHelp;
  simpleMode: boolean;
  onClick: () => void;
}): ReactElement {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'w-full text-left border transition-colors',
        spacing.pad.sm,
        radius.lg,
        'border-surface-border hover:border-brand-primary hover:bg-surface-hover',
      )}
    >
      <div className={cn(layout.flex.between, 'items-start')}>
        <div className="flex-1">
          <div className="font-medium body-small text-text-primary">{test.name}</div>
          <div className="caption mt-0.5">{test.standard}</div>
          <p className="caption text-text-secondary mt-1 line-clamp-2">
            {simpleMode ? test.laymanDesc.split('\n')[0] : test.summary}
          </p>
        </div>
        <ChevronRight className={cn(iconTokens.size.sm, 'text-text-muted flex-shrink-0 mt-1')} />
      </div>
    </button>
  );
}

// Test Detail View Component
function TestDetailView({
  test,
  simpleMode,
  onBack,
  onCopy,
  copiedCommand,
}: {
  test: TestHelp;
  simpleMode: boolean;
  onBack: () => void;
  onCopy: (cmd: string) => void;
  copiedCommand: string | null;
}) {
  return (
    <div className="section-gap">
      {/* Back Button */}
      <button
        type="button"
        onClick={onBack}
        className={cn(layout.inline.tight, 'body-small text-text-muted hover:text-text-primary')}
      >
        <ChevronRight className={cn(iconTokens.size.sm, 'rotate-180')} />
        Back to Tests
      </button>

      {/* Header */}
      <div>
        <h3 className="heading-3">{test.name}</h3>
        <p className="caption">{test.standard}</p>
      </div>

      {/* Description */}
      <div className={cn('bg-surface-base', radius.lg, spacing.pad.default)}>
        <h4 className={cn('section-title', spacing.margin.bottom.inline)}>
          {simpleMode ? 'What This Test Does' : 'Technical Description'}
        </h4>
        <p className="body-small text-text-primary whitespace-pre-line">
          {simpleMode ? test.laymanDesc : test.techDesc}
        </p>
      </div>

      {/* When to Use */}
      <div className={cn('bg-status-success/10', radius.lg, spacing.pad.default)}>
        <h4 className={cn('section-title text-status-success', spacing.margin.bottom.inline)}>
          When to Use This Test
        </h4>
        <p className="body-small text-text-primary whitespace-pre-line">{test.whenToUse}</p>
      </div>

      {/* When NOT to Use */}
      {test.whenNotToUse && (
        <div className={cn('bg-status-warning/10', radius.lg, spacing.pad.default)}>
          <h4 className={cn('section-title text-status-warning', spacing.margin.bottom.inline)}>
            When NOT to Use
          </h4>
          <p className="body-small text-text-primary whitespace-pre-line">{test.whenNotToUse}</p>
        </div>
      )}

      {/* Parameters */}
      {test.parameters.length > 0 && (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>Parameters</h4>
          <div className="stack-sm">
            {test.parameters.map((param) => (
              <div key={param.flag} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.inline.default}>
                  <code className={cn('code', 'bg-surface-raised px-1.5 py-0.5', radius.default)}>
                    {param.flag}
                  </code>
                  {param.required && <span className="caption text-status-warning">required</span>}
                </div>
                <p className="caption mt-1">
                  Type: {param.type} | Default: {param.defaultValue}
                </p>
                <p className="body-small text-text-secondary mt-1">
                  {simpleMode ? param.laymanDesc : param.techDesc}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Metrics */}
      {test.metrics.length > 0 && (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>Metrics</h4>
          <div className="stack-sm">
            {test.metrics.map((metric) => (
              <div key={metric.name} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.flex.between}>
                  <span className="font-medium body-small">{metric.name}</span>
                  <span className="caption">{metric.unit}</span>
                </div>
                <p className="caption text-status-success mt-1">Good: {metric.goodRange}</p>
                <p className="caption text-status-warning mt-0.5">Bad: {metric.badMeaning}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Examples */}
      {test.examples.length > 0 && (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>Examples</h4>
          <div className="stack-sm">
            {test.examples.map((example) => (
              <div
                key={example.command}
                className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}
              >
                <p className="caption mb-1">{example.desc}</p>
                <div className={cn(layout.flex.between, 'bg-surface-raised p-2', radius.default)}>
                  <code className="caption font-mono text-brand-primary">$ {example.command}</code>
                  <button
                    type="button"
                    onClick={() => onCopy(example.command)}
                    className={cn('p-1 hover:bg-surface-hover', radius.default)}
                    title="Copy command"
                  >
                    {copiedCommand === example.command ? (
                      <Check className={cn(iconTokens.size.xs, 'text-status-success')} />
                    ) : (
                      <Copy className={cn(iconTokens.size.xs, 'text-text-muted')} />
                    )}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tips */}
      {test.tips.length > 0 && (
        <div>
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>Tips</h4>
          <ul className="stack-xs">
            {test.tips.map((tip) => (
              <li
                key={tip}
                className={cn(layout.inline.default, 'body-small text-text-secondary items-start')}
              >
                <span className="text-brand-primary">-</span>
                {tip}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* See Also */}
      {test.seeAlso.length > 0 && (
        <div className="pt-4 border-t border-surface-border">
          <h4 className={cn('section-title', spacing.margin.bottom.inline)}>Related Tests</h4>
          <div className={layout.inline.wrap}>
            {test.seeAlso.map((related) => (
              <span
                key={related}
                className={cn('caption bg-surface-base px-2 py-1', radius.default)}
              >
                {related}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Tutorials Tab Component
function TutorialsTab({
  searchQuery,
  onSelectTutorial,
}: {
  searchQuery: string;
  onSelectTutorial: (tutorial: Tutorial) => void;
}) {
  const tutorialList = Object.values(tutorials);
  const filtered = searchQuery
    ? tutorialList.filter(
        (t) =>
          t.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
          t.description.toLowerCase().includes(searchQuery.toLowerCase()),
      )
    : tutorialList;

  return (
    <div className="section-gap">
      <p className="body-small text-text-muted">
        Step-by-step guides to help you get started with network testing.
      </p>
      <div className="stack-sm">
        {filtered.map((tutorial) => (
          <button
            type="button"
            key={tutorial.id}
            onClick={() => onSelectTutorial(tutorial)}
            className={cn(
              'w-full text-left border transition-colors',
              spacing.pad.default,
              radius.lg,
              'border-surface-border hover:border-brand-primary hover:bg-surface-hover',
            )}
          >
            <div className={cn(layout.flex.between, 'items-start')}>
              <div className="flex-1">
                <div className="font-medium body-small text-text-primary">{tutorial.title}</div>
                <div className={cn(layout.inline.default, 'mt-1')}>
                  <span className="caption">{tutorial.duration}</span>
                  <span className={cn('caption px-1.5 py-0.5 bg-surface-base', radius.default)}>
                    {tutorial.level}
                  </span>
                </div>
                <p className="caption text-text-secondary mt-2">{tutorial.description}</p>
              </div>
              <ChevronRight className={cn(iconTokens.size.sm, 'text-text-muted flex-shrink-0')} />
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

// Tutorial Detail View Component
function TutorialDetailView({
  tutorial,
  onBack,
  onCopy,
  copiedCommand,
}: {
  tutorial: Tutorial;
  onBack: () => void;
  onCopy: (cmd: string) => void;
  copiedCommand: string | null;
}): ReactElement {
  return (
    <div className="section-gap">
      {/* Back Button */}
      <button
        type="button"
        onClick={onBack}
        className={cn(layout.inline.tight, 'body-small text-text-muted hover:text-text-primary')}
      >
        <ChevronRight className={cn(iconTokens.size.sm, 'rotate-180')} />
        Back to Tutorials
      </button>

      {/* Header */}
      <div>
        <h3 className="heading-3">{tutorial.title}</h3>
        <div className={cn(layout.inline.default, 'mt-1')}>
          <span className="caption">{tutorial.duration}</span>
          <span className={cn('caption px-1.5 py-0.5 bg-surface-base', radius.default)}>
            {tutorial.level}
          </span>
        </div>
        <p className="body-small text-text-secondary mt-2">{tutorial.description}</p>
      </div>

      {/* Steps */}
      <div className="stack-lg">
        {tutorial.steps.map((step, stepIndex) => (
          <div key={step.title} className={cn('bg-surface-base', radius.lg, spacing.pad.default)}>
            <div className={cn(layout.inline.default, spacing.margin.bottom.inline)}>
              <span
                className={cn(
                  'w-6 h-6 text-xs font-medium',
                  radius.full,
                  'bg-brand-primary text-text-inverse',
                  layout.flex.center,
                )}
              >
                {stepIndex + 1}
              </span>
              <h4 className="font-medium body-small text-text-primary">{step.title}</h4>
            </div>
            <p
              className={cn(
                'body-small text-text-secondary whitespace-pre-line',
                spacing.margin.bottom.heading,
              )}
            >
              {step.content}
            </p>
            {step.command && (
              <div
                className={cn(layout.flex.between, 'bg-surface-raised p-2 mb-2', radius.default)}
              >
                <code className="caption font-mono text-brand-primary">$ {step.command}</code>
                <button
                  type="button"
                  onClick={() => onCopy(step.command as string)}
                  className={cn('p-1 hover:bg-surface-hover', radius.default)}
                  title="Copy command"
                >
                  {copiedCommand === step.command ? (
                    <Check className={cn(iconTokens.size.xs, 'text-status-success')} />
                  ) : (
                    <Copy className={cn(iconTokens.size.xs, 'text-text-muted')} />
                  )}
                </button>
              </div>
            )}
            {step.expected && (
              <div className={cn('caption bg-surface-raised p-2', radius.default)}>
                Expected: {step.expected}
              </div>
            )}
            {step.tip && (
              <div
                className={cn(
                  'mt-2 caption text-status-info bg-status-info/10 p-2',
                  radius.default,
                )}
              >
                Tip: {step.tip}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// Glossary Tab Component
function GlossaryTab({
  searchQuery,
  filteredGlossary,
  simpleMode,
}: {
  searchQuery: string;
  filteredGlossary: GlossaryEntry[] | null;
  simpleMode: boolean;
}) {
  const glossaryEntries = filteredGlossary || Object.values(glossary);

  // Group by category
  const byCategory = glossaryEntries.reduce(
    (acc, entry) => {
      if (!acc[entry.category]) {
        acc[entry.category] = [];
      }
      acc[entry.category].push(entry);
      return acc;
    },
    {} as Record<string, GlossaryEntry[]>,
  );

  const categoryNames = Object.keys(byCategory).sort();

  return (
    <div className="section-gap">
      <p className="body-small text-text-muted">
        Network testing terminology explained {simpleMode ? 'simply' : 'technically'}.
      </p>
      {searchQuery && (
        <p className="caption">
          Found {glossaryEntries.length} term{glossaryEntries.length !== 1 ? 's' : ''}
        </p>
      )}
      {categoryNames.map((category) => (
        <CollapsibleSection
          key={category}
          title={<span>{category}</span>}
          defaultOpen={searchQuery.length > 0}
        >
          <div className="stack-sm">
            {byCategory[category].map((entry) => (
              <div key={entry.term} className={cn('bg-surface-base', radius.lg, spacing.pad.sm)}>
                <div className={layout.inline.default}>
                  <span className="font-medium body-small text-text-primary">{entry.term}</span>
                  <span className="caption">{entry.fullName}</span>
                </div>
                <p className="body-small text-text-secondary mt-1">
                  {simpleMode ? entry.laymanDef : entry.techDef}
                </p>
                {entry.related.length > 0 && (
                  <div className={cn(layout.inline.tight, 'mt-2')}>
                    <span className="caption">Related:</span>
                    {entry.related.map((r) => (
                      <span key={r} className="caption text-brand-primary">
                        {r}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </CollapsibleSection>
      ))}
    </div>
  );
}

export default HelpDrawer;
