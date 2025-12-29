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
      <button
        type="button"
        className="fixed inset-0 bg-black/50 z-40 cursor-default"
        onClick={onClose}
        aria-label="Close help drawer"
      />

      {/* Drawer */}
      <div className="fixed right-0 top-0 h-full w-[480px] max-w-full bg-[var(--color-surface-raised)] border-l border-[var(--color-surface-border)] z-50 flex flex-col">
        {/* Header */}
        <div className="sticky top-0 bg-[var(--color-surface-raised)] border-b border-[var(--color-surface-border)] px-4 py-3">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <BookOpen className="w-5 h-5 text-[var(--color-stem-green)]" />
              <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">
                Help & Documentation
              </h2>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="p-2 hover:bg-[var(--color-surface-hover)] rounded-lg transition-colors"
            >
              <X className="w-5 h-5 text-[var(--color-text-muted)]" />
            </button>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 bg-[var(--color-surface-base)] rounded-lg p-1">
            <button
              type="button"
              onClick={() => setActiveTab('tests')}
              className={`flex-1 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'tests'
                  ? 'bg-[var(--color-stem-green)] text-white'
                  : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              <Book className="w-4 h-4 inline mr-1" />
              Tests
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('tutorials')}
              className={`flex-1 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'tutorials'
                  ? 'bg-[var(--color-stem-green)] text-white'
                  : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              <GraduationCap className="w-4 h-4 inline mr-1" />
              Tutorials
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('glossary')}
              className={`flex-1 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === 'glossary'
                  ? 'bg-[var(--color-stem-green)] text-white'
                  : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              <BookOpen className="w-4 h-4 inline mr-1" />
              Glossary
            </button>
          </div>

          {/* Search and Mode Toggle */}
          <div className="flex gap-2 mt-3">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder={`Search ${activeTab}...`}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-3 py-2 bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] rounded-lg text-sm"
              />
            </div>
            <button
              type="button"
              onClick={() => setSimpleMode(!simpleMode)}
              className={`px-3 py-2 rounded-lg text-xs font-medium transition-colors ${
                simpleMode
                  ? 'bg-[var(--color-status-info)] text-white'
                  : 'bg-[var(--color-surface-base)] text-[var(--color-text-muted)] border border-[var(--color-surface-border)]'
              }`}
              title={simpleMode ? 'Showing simple explanations' : 'Showing technical explanations'}
            >
              {simpleMode ? 'Simple' : 'Technical'}
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
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
      className="w-full text-left p-3 rounded-lg border border-[var(--color-surface-border)] hover:border-[var(--color-stem-green)] hover:bg-[var(--color-surface-hover)] transition-colors"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="font-medium text-sm text-[var(--color-text-primary)]">{test.name}</div>
          <div className="text-xs text-[var(--color-text-muted)] mt-0.5">{test.standard}</div>
          <p className="text-xs text-[var(--color-text-secondary)] mt-1 line-clamp-2">
            {simpleMode ? test.laymanDesc.split('\n')[0] : test.summary}
          </p>
        </div>
        <ChevronRight className="w-4 h-4 text-[var(--color-text-muted)] flex-shrink-0 mt-1" />
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
    <div className="space-y-4">
      {/* Back Button */}
      <button
        type="button"
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
      >
        <ChevronRight className="w-4 h-4 rotate-180" />
        Back to Tests
      </button>

      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">{test.name}</h3>
        <p className="text-xs text-[var(--color-text-muted)]">{test.standard}</p>
      </div>

      {/* Description */}
      <div className="bg-[var(--color-surface-base)] rounded-lg p-4">
        <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
          {simpleMode ? 'What This Test Does' : 'Technical Description'}
        </h4>
        <p className="text-sm text-[var(--color-text-primary)] whitespace-pre-line">
          {simpleMode ? test.laymanDesc : test.techDesc}
        </p>
      </div>

      {/* When to Use */}
      <div className="bg-[var(--color-status-success)]/10 rounded-lg p-4">
        <h4 className="text-xs font-medium text-[var(--color-status-success)] uppercase mb-2">
          When to Use This Test
        </h4>
        <p className="text-sm text-[var(--color-text-primary)] whitespace-pre-line">
          {test.whenToUse}
        </p>
      </div>

      {/* When NOT to Use */}
      {test.whenNotToUse && (
        <div className="bg-[var(--color-status-warning)]/10 rounded-lg p-4">
          <h4 className="text-xs font-medium text-[var(--color-status-warning)] uppercase mb-2">
            When NOT to Use
          </h4>
          <p className="text-sm text-[var(--color-text-primary)] whitespace-pre-line">
            {test.whenNotToUse}
          </p>
        </div>
      )}

      {/* Parameters */}
      {test.parameters.length > 0 && (
        <div>
          <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
            Parameters
          </h4>
          <div className="space-y-2">
            {test.parameters.map((param) => (
              <div key={param.flag} className="bg-[var(--color-surface-base)] rounded-lg p-3">
                <div className="flex items-center gap-2">
                  <code className="text-xs font-mono bg-[var(--color-surface-raised)] px-1.5 py-0.5 rounded">
                    {param.flag}
                  </code>
                  {param.required && (
                    <span className="text-xs text-[var(--color-status-warning)]">required</span>
                  )}
                </div>
                <p className="text-xs text-[var(--color-text-muted)] mt-1">
                  Type: {param.type} | Default: {param.defaultValue}
                </p>
                <p className="text-sm text-[var(--color-text-secondary)] mt-1">
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
          <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
            Metrics
          </h4>
          <div className="space-y-2">
            {test.metrics.map((metric) => (
              <div key={metric.name} className="bg-[var(--color-surface-base)] rounded-lg p-3">
                <div className="flex items-center justify-between">
                  <span className="font-medium text-sm">{metric.name}</span>
                  <span className="text-xs text-[var(--color-text-muted)]">{metric.unit}</span>
                </div>
                <p className="text-xs text-[var(--color-status-success)] mt-1">
                  Good: {metric.goodRange}
                </p>
                <p className="text-xs text-[var(--color-status-warning)] mt-0.5">
                  Bad: {metric.badMeaning}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Examples */}
      {test.examples.length > 0 && (
        <div>
          <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
            Examples
          </h4>
          <div className="space-y-2">
            {test.examples.map((example) => (
              <div key={example.command} className="bg-[var(--color-surface-base)] rounded-lg p-3">
                <p className="text-xs text-[var(--color-text-muted)] mb-1">{example.desc}</p>
                <div className="flex items-center justify-between bg-[var(--color-surface-raised)] rounded p-2">
                  <code className="text-xs font-mono text-[var(--color-stem-green)]">
                    $ {example.command}
                  </code>
                  <button
                    type="button"
                    onClick={() => onCopy(example.command)}
                    className="p-1 hover:bg-[var(--color-surface-hover)] rounded"
                    title="Copy command"
                  >
                    {copiedCommand === example.command ? (
                      <Check className="w-3 h-3 text-[var(--color-status-success)]" />
                    ) : (
                      <Copy className="w-3 h-3 text-[var(--color-text-muted)]" />
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
          <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
            Tips
          </h4>
          <ul className="space-y-1">
            {test.tips.map((tip) => (
              <li
                key={tip}
                className="text-sm text-[var(--color-text-secondary)] flex items-start gap-2"
              >
                <span className="text-[var(--color-stem-green)]">-</span>
                {tip}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* See Also */}
      {test.seeAlso.length > 0 && (
        <div className="pt-4 border-t border-[var(--color-surface-border)]">
          <h4 className="text-xs font-medium text-[var(--color-text-muted)] uppercase mb-2">
            Related Tests
          </h4>
          <div className="flex flex-wrap gap-2">
            {test.seeAlso.map((related) => (
              <span
                key={related}
                className="text-xs bg-[var(--color-surface-base)] px-2 py-1 rounded"
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
    <div className="space-y-4">
      <p className="text-sm text-[var(--color-text-muted)]">
        Step-by-step guides to help you get started with network testing.
      </p>
      <div className="space-y-2">
        {filtered.map((tutorial) => (
          <button
            type="button"
            key={tutorial.id}
            onClick={() => onSelectTutorial(tutorial)}
            className="w-full text-left p-4 rounded-lg border border-[var(--color-surface-border)] hover:border-[var(--color-stem-green)] hover:bg-[var(--color-surface-hover)] transition-colors"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="font-medium text-sm text-[var(--color-text-primary)]">
                  {tutorial.title}
                </div>
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-xs text-[var(--color-text-muted)]">
                    {tutorial.duration}
                  </span>
                  <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-surface-base)]">
                    {tutorial.level}
                  </span>
                </div>
                <p className="text-xs text-[var(--color-text-secondary)] mt-2">
                  {tutorial.description}
                </p>
              </div>
              <ChevronRight className="w-4 h-4 text-[var(--color-text-muted)] flex-shrink-0" />
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
    <div className="space-y-4">
      {/* Back Button */}
      <button
        type="button"
        onClick={onBack}
        className="flex items-center gap-1 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
      >
        <ChevronRight className="w-4 h-4 rotate-180" />
        Back to Tutorials
      </button>

      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">{tutorial.title}</h3>
        <div className="flex items-center gap-2 mt-1">
          <span className="text-xs text-[var(--color-text-muted)]">{tutorial.duration}</span>
          <span className="text-xs px-1.5 py-0.5 rounded bg-[var(--color-surface-base)]">
            {tutorial.level}
          </span>
        </div>
        <p className="text-sm text-[var(--color-text-secondary)] mt-2">{tutorial.description}</p>
      </div>

      {/* Steps */}
      <div className="space-y-4">
        {tutorial.steps.map((step, stepIndex) => (
          <div key={step.title} className="bg-[var(--color-surface-base)] rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <span className="w-6 h-6 rounded-full bg-[var(--color-stem-green)] text-white text-xs flex items-center justify-center font-medium">
                {stepIndex + 1}
              </span>
              <h4 className="font-medium text-sm text-[var(--color-text-primary)]">{step.title}</h4>
            </div>
            <p className="text-sm text-[var(--color-text-secondary)] whitespace-pre-line mb-3">
              {step.content}
            </p>
            {step.command && (
              <div className="flex items-center justify-between bg-[var(--color-surface-raised)] rounded p-2 mb-2">
                <code className="text-xs font-mono text-[var(--color-stem-green)]">
                  $ {step.command}
                </code>
                <button
                  type="button"
                  onClick={() => onCopy(step.command as string)}
                  className="p-1 hover:bg-[var(--color-surface-hover)] rounded"
                  title="Copy command"
                >
                  {copiedCommand === step.command ? (
                    <Check className="w-3 h-3 text-[var(--color-status-success)]" />
                  ) : (
                    <Copy className="w-3 h-3 text-[var(--color-text-muted)]" />
                  )}
                </button>
              </div>
            )}
            {step.expected && (
              <div className="text-xs text-[var(--color-text-muted)] bg-[var(--color-surface-raised)] rounded p-2">
                Expected: {step.expected}
              </div>
            )}
            {step.tip && (
              <div className="mt-2 text-xs text-[var(--color-status-info)] bg-[var(--color-status-info)]/10 rounded p-2">
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
    <div className="space-y-4">
      <p className="text-sm text-[var(--color-text-muted)]">
        Network testing terminology explained {simpleMode ? 'simply' : 'technically'}.
      </p>
      {searchQuery && (
        <p className="text-xs text-[var(--color-text-muted)]">
          Found {glossaryEntries.length} term{glossaryEntries.length !== 1 ? 's' : ''}
        </p>
      )}
      {categoryNames.map((category) => (
        <CollapsibleSection
          key={category}
          title={<span>{category}</span>}
          defaultOpen={searchQuery.length > 0}
        >
          <div className="space-y-3">
            {byCategory[category].map((entry) => (
              <div key={entry.term} className="bg-[var(--color-surface-base)] rounded-lg p-3">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm text-[var(--color-text-primary)]">
                    {entry.term}
                  </span>
                  <span className="text-xs text-[var(--color-text-muted)]">{entry.fullName}</span>
                </div>
                <p className="text-sm text-[var(--color-text-secondary)] mt-1">
                  {simpleMode ? entry.laymanDef : entry.techDef}
                </p>
                {entry.related.length > 0 && (
                  <div className="flex items-center gap-1 mt-2">
                    <span className="text-xs text-[var(--color-text-muted)]">Related:</span>
                    {entry.related.map((r) => (
                      <span key={r} className="text-xs text-[var(--color-stem-green)]">
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
