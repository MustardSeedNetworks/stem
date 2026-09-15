/**
 * Which selected test ids belong to which config form.
 *
 * Each ConfigForm used to carry its own prefix test inline, so a page had no
 * way to ask "is anything of mine selected?" without repeating the form's
 * private rule. The pages need that answer to decide between the form and an
 * empty state (#1257), so the rule lives here once and the forms read it too.
 */

export type TestGroup =
  | 'rfc2544'
  | 'y1564'
  | 'trafficgen'
  | 'y1731'
  | 'rfc2889'
  | 'rfc6349'
  | 'tsn';

const belongsTo: Record<TestGroup, (id: string) => boolean> = {
  rfc2544: (id) => id.startsWith('rfc2544'),
  y1564: (id) => id.startsWith('y1564') || id.startsWith('mef'),
  trafficgen: (id) => id.startsWith('trafficgen_') || id === 'custom_stream',
  y1731: (id) => id.startsWith('y1731'),
  rfc2889: (id) => id.startsWith('rfc2889'),
  rfc6349: (id) => id.startsWith('rfc6349'),
  tsn: (id) => id.startsWith('tsn_'),
};

export function hasGroupTests(group: TestGroup, selectedTests: readonly string[]): boolean {
  return selectedTests.some(belongsTo[group]);
}

export function hasAnyGroupTests(
  groups: readonly TestGroup[],
  selectedTests: readonly string[],
): boolean {
  return groups.some((group) => hasGroupTests(group, selectedTests));
}
