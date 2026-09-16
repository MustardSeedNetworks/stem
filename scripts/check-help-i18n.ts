#!/usr/bin/env node
/** Reject help locale entries with no production consumer. */
import { parse } from '../ui/node_modules/@babel/parser/lib/index.js';
import * as t from '../ui/node_modules/@babel/types/lib/index.js';
import { readFileSync, readdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

export function leafKeys(value: unknown, prefix = ''): string[] {
  if (typeof value === 'string') return [prefix];
  if (!value || typeof value !== 'object') return [];
  return Object.entries(value).flatMap(([key, child]) =>
    leafKeys(child, prefix ? `${prefix}.${key}` : key),
  );
}

const visit = (node: t.Node, inspect: (node: t.Node) => void): void => {
  inspect(node);
  const fields = node as unknown as Record<string, unknown>;
  for (const key of t.VISITOR_KEYS[node.type] ?? []) {
    const value = fields[key];
    for (const child of Array.isArray(value) ? value : [value]) {
      if (child && typeof child === 'object' && 'type' in child) visit(child as t.Node, inspect);
    }
  }
};

export function consumedHelpKeys(
  source: string,
  domains: Record<string, string[]> = {},
): Set<string> {
  const ast = parse(source, { sourceType: 'module', plugins: ['typescript', 'jsx'] });
  const keys = new Set<string>();
  const translators = new Set<string>();
  const helpTranslators = new Set<string>();
  visit(ast, (node) => {
    if (
      !t.isVariableDeclarator(node) ||
      !t.isObjectPattern(node.id) ||
      !t.isCallExpression(node.init) ||
      !t.isIdentifier(node.init.callee, { name: 'useTranslation' })
    )
      return;
    const ns = node.init.arguments[0];
    const first = t.isArrayExpression(ns) ? ns.elements[0] : ns;
    for (const property of node.id.properties) {
      if (
        t.isObjectProperty(property) &&
        t.isIdentifier(property.key, { name: 't' }) &&
        t.isIdentifier(property.value)
      ) {
        translators.add(property.value.name);
        if (t.isStringLiteral(first, { value: 'help' })) helpTranslators.add(property.value.name);
      }
    }
  });
  visit(ast, (node) => {
    if (!t.isIdentifier(node) || !t.isTSTypeAnnotation(node.typeAnnotation)) return;
    const type = node.typeAnnotation.typeAnnotation;
    if (!t.isTSTypeReference(type) || !t.isIdentifier(type.typeName, { name: 'TFunction' })) return;
    translators.add(node.name);
    const namespace = type.typeArguments?.params[0];
    if (t.isTSLiteralType(namespace) && t.isStringLiteral(namespace.literal, { value: 'help' }))
      helpTranslators.add(node.name);
  });
  visit(ast, (node) => {
    if (t.isCallExpression(node) && t.isIdentifier(node.callee)) {
      const arg = node.arguments[0];
      if (t.isStringLiteral(arg)) {
        if (translators.has(node.callee.name) && arg.value.startsWith('help:'))
          keys.add(arg.value.slice(5));
        else if (helpTranslators.has(node.callee.name)) keys.add(arg.value);
      }
    }
    if (t.isCallExpression(node) && t.isIdentifier(node.callee)) {
      const arg = node.arguments[0];
      if (
        helpTranslators.has(node.callee.name) &&
        t.isTemplateLiteral(arg) &&
        arg.expressions.length === 1
      ) {
        const prefix = arg.quasis[0]?.value.cooked ?? '';
        const suffix = arg.quasis[1]?.value.cooked ?? '';
        for (const id of domains[prefix] ?? []) keys.add(`${prefix}${id}${suffix}`);
      }
    }
  });
  return keys;
}

export function checkHelp(root: string): string[] {
  const used = new Set<string>();
  const domains: Record<string, string[]> = {};
  const corpus = join(root, 'ui/src/data/help/tests');
  if (existsSync(corpus)) {
    domains['tests.'] = [];
    for (const file of readdirSync(corpus).filter(
      (name) => name.endsWith('.ts') && !name.includes('.test.'),
    )) {
      visit(
        parse(readFileSync(join(corpus, file), 'utf8'), {
          sourceType: 'module',
          plugins: ['typescript'],
        }),
        (node) => {
          if (
            t.isObjectProperty(node) &&
            t.isIdentifier(node.key, { name: 'id' }) &&
            t.isStringLiteral(node.value)
          ) {
            domains['tests.'].push(node.value.value);
          }
        },
      );
    }
  }
  const glossary = join(root, 'ui/src/data/help/glossary.ts');
  if (existsSync(glossary)) {
    domains['glossary.entries.'] = [];
    visit(
      parse(readFileSync(glossary, 'utf8'), { sourceType: 'module', plugins: ['typescript'] }),
      (node) => {
        if (
          !t.isVariableDeclarator(node) ||
          !t.isIdentifier(node.id, { name: 'glossaryRelated' }) ||
          !t.isObjectExpression(node.init)
        )
          return;
        for (const property of node.init.properties) {
          if (t.isObjectProperty(property) && t.isIdentifier(property.key))
            domains['glossary.entries.'].push(property.key.name);
        }
      },
    );
  }
  const drawer = join(root, 'ui/src/components/HelpDrawer.tsx');
  if (existsSync(drawer)) {
    domains['tabs.'] = [];
    visit(
      parse(readFileSync(drawer, 'utf8'), { sourceType: 'module', plugins: ['typescript', 'jsx'] }),
      (node) => {
        if (
          !t.isTSTypeAliasDeclaration(node) ||
          node.id.name !== 'HelpTab' ||
          !t.isTSUnionType(node.typeAnnotation)
        )
          return;
        for (const type of node.typeAnnotation.types) {
          if (t.isTSLiteralType(type) && t.isStringLiteral(type.literal))
            domains['tabs.'].push(type.literal.value);
        }
      },
    );
  }
  const scan = (dir: string): void => {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const file = join(dir, entry.name);
      if (entry.isDirectory()) scan(file);
      else if (
        /\.tsx?$/.test(file) &&
        !/\.(test|stories)\.tsx?$/.test(file) &&
        !file.includes('/test/')
      ) {
        for (const key of consumedHelpKeys(readFileSync(file, 'utf8'), domains)) used.add(key);
      }
    }
  };
  scan(join(root, 'ui/src'));
  const locale: unknown = JSON.parse(readFileSync(join(root, 'ui/locales/en/help.json'), 'utf8'));
  return leafKeys(locale)
    .filter((key) => !used.has(key))
    .sort();
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const root = process.argv[2] ?? dirname(dirname(fileURLToPath(import.meta.url)));
  const orphaned = checkHelp(root);
  if (orphaned.length) {
    process.stderr.write(`Help keys without production consumers:\n${orphaned.join('\n')}\n`);
    process.exitCode = 1;
  } else process.stdout.write('Help locale: every key has a production consumer.\n');
}
