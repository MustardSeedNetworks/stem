import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { test } from 'node:test';
import { checkHelp, consumedHelpKeys } from './check-help-i18n.ts';

test('an unused help key fails even when mentioned in a comment or test', () => {
  const root = mkdtempSync(join(tmpdir(), 'stem-help-i18n-'));
  try {
    mkdirSync(join(root, 'ui/src'), { recursive: true });
    mkdirSync(join(root, 'ui/locales/en'), { recursive: true });
    writeFileSync(
      join(root, 'ui/locales/en/help.json'),
      JSON.stringify({ modal: { title: 'Help', orphan: 'Unused' } }),
    );
    writeFileSync(
      join(root, 'ui/src/Help.tsx'),
      "const {t}=useTranslation('help'); t('modal.title'); // t('modal.orphan')\n",
    );
    writeFileSync(join(root, 'ui/src/Help.test.tsx'), "t('help:modal.orphan');");
    assert.deepEqual(checkHelp(root), ['modal.orphan']);
    writeFileSync(
      join(root, 'ui/src/Help.tsx'),
      "const {t}=useTranslation('help'); t('modal.title'); t('modal.orphan');\n",
    );
    assert.deepEqual(checkHelp(root), []);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test('expands dynamic keys only across production registry IDs', () => {
  const source = "const {t}=useTranslation('help'); t(`tests.${test.id}.whenToUse`);";
  assert.deepEqual(
    [...consumedHelpKeys(source, { 'tests.': ['throughput'] })],
    ['tests.throughput.whenToUse'],
  );
});

test('commented registry IDs cannot hide orphaned dynamic translations', () => {
  const root = mkdtempSync(join(tmpdir(), 'stem-help-registry-'));
  try {
    mkdirSync(join(root, 'ui/src/data/help/tests'), { recursive: true });
    mkdirSync(join(root, 'ui/locales/en'), { recursive: true });
    writeFileSync(
      join(root, 'ui/locales/en/help.json'),
      JSON.stringify({
        tests: { throughput: { whenToUse: 'Use this' }, orphan: { whenToUse: 'Unused' } },
      }),
    );
    writeFileSync(
      join(root, 'ui/src/Help.tsx'),
      "const {t}=useTranslation('help'); t(`tests.${test.id}.whenToUse`);",
    );
    writeFileSync(
      join(root, 'ui/src/data/help/tests/tests.ts'),
      "export const tests = { throughput: { id: 'throughput' } }; /* id: 'orphan' */",
    );
    assert.deepEqual(checkHelp(root), ['tests.orphan.whenToUse']);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test('non-translation calls do not count as help consumers', () => {
  assert.deepEqual(
    [
      ...consumedHelpKeys(
        "String('help:orphan'); const {t}=useTranslation('settings'); t('help:real');",
      ),
    ],
    ['real'],
  );
});

test('typed help translators expand production glossary IDs', () => {
  assert.deepEqual(
    [
      ...consumedHelpKeys(
        "function glossary(t: TFunction<'help'>) { return t(`glossary.entries.${id}.term`); }",
        { 'glossary.entries.': ['throughput'] },
      ),
    ],
    ['glossary.entries.throughput.term'],
  );
});
