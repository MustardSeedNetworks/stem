# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.20.0](https://github.com/krisarmstrong/stem/compare/v0.19.0...v0.20.0) (2026-05-27)


### Features

* **api:** port seed's strict JSON decode helpers ([#320](https://github.com/krisarmstrong/stem/issues/320)) ([#322](https://github.com/krisarmstrong/stem/issues/322)) ([9499f86](https://github.com/krisarmstrong/stem/commit/9499f86612c168fad10c4110c581f4bd28415abf))
* **forms:** adopt react-hook-form + valibot resolver, Y.1564 pilot ([#325](https://github.com/krisarmstrong/stem/issues/325)) ([#328](https://github.com/krisarmstrong/stem/issues/328)) ([bf7937a](https://github.com/krisarmstrong/stem/commit/bf7937aa0dfa062469e7ace706c3dbc2254c944c))
* **forms:** adopt react-hook-form for MFA setup + disable ([#325](https://github.com/krisarmstrong/stem/issues/325)) ([#331](https://github.com/krisarmstrong/stem/issues/331)) ([92cffef](https://github.com/krisarmstrong/stem/commit/92cffef7240aad661ece2130fe8006f9670cab40))
* **forms:** finish react-hook-form rollout — App.tsx auth + Recovery + Setup ([#332](https://github.com/krisarmstrong/stem/issues/332)) ([#333](https://github.com/krisarmstrong/stem/issues/333)) ([2afb218](https://github.com/krisarmstrong/stem/commit/2afb218a9a40e030cf89dd3c3b2b8f40d84ff1f0))
* **forms:** sweep 6 remaining ConfigForms onto react-hook-form ([#325](https://github.com/krisarmstrong/stem/issues/325)) ([#330](https://github.com/krisarmstrong/stem/issues/330)) ([224b39d](https://github.com/krisarmstrong/stem/commit/224b39d5d5e4d7f873bdce2162db0502ee16e2e7))
* **i18n:** add per-repo dynamic-prefixes allowlist for check-keys.py ([#335](https://github.com/krisarmstrong/stem/issues/335)) ([cd6a793](https://github.com/krisarmstrong/stem/commit/cd6a793e5faf13b5d4efc93d8c2b94de8564c9f5))
* **i18n:** add useLocale hook + migrate HelpDrawer plurals ([#324](https://github.com/krisarmstrong/stem/issues/324)) ([771a07c](https://github.com/krisarmstrong/stem/commit/771a07c091b79223bef45d00b0987ec4ff8c394b))
* **i18n:** extract ModuleCard a11y strings to locale keys ([#319](https://github.com/krisarmstrong/stem/issues/319)) ([56eacd5](https://github.com/krisarmstrong/stem/commit/56eacd59917218651c9d179067a7b2709ea2d3ee))
* **i18n:** port check-keys.py + add phase 6 unit tests ([#327](https://github.com/krisarmstrong/stem/issues/327)) ([85a3382](https://github.com/krisarmstrong/stem/commit/85a3382a3597b893223fe9e0461a740919046fe8))


### Bug Fixes

* **auth:** remove t.Parallel from HIBP tests that mutate shared endpoint ([#315](https://github.com/krisarmstrong/stem/issues/315)) ([08ef121](https://github.com/krisarmstrong/stem/commit/08ef121f47546d52e846afb741e51f58ab32dee9))
* **i18n-es:** normalize accents and add missing diacritics (107 fixes) ([#321](https://github.com/krisarmstrong/stem/issues/321)) ([c4e5870](https://github.com/krisarmstrong/stem/commit/c4e58708a596b760447656261849a23e7b494abd))
* **i18n:** resolve 48 t() calls referencing missing EN locale keys ([#329](https://github.com/krisarmstrong/stem/issues/329)) ([b0232aa](https://github.com/krisarmstrong/stem/commit/b0232aaeee5362edd79c0dd3c42165a366b186bf))
* **i18n:** update document.lang on locale change for a11y ([#316](https://github.com/krisarmstrong/stem/issues/316)) ([963bf12](https://github.com/krisarmstrong/stem/commit/963bf128daa57e09a00d2da4414604922a5d5ce2))

## [0.19.0](https://github.com/krisarmstrong/stem/compare/v0.18.1...v0.19.0) (2026-05-26)


### Features

* **i18n:** add errors.license.* keys for tier-gating UI ([#313](https://github.com/krisarmstrong/stem/issues/313)) ([0e406de](https://github.com/krisarmstrong/stem/commit/0e406de96680b32f3dcdf54a794d9331ec8b47cd))

## [0.18.1](https://github.com/krisarmstrong/stem/compare/v0.18.0...v0.18.1) (2026-05-26)


### Bug Fixes

* **license:** add RWMutex to Manager for safe concurrent access ([#312](https://github.com/krisarmstrong/stem/issues/312)) ([cf8afe3](https://github.com/krisarmstrong/stem/commit/cf8afe3c1559833a57bcba858431517b824e560a))
* **scripts:** clean up all shellcheck warnings + pin severity=warning ([#307](https://github.com/krisarmstrong/stem/issues/307)) ([43c2d62](https://github.com/krisarmstrong/stem/commit/43c2d62f64ffaad5636343adfa51568f53105a8b))

## [0.18.0](https://github.com/krisarmstrong/stem/compare/v0.17.2...v0.18.0) (2026-05-25)


### Features

* **api:** add go-playground/validator + tags on hot DTOs ([#294](https://github.com/krisarmstrong/stem/issues/294)) ([95236b5](https://github.com/krisarmstrong/stem/commit/95236b564fb25312a433fef35b3f6a453d469e33))
* **api:** port invopop/jsonschema generator from NIAC ([#297](https://github.com/krisarmstrong/stem/issues/297)) ([d489eef](https://github.com/krisarmstrong/stem/commit/d489eef332981b2b18ecdf2b30afd50385ebb4d4)), closes [#269](https://github.com/krisarmstrong/stem/issues/269)
* **ui:** harden RoleContext with valibot schemas ([#295](https://github.com/krisarmstrong/stem/issues/295)) ([ef93f25](https://github.com/krisarmstrong/stem/commit/ef93f251f5a0dbd3dc0028f372b33412f5ef956a)), closes [#272](https://github.com/krisarmstrong/stem/issues/272)


### Bug Fixes

* **ci:** inject UIBuildHash ldflag (Universal Build Contract) ([#282](https://github.com/krisarmstrong/stem/issues/282)) ([ca443be](https://github.com/krisarmstrong/stem/commit/ca443be20f491f7dcb36785e4eb485248006c2f6))
* **ci:** trigger CodeQL on PR + push + weekly schedule ([#293](https://github.com/krisarmstrong/stem/issues/293)) ([736fed4](https://github.com/krisarmstrong/stem/commit/736fed43badd1852109845ee1011f4b452f8f539))
* **ci:** verify UIBuildHash embedded in built binary ([#286](https://github.com/krisarmstrong/stem/issues/286)) ([b35a6f4](https://github.com/krisarmstrong/stem/commit/b35a6f4d1383f31a41fe8510a8e64396a9cf310a))
* **docs:** correct PR template 'cd web' -&gt; 'cd ui' ([#283](https://github.com/krisarmstrong/stem/issues/283)) ([79c2782](https://github.com/krisarmstrong/stem/commit/79c27821344623197fe24006c34b4736f0f05379))
* **scripts:** deploy-validate default scheme=https + port=8444 ([#292](https://github.com/krisarmstrong/stem/issues/292)) ([acdf2a6](https://github.com/krisarmstrong/stem/commit/acdf2a684758a25810a972e11582a55f388fd797))
* **ui:** enable erasableSyntaxOnly + refactor ApiError TS-only syntax ([#290](https://github.com/krisarmstrong/stem/issues/290)) ([1cc7d6a](https://github.com/krisarmstrong/stem/commit/1cc7d6a5f3b19437659ee3d7c7aa2bbdcf1842e9)), closes [#285](https://github.com/krisarmstrong/stem/issues/285)

## [0.17.2](https://github.com/krisarmstrong/stem/compare/v0.17.1...v0.17.2) (2026-05-25)


### Bug Fixes

* **lint:** Clear gocognit, godoclint, nestif, tparallel ([#262](https://github.com/krisarmstrong/stem/issues/262)) ([19a2234](https://github.com/krisarmstrong/stem/commit/19a223489207abdcb3326a09190c9ef749301d7b))
* **lint:** Extract test-type + standard-name consts (partial goconst cleanup) ([#263](https://github.com/krisarmstrong/stem/issues/263)) ([9bcb8aa](https://github.com/krisarmstrong/stem/commit/9bcb8aac7feb4b53b19b1cabe673781db5b90698))
* **security:** Hardcode HTTPS-only auth, cipher overflow safety, fixture renames ([#1070](https://github.com/krisarmstrong/stem/issues/1070)) ([#260](https://github.com/krisarmstrong/stem/issues/260)) ([de2d166](https://github.com/krisarmstrong/stem/commit/de2d16677f25ac2a5fc1f9200a7f00703fcbec13))

## [0.17.1](https://github.com/krisarmstrong/stem/compare/v0.17.0...v0.17.1) (2026-05-22)


### Performance Improvements

* **e2e:** bump CI workers 1-&gt;2 and retries 2-&gt;1 ([#255](https://github.com/krisarmstrong/stem/issues/255)) ([6b8c658](https://github.com/krisarmstrong/stem/commit/6b8c65891f95b62c0a6b9200b22c3dc61739d5ef))

## [0.17.0](https://github.com/krisarmstrong/stem/compare/v0.16.0...v0.17.0) (2026-05-22)


### Features

* **theme:** add themeTypography barrel module (Phase 3) ([0f69005](https://github.com/krisarmstrong/stem/commit/0f690053c698696fe7bfc860b4b7690c4fcf5c1f))
* **theme:** adopt botanical-earth surface palette (Phase 4) ([d82ae9d](https://github.com/krisarmstrong/stem/commit/d82ae9d29a1f28b8d56dac4fc38746f9fae43549))
* **theme:** Apply 2026-05-22 brand audit — Stem becomes blue ([24576de](https://github.com/krisarmstrong/stem/commit/24576de60478f062cd23430bfe21c18848d3ec91))
* **theme:** fix button contrast against constant brand anchor (Phase 7) ([901eb9b](https://github.com/krisarmstrong/stem/commit/901eb9b04bb4797ddf9c96771102ace018b0505b))
* **theme:** identity shift — Stem becomes blue (Phase 5) ([0475681](https://github.com/krisarmstrong/stem/commit/04756815530f0854c8a580003ce06c7ab33ac28a))
* **theme:** self-host Inter + JetBrains Mono via [@fontsource-variable](https://github.com/fontsource-variable) (Phase 2) ([78459f0](https://github.com/krisarmstrong/stem/commit/78459f0e1eb58b146c4fb284dc66f23e246eb562))
* **theme:** swap status palette to canonical brand-tied anchors (Phase 1) ([40e298c](https://github.com/krisarmstrong/stem/commit/40e298c63daa676d2c3d8b66b070d6e0dd8c9d48))


### Bug Fixes

* **deps:** bump golang.org/x/net to v0.55.0 (GO-2026-5026) ([855f165](https://github.com/krisarmstrong/stem/commit/855f1659df1b4ade02bde6b1678de9705070db32))
* **deps:** Bump golang.org/x/net to v0.55.0 (GO-2026-5026) ([4011ac4](https://github.com/krisarmstrong/stem/commit/4011ac41a5598ce1268636d508ac224305c0e52d))
* **vite:** stop inlining font assets as data: URLs (CSP fix) ([2f3099f](https://github.com/krisarmstrong/stem/commit/2f3099fef8ed508bfc1fe1651a31aafa639d90c4))
* **vite:** Stop inlining font assets as data: URLs (CSP fix) ([96b4b8a](https://github.com/krisarmstrong/stem/commit/96b4b8a812dcaacb79907df73cc017755949e0c2))

## [0.16.0](https://github.com/krisarmstrong/stem/compare/v0.15.0...v0.16.0) (2026-05-22)


### Features

* **stories:** Primitive Storybook coverage + biome pin (Wave 5 / [#236](https://github.com/krisarmstrong/stem/issues/236)) ([#241](https://github.com/krisarmstrong/stem/issues/241)) ([b26dc80](https://github.com/krisarmstrong/stem/commit/b26dc804f04768ca20d85a5515d5f79d971fd308))
* **ui:** expand UI primitive barrel exports (Wave 5 / [#236](https://github.com/krisarmstrong/stem/issues/236)) ([#240](https://github.com/krisarmstrong/stem/issues/240)) ([798772b](https://github.com/krisarmstrong/stem/commit/798772b96fa9c2d954d1eac2982070d2f4123df1))

## [0.15.0](https://github.com/krisarmstrong/stem/compare/v0.14.0...v0.15.0) (2026-05-20)


### Features

* **auth:** argon2id password hashing + zxcvbn strength + hibp breach check ([#233](https://github.com/krisarmstrong/stem/issues/233)) ([4d85f83](https://github.com/krisarmstrong/stem/commit/4d85f83a626c25b07ae683365f98a0672c8957f8))
* **auth:** TOTP MFA + WebAuthn passkeys (Wave 3) ([#234](https://github.com/krisarmstrong/stem/issues/234)) ([91fcfac](https://github.com/krisarmstrong/stem/commit/91fcfacfdeebe2eadc81579cc0cf8ce7980991e9))
* **ci:** Add provenance_only mode for SLSA backfill ([#75](https://github.com/krisarmstrong/stem/issues/75)) ([#226](https://github.com/krisarmstrong/stem/issues/226)) ([04af510](https://github.com/krisarmstrong/stem/commit/04af510af5e4cd95b610e17c3179769fdaa18a53))
* tls by default + canonical port 8444 + http redirector + csrf fail-closed ([#232](https://github.com/krisarmstrong/stem/issues/232)) ([406bc43](https://github.com/krisarmstrong/stem/commit/406bc43d68675aa71b0828ec029523c385abe19e))
* **ui,api:** Reflector platform-guard + E2E cleanup of imaginary-UI specs ([#70](https://github.com/krisarmstrong/stem/issues/70) / [#64](https://github.com/krisarmstrong/stem/issues/64)) ([#224](https://github.com/krisarmstrong/stem/issues/224)) ([d765f62](https://github.com/krisarmstrong/stem/commit/d765f6224a2e0e302b579a71b19b94a70621c6e3))
* **ui,api:** Wire RoleChip to backend mode-switch endpoint ([#74](https://github.com/krisarmstrong/stem/issues/74)) ([#225](https://github.com/krisarmstrong/stem/issues/225)) ([cf69a9d](https://github.com/krisarmstrong/stem/commit/cf69a9d38feba0b8add742e8a808885dfa41f5e0))


### Bug Fixes

* **auth:** Serialise HIBP test seams behind a sync.RWMutex ([#235](https://github.com/krisarmstrong/stem/issues/235)) ([5f87f35](https://github.com/krisarmstrong/stem/commit/5f87f35a7f7e5358056e0adc9d7c54470df49cc1))
* **ci:** add target_tag input to SLSA backfill ([#75](https://github.com/krisarmstrong/stem/issues/75) follow-up) ([#228](https://github.com/krisarmstrong/stem/issues/228)) ([6e00400](https://github.com/krisarmstrong/stem/commit/6e0040087d2fdf81baddff14d5f544e2158ffa52))
* **ci:** unescape apostrophe in target_tag description ([#229](https://github.com/krisarmstrong/stem/issues/229)) ([e0c3d16](https://github.com/krisarmstrong/stem/commit/e0c3d16120d2265e050a1e5c5c7cbc31be5bc5c0))

## [0.14.0](https://github.com/krisarmstrong/stem/compare/v0.13.3...v0.14.0) (2026-05-19)


### Features

* Graceful port fallback when canonical port is in use ([#69](https://github.com/krisarmstrong/stem/issues/69)) ([#222](https://github.com/krisarmstrong/stem/issues/222)) ([750704b](https://github.com/krisarmstrong/stem/commit/750704b766b6e3d46be02de5628593196c0dacec))

## [0.13.3](https://github.com/krisarmstrong/stem/compare/v0.13.2...v0.13.3) (2026-05-19)


### Bug Fixes

* **ci:** point Lighthouse at the real served URLs ([#65](https://github.com/krisarmstrong/stem/issues/65)) ([#220](https://github.com/krisarmstrong/stem/issues/220)) ([cde7653](https://github.com/krisarmstrong/stem/commit/cde7653e76c771bcc8f497c0cba8cdd419f974ed))

## [0.13.2](https://github.com/krisarmstrong/stem/compare/v0.13.1...v0.13.2) (2026-05-18)


### Bug Fixes

* **api:** add SPA fallback for client-side routes ([#214](https://github.com/krisarmstrong/stem/issues/214)) ([ae5a51a](https://github.com/krisarmstrong/stem/commit/ae5a51aae68002b0b83f7f7624a2e423d765bef0))

## [0.13.1](https://github.com/krisarmstrong/stem/compare/v0.13.0...v0.13.1) (2026-05-18)


### Bug Fixes

* **ui,api:** replace hardcoded "0.1.0" with /__version + add the endpoint ([#212](https://github.com/krisarmstrong/stem/issues/212)) ([69fe359](https://github.com/krisarmstrong/stem/commit/69fe359dbaffcaf7f8a5fd73bd62a175ed9c0948))

## [0.13.0](https://github.com/krisarmstrong/stem/compare/v0.12.1...v0.13.0) (2026-05-18)


### Features

* **ui:** Flat sidebar + header role-chip + slimmed Settings + valid-interface filter ([#210](https://github.com/krisarmstrong/stem/issues/210)) ([1cb58bd](https://github.com/krisarmstrong/stem/commit/1cb58bd04693f1cd72597a3a1a868ecd504c8e19))

## [0.12.1](https://github.com/krisarmstrong/stem/compare/v0.12.0...v0.12.1) (2026-05-18)


### Bug Fixes

* **release:** Replace broken SLSA generator with attest-build-provenance ([#208](https://github.com/krisarmstrong/stem/issues/208)) ([4af33d0](https://github.com/krisarmstrong/stem/commit/4af33d0d4b56bcb02da8cdcd9babce8b09550088))

## [0.12.0](https://github.com/krisarmstrong/stem/compare/v0.11.0...v0.12.0) (2026-05-18)


### Features

* **ui:** lift primitive kit, add command palette, polish dark mode ([#206](https://github.com/krisarmstrong/stem/issues/206)) ([b4339de](https://github.com/krisarmstrong/stem/commit/b4339dee8b13f0bdec1db10b30a4309b238cfe49))

## [0.11.0](https://github.com/krisarmstrong/stem/compare/v0.10.0...v0.11.0) (2026-05-18)


### Features

* **make:** add capability-aware dev-run target ([#197](https://github.com/krisarmstrong/stem/issues/197)) ([ba3f344](https://github.com/krisarmstrong/stem/commit/ba3f344711064fe12a8dd5e21d0aa2aeca385eb6))
* product favicons + drop per-file copyright headers (SPDX for Go) ([#198](https://github.com/krisarmstrong/stem/issues/198)) ([faef765](https://github.com/krisarmstrong/stem/commit/faef765944195980af4c398dea22541cc0a0aedf))


### Bug Fixes

* **ci:** race detector needs C dataplane deps + serialize SSE tests ([#199](https://github.com/krisarmstrong/stem/issues/199)) ([34fad0d](https://github.com/krisarmstrong/stem/commit/34fad0d5337e9b1dc03315599d39c7dd4087d483))
* **tests:** gate remaining measure tests under -short ([#201](https://github.com/krisarmstrong/stem/issues/201)) ([b0fc1be](https://github.com/krisarmstrong/stem/commit/b0fc1be9382e540c9ae252445de392db22e7a696))
* **tests:** make race detector pass on Linux + CGO ([#200](https://github.com/krisarmstrong/stem/issues/200)) ([23cb945](https://github.com/krisarmstrong/stem/commit/23cb9458dd5328361591743b2ccb1de468308597))

## [0.10.0](https://github.com/krisarmstrong/stem/compare/v0.9.12...v0.10.0) (2026-05-18)


### Features

* **ui:** comprehensive tooltip parity — add ~42 tooltips for icon-only buttons + complex actions ([5a9ef39](https://github.com/krisarmstrong/stem/commit/5a9ef39aa0482871c77bd3cdecb612cb6d81927e))
* **ui:** phase A router + sidebar architecture (multi-page) ([207129b](https://github.com/krisarmstrong/stem/commit/207129b802ebe8212d281ad29033bc9f01647b1c))
* **ui:** port useTheme hook from seed for cross-repo parity ([a6d7494](https://github.com/krisarmstrong/stem/commit/a6d74945029ed4a9efc69d68edac5a013e29b2dd))


### Bug Fixes

* **ci:** rename status import to statusColor to avoid noShadow lint ([da4d3d9](https://github.com/krisarmstrong/stem/commit/da4d3d9de1535eb94d7c030e6352f5ce8c703c8d))
* **ci:** suppress biome noBarrelFile on intentional theme barrel ([ee76bd3](https://github.com/krisarmstrong/stem/commit/ee76bd3ac7de18181a02386e1d30f38f39078b38))

## [0.9.12](https://github.com/krisarmstrong/stem/compare/v0.9.11...v0.9.12) (2026-05-18)


### Bug Fixes

* **api:** update fs.Sub subdir to "ui" to match embed glob ([058d44f](https://github.com/krisarmstrong/stem/commit/058d44fdf297cb15b689eb3c5329260b98526460))
* **ci:** auto-trigger release-please on CI completion (was manual-only) ([5334db2](https://github.com/krisarmstrong/stem/commit/5334db21fa76875e2a7ded4a24e14a8a52f31147))
* **ci:** bump Dockerfile go-build to golang:1.26-bookworm ([032a37e](https://github.com/krisarmstrong/stem/commit/032a37e2d50e3d774469132756532ee783eaae38))
* **ci:** correct artifact path + Docker [@locales](https://github.com/locales) copy ([b4902e4](https://github.com/krisarmstrong/stem/commit/b4902e4ac2ae194aa06925c48fab173c33f74804))
* **metrics:** serialize tests that share Prometheus counter labels ([3e413bc](https://github.com/krisarmstrong/stem/commit/3e413bc196564221a31f5a4ced920cc446623e15))

## [0.9.11](https://github.com/krisarmstrong/stem/compare/v0.9.10...v0.9.11) (2026-05-14)


### Bug Fixes

* **build:** expose linux feature APIs for c23 ([ef93e2a](https://github.com/krisarmstrong/stem/commit/ef93e2ad74b7080d8a30e0e334c776bb7e0593d6))
* **ci:** align container and license validation ([655c917](https://github.com/krisarmstrong/stem/commit/655c9171e8194e45c76d2a499a07353c638942e7))
* **ci:** allow gitleaks to inspect pull requests ([cd5728a](https://github.com/krisarmstrong/stem/commit/cd5728a6ccf84af1c460a518186e8df59f1c15cd))
* **ci:** allow MPL npm dependencies ([5b03f31](https://github.com/krisarmstrong/stem/commit/5b03f3139d72c6a18b6dd8efe202221c9c07821f))
* **ci:** build browser test server without cgo ([46d3a3b](https://github.com/krisarmstrong/stem/commit/46d3a3ba31a1bdd77d1fbc434f42f6b9f4767242))
* **ci:** build stem native library with clang ([59f46a0](https://github.com/krisarmstrong/stem/commit/59f46a0fa7d6bef2a24e6f5558b27fd03b2c15ca))
* **ci:** build stem native test dependencies ([dfb6d45](https://github.com/krisarmstrong/stem/commit/dfb6d45d0128dfc2f31aa38347dd4fddeb0e2818))
* **ci:** fetch full history for security scans ([655c135](https://github.com/krisarmstrong/stem/commit/655c135c05b9d7c025cc1138bbd1f3826932acb9))
* **ci:** handle missing dataplane contexts ([8736134](https://github.com/krisarmstrong/stem/commit/8736134b10b1a8a23a23d9b2007bad41ed7dac2f))
* **ci:** keep stem analysis advisory ([74f779e](https://github.com/krisarmstrong/stem/commit/74f779e0de00fa7bd4c2fef92f0bed0cce4347ac))
* **ci:** link native dataplane tests ([b6da226](https://github.com/krisarmstrong/stem/commit/b6da22688638460abb5b2279024cfcf1b00793b8))
* **ci:** repair buildpacks project metadata ([cdcb63f](https://github.com/krisarmstrong/stem/commit/cdcb63f4965cc080cae68daa7b9be0fd7d0033f0))
* **ci:** repair label sync workflow ([7acb464](https://github.com/krisarmstrong/stem/commit/7acb4647a4eb80d138f01a10a5a3b113bebaae40))
* **ci:** report stem analyzer findings ([d726b50](https://github.com/krisarmstrong/stem/commit/d726b501d973ee8fbf1bda2975d9ed13ff7feb48))
* **ci:** resolve stem workflow blockers ([314785d](https://github.com/krisarmstrong/stem/commit/314785d6c3f3a0f763e3758b3ba64fffdddf50c5))
* **ci:** restore stem validation pipeline ([c1a26b2](https://github.com/krisarmstrong/stem/commit/c1a26b20afce1f59e5a0b694d263d62860b1c41f))
* **ci:** run stub unit tests without race ([6272714](https://github.com/krisarmstrong/stem/commit/62727147bada8993d1ce1682e64925c09aee02b6))
* **ci:** run stem intel macos release on current runner ([7f9d234](https://github.com/krisarmstrong/stem/commit/7f9d23427a7a4466b8626f6b6d8ee76179df6f10))
* **ci:** satisfy servicetest lint ([ec275df](https://github.com/krisarmstrong/stem/commit/ec275df79aa63360ee069f492469d13c6633fc70))
* **ci:** scope stem container and license checks ([d267154](https://github.com/krisarmstrong/stem/commit/d2671547ae280830d09777768d5635d58721dfd6))
* **ci:** scope stem e2e smoke suite ([4ce2153](https://github.com/krisarmstrong/stem/commit/4ce2153966bff419ad4fb47f75edbd336db2c9a9))
* **ci:** skip stem docker publish without dockerfile ([a5a9deb](https://github.com/krisarmstrong/stem/commit/a5a9deb1064f7ee462c400b3e3138918940e2a20))
* **ci:** split native compile from unit tests ([f1f8c82](https://github.com/krisarmstrong/stem/commit/f1f8c82c6be3026e969a8917cffc075841eafeba))
* **ci:** stabilize automated validation ([76209fa](https://github.com/krisarmstrong/stem/commit/76209faef490df7baa09d161222ec7fc5da838e8))
* **ci:** stabilize stem browser smoke gate ([7dc7655](https://github.com/krisarmstrong/stem/commit/7dc765542a92fcd6465aa1f483e19aadea440ab1))
* **ci:** start stem web server in browser jobs ([2c9f44b](https://github.com/krisarmstrong/stem/commit/2c9f44b0c29dc60aaf97345a781a9748355defac))
* **ci:** use compatible labeler action ([99c9c57](https://github.com/krisarmstrong/stem/commit/99c9c57eab8ee28c0a69d6a1570046cd6b49c596))
* **ci:** use hosted node setup in container workflow ([9023b15](https://github.com/krisarmstrong/stem/commit/9023b15e74f79c4d929145aaca4dd1067da8b718))
* **ci:** use labeler yaml format ([8d68517](https://github.com/krisarmstrong/stem/commit/8d6851793528dd8862dd6c5bd9fde29866b485b2))
* **security:** scope generated TLS certificate writes ([83f6cef](https://github.com/krisarmstrong/stem/commit/83f6cef51e216a8c2a9b7c6e713fc064541de697))

## [Unreleased]

## [0.1.13] - 2026-01-04

### Changed

- Standardize branding to use "The Stem" in CLI and documentation headings.

## [0.1.12] - 2026-01-04

### Added

- Wire RFC 2889, RFC 6349, Y.1731, MEF, TSN, and custom stream configs into the dataplane wrapper.

### Changed

- Route Measure, TrafficGen, ServiceTest, and Certify executors through the dataplane API.
- Update module status documentation to reflect implemented test execution.

## [0.1.11] - 2026-01-04

### Changed

- Document the SSE-based UI transport for real-time updates.

## [0.1.10] - 2026-01-04

### Added

- Document current API with a Target API vNext section.

### Fixed

- Avoid inline error handling in writeJSON to satisfy lint rules.

### Changed

- Allow golangci-lint parallel runners in Makefile.

## [0.1.0] - 2025-12-30

### Added

- Initial project structure
- Module-oriented architecture (Benchmark, ServiceTest, TrafficGen, Measure, Certify)
- Reflector mode (Tier 1)
- RFC 2544 test support (throughput, latency, frame loss, back-to-back)
- ITU-T Y.1564 service activation testing
- CLI interface with `stem` binary
- WebUI with React/TypeScript
- TUI dashboard
- License management system (Tier 1/2/3)
- Go 1.25+ backend
- C23 dataplane with AF_PACKET support
- Biome linting for TypeScript
- golangci-lint for Go

### Infrastructure

- Makefile build system
- Development documentation
- CLAUDE.md for AI-assisted development

---

For detailed commit history, see: https://github.com/krisarmstrong/stem/commits/main
