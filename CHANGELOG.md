# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0](https://github.com/MustardSeedNetworks/stem/compare/v1.0.0...v2.0.0) (2026-07-09)


### ⚠ BREAKING CHANGES

* **i18n:** Translation keys changed from flat strings to nested objects

### Features

* **a11y:** axe-core test harness for the Button primitives ([#370](https://github.com/MustardSeedNetworks/stem/issues/370)) ([991c92f](https://github.com/MustardSeedNetworks/stem/commit/991c92f9fbf572a6d2278e1315a3995f0a1cb29b))
* achieve 100% config parity between CLI, TUI, and WebUI ([e42fe2a](https://github.com/MustardSeedNetworks/stem/commit/e42fe2af1ba9702002c3c10b87f1f712959272b3))
* add _project_specs session management structure ([b8f3373](https://github.com/MustardSeedNetworks/stem/commit/b8f3373d68c4d2ad03c7a572a73a8658641b5c2e))
* add Cloud Native Buildpacks support ([1ffde0f](https://github.com/MustardSeedNetworks/stem/commit/1ffde0f9b0b61a08f65f8880b841126e3605f998))
* add lighthouse, i18n validation, and package targets ([3043e6b](https://github.com/MustardSeedNetworks/stem/commit/3043e6b086fc6a28b9bf2da5fd6ec5621c38753d))
* add packaging infrastructure (DEB, RPM, PKG) ([47f016c](https://github.com/MustardSeedNetworks/stem/commit/47f016cba316dd111ff8becaf2c2818591600023))
* add Phase 2-5 production readiness improvements ([eaab585](https://github.com/MustardSeedNetworks/stem/commit/eaab58515a819e6183ae6034d25320f76b631e8a))
* add Phase 3 documentation and i18n support ([6a69e5b](https://github.com/MustardSeedNetworks/stem/commit/6a69e5b42a61f1a54ed12b57e2c02206eedfc9f9))
* add Phase 3 operational readiness improvements ([ac2533b](https://github.com/MustardSeedNetworks/stem/commit/ac2533b376a6f87eab42804b614fa8000ab35a41))
* add production readiness features and WebUI improvements ([366865d](https://github.com/MustardSeedNetworks/stem/commit/366865d80c8e56d2400fd8e557b4168320a67862))
* add test parallelization, build profiles, and iperf3 bundling ([f42bf7a](https://github.com/MustardSeedNetworks/stem/commit/f42bf7a63b3db4744dc33ac74257bb407ff42ace))
* add tests/integration/ directory ([13b6e8d](https://github.com/MustardSeedNetworks/stem/commit/13b6e8d6c98486e2d9582c78aab9cb5d733f7798))
* add websocket streaming and api auth ([69c569b](https://github.com/MustardSeedNetworks/stem/commit/69c569bc61924bcd26a9d9adbda05f0501d05b26))
* **api:** add go-playground/validator + tags on hot DTOs ([#294](https://github.com/MustardSeedNetworks/stem/issues/294)) ([95236b5](https://github.com/MustardSeedNetworks/stem/commit/95236b564fb25312a433fef35b3f6a453d469e33))
* **api:** capability registry for declarative route policy (register) ([#401](https://github.com/MustardSeedNetworks/stem/issues/401)) ([55ae19a](https://github.com/MustardSeedNetworks/stem/commit/55ae19ab9578eaadcbd6fc9f8346a25982e48ffd))
* **api:** complete frontend-to-backend config plumbing ([7b8b2bc](https://github.com/MustardSeedNetworks/stem/commit/7b8b2bc62750e451319313a6f0cd0c9486c65ce6))
* **api:** make request body-size limit declarative in the route registry ([#415](https://github.com/MustardSeedNetworks/stem/issues/415)) ([1aa020e](https://github.com/MustardSeedNetworks/stem/commit/1aa020e17072d28c11b4fe8d52a3c693440869cc))
* **api:** port invopop/jsonschema generator from NIAC ([#297](https://github.com/MustardSeedNetworks/stem/issues/297)) ([d489eef](https://github.com/MustardSeedNetworks/stem/commit/d489eef332981b2b18ecdf2b30afd50385ebb4d4)), closes [#269](https://github.com/MustardSeedNetworks/stem/issues/269)
* **api:** port seed's strict JSON decode helpers ([#320](https://github.com/MustardSeedNetworks/stem/issues/320)) ([#322](https://github.com/MustardSeedNetworks/stem/issues/322)) ([9499f86](https://github.com/MustardSeedNetworks/stem/commit/9499f86612c168fad10c4110c581f4bd28415abf))
* **api:** stream live test progress + reflector stats + mode broadcasts via SSE ([#302](https://github.com/MustardSeedNetworks/stem/issues/302)) ([6dd72ff](https://github.com/MustardSeedNetworks/stem/commit/6dd72ff2b97c05181188f35daac29cfa1c144008)), closes [#296](https://github.com/MustardSeedNetworks/stem/issues/296)
* **auth:** add ACME/Let's Encrypt, OAuth SSO, and UserStore from Seed ([966fd91](https://github.com/MustardSeedNetworks/stem/commit/966fd914759f63aa63e35dca4751d6d05fc15e59))
* **auth:** argon2id password hashing + zxcvbn strength + hibp breach check ([#233](https://github.com/MustardSeedNetworks/stem/issues/233)) ([4d85f83](https://github.com/MustardSeedNetworks/stem/commit/4d85f83a626c25b07ae683365f98a0672c8957f8))
* **auth:** TOTP MFA + WebAuthn passkeys (Wave 3) ([#234](https://github.com/MustardSeedNetworks/stem/issues/234)) ([91fcfac](https://github.com/MustardSeedNetworks/stem/commit/91fcfacfdeebe2eadc81579cc0cf8ce7980991e9))
* **ci:** add frontend artifact sharing and todo tracker ([2ddbf1a](https://github.com/MustardSeedNetworks/stem/commit/2ddbf1aa6f708555cdc7f358d249a7e95f28cf7e))
* **ci:** Add provenance_only mode for SLSA backfill ([#75](https://github.com/MustardSeedNetworks/stem/issues/75)) ([#226](https://github.com/MustardSeedNetworks/stem/issues/226)) ([04af510](https://github.com/MustardSeedNetworks/stem/commit/04af510af5e4cd95b610e17c3179769fdaa18a53))
* **ci:** add unsafe logging pattern detection ([43ccf06](https://github.com/MustardSeedNetworks/stem/commit/43ccf064b1f34afe0a52fe760cd226b66e61f833))
* **ci:** pin actions to v6 and enhance security scanning ([488132b](https://github.com/MustardSeedNetworks/stem/commit/488132be2ef261612801b6cd8632a5bc7bcdccfe))
* **cli:** help entries for tui/list-tests/install-ca + completeness test ([#371](https://github.com/MustardSeedNetworks/stem/issues/371)) ([85d56a1](https://github.com/MustardSeedNetworks/stem/commit/85d56a1b00ff71e5693906eecca4bfb866414b52))
* close all UI parity gaps between CLI, TUI, and WebUI ([317803c](https://github.com/MustardSeedNetworks/stem/commit/317803c15d805df5b8ef77bd3282d8b3920e12b9))
* **config:** make biome.json significantly stricter ([4ce8338](https://github.com/MustardSeedNetworks/stem/commit/4ce83380ffb21223cc439a067ddd6237466e9b62))
* configure C23 linting and formatting (NOT C++) ([c0337e4](https://github.com/MustardSeedNetworks/stem/commit/c0337e4c4d8de5ffc3f12ed041d364587844d689))
* **database:** add SQLite database layer with repositories ([311952f](https://github.com/MustardSeedNetworks/stem/commit/311952f287e945427b15d0cf0b16c6bb38f69b40))
* **forms:** adopt react-hook-form + valibot resolver, Y.1564 pilot ([#325](https://github.com/MustardSeedNetworks/stem/issues/325)) ([#328](https://github.com/MustardSeedNetworks/stem/issues/328)) ([bf7937a](https://github.com/MustardSeedNetworks/stem/commit/bf7937aa0dfa062469e7ace706c3dbc2254c944c))
* **forms:** adopt react-hook-form for MFA setup + disable ([#325](https://github.com/MustardSeedNetworks/stem/issues/325)) ([#331](https://github.com/MustardSeedNetworks/stem/issues/331)) ([92cffef](https://github.com/MustardSeedNetworks/stem/commit/92cffef7240aad661ece2130fe8006f9670cab40))
* **forms:** finish react-hook-form rollout — App.tsx auth + Recovery + Setup ([#332](https://github.com/MustardSeedNetworks/stem/issues/332)) ([#333](https://github.com/MustardSeedNetworks/stem/issues/333)) ([2afb218](https://github.com/MustardSeedNetworks/stem/commit/2afb218a9a40e030cf89dd3c3b2b8f40d84ff1f0))
* **forms:** sweep 6 remaining ConfigForms onto react-hook-form ([#325](https://github.com/MustardSeedNetworks/stem/issues/325)) ([#330](https://github.com/MustardSeedNetworks/stem/issues/330)) ([224b39d](https://github.com/MustardSeedNetworks/stem/commit/224b39d5d5e4d7f873bdce2162db0502ee16e2e7))
* Graceful port fallback when canonical port is in use ([#69](https://github.com/MustardSeedNetworks/stem/issues/69)) ([#222](https://github.com/MustardSeedNetworks/stem/issues/222)) ([750704b](https://github.com/MustardSeedNetworks/stem/commit/750704b766b6e3d46be02de5628593196c0dacec))
* **help:** surface version in HelpDrawer header (in-app About) ([#372](https://github.com/MustardSeedNetworks/stem/issues/372)) ([86b76b5](https://github.com/MustardSeedNetworks/stem/commit/86b76b5669d022b8dfe95a087d19421f727312f5))
* **i18n:** add errors.license.* keys for tier-gating UI ([#313](https://github.com/MustardSeedNetworks/stem/issues/313)) ([0e406de](https://github.com/MustardSeedNetworks/stem/commit/0e406de96680b32f3dcdf54a794d9331ec8b47cd))
* **i18n:** add per-repo dynamic-prefixes allowlist for check-keys.py ([#335](https://github.com/MustardSeedNetworks/stem/issues/335)) ([cd6a793](https://github.com/MustardSeedNetworks/stem/commit/cd6a793e5faf13b5d4efc93d8c2b94de8564c9f5))
* **i18n:** add TypeScript type safety for translations ([0720b05](https://github.com/MustardSeedNetworks/stem/commit/0720b05f00d9338376d672554b2ee97578ce15fa))
* **i18n:** add useLocale hook + migrate HelpDrawer plurals ([#324](https://github.com/MustardSeedNetworks/stem/issues/324)) ([771a07c](https://github.com/MustardSeedNetworks/stem/commit/771a07c091b79223bef45d00b0987ec4ff8c394b))
* **i18n:** en/es key parity + DNT compliance test (stem) ([#373](https://github.com/MustardSeedNetworks/stem/issues/373)) ([f36d954](https://github.com/MustardSeedNetworks/stem/commit/f36d95410515398467e17f9d72c23bb5b2bde15b))
* **i18n:** expand settings translations for 95%+ coverage ([#132](https://github.com/MustardSeedNetworks/stem/issues/132)) ([316aa7e](https://github.com/MustardSeedNetworks/stem/commit/316aa7e5d9db7a6b228651ef14e608c18a6880a2))
* **i18n:** extract ModuleCard a11y strings to locale keys ([#319](https://github.com/MustardSeedNetworks/stem/issues/319)) ([56eacd5](https://github.com/MustardSeedNetworks/stem/commit/56eacd59917218651c9d179067a7b2709ea2d3ee))
* **i18n:** port check-keys.py + add phase 6 unit tests ([#327](https://github.com/MustardSeedNetworks/stem/issues/327)) ([85a3382](https://github.com/MustardSeedNetworks/stem/commit/85a3382a3597b893223fe9e0461a740919046fe8))
* **i18n:** unified frontend/backend i18n with JSON single source of truth ([40a74d0](https://github.com/MustardSeedNetworks/stem/commit/40a74d0334b7859804e58f506336933a4703f688))
* implement dataplane-backed certification and measurement tests ([d01e4bd](https://github.com/MustardSeedNetworks/stem/commit/d01e4bdb022a7faa7ea15f6f586a71d14d536e68))
* implement versioning and CI/CD infrastructure ([e72eafb](https://github.com/MustardSeedNetworks/stem/commit/e72eafb72dd4a9f3aa54b64e83a3063ef67ea9f4))
* **license:** replace forgeable rotor cipher with Ed25519-signed tokens ([#409](https://github.com/MustardSeedNetworks/stem/issues/409)) ([b78e407](https://github.com/MustardSeedNetworks/stem/commit/b78e407a5499f0398d92c32de03ef0f5ae5a44f7))
* **make:** add capability-aware dev-run target ([#197](https://github.com/MustardSeedNetworks/stem/issues/197)) ([ba3f344](https://github.com/MustardSeedNetworks/stem/commit/ba3f344711064fe12a8dd5e21d0aa2aeca385eb6))
* **makefile:** add deployment automation targets ([204f80c](https://github.com/MustardSeedNetworks/stem/commit/204f80c7af585f0e8e161d00dd632bfb621397f2))
* **makefile:** add developer experience enhancements ([726f0fb](https://github.com/MustardSeedNetworks/stem/commit/726f0fb4d0b36c502774d61534b1d6f593e61a56))
* **makefile:** add testing and quality enhancements ([b76708f](https://github.com/MustardSeedNetworks/stem/commit/b76708f1fa2205abc6c8fdedfc815704810df47e))
* **modules:** add 6-module architecture with Reflector separation ([5f7cd6f](https://github.com/MustardSeedNetworks/stem/commit/5f7cd6fc6c80e04006560b79a46af4ceb4930fd8))
* **platform:** add platform detection with improved error messages ([#56](https://github.com/MustardSeedNetworks/stem/issues/56)) ([8187ce7](https://github.com/MustardSeedNetworks/stem/commit/8187ce7b6ce2a7fb21186b0acbbfd4447f0c99d3))
* product favicons + drop per-file copyright headers (SPDX for Go) ([#198](https://github.com/MustardSeedNetworks/stem/issues/198)) ([faef765](https://github.com/MustardSeedNetworks/stem/commit/faef765944195980af4c398dea22541cc0a0aedf))
* **security:** add CSRF protection, setup wizard, and password recovery ([2769f21](https://github.com/MustardSeedNetworks/stem/commit/2769f2118fc37b6b82c72762309beac3e61cac07))
* **security:** add RFC 1918 CORS validation and fix CSP ([71f6da4](https://github.com/MustardSeedNetworks/stem/commit/71f6da4c15b026bfb61c2a852de68bf2ebf89375))
* **stories:** Primitive Storybook coverage + biome pin (Wave 5 / [#236](https://github.com/MustardSeedNetworks/stem/issues/236)) ([#241](https://github.com/MustardSeedNetworks/stem/issues/241)) ([b26dc80](https://github.com/MustardSeedNetworks/stem/commit/b26dc804f04768ca20d85a5515d5f79d971fd308))
* **storybook:** add Storybook 10.x with component stories ([#131](https://github.com/MustardSeedNetworks/stem/issues/131)) ([c5fff23](https://github.com/MustardSeedNetworks/stem/commit/c5fff239251b987af5266ef0434982a9e3e26f57))
* **theme:** add themeTypography barrel module (Phase 3) ([0f69005](https://github.com/MustardSeedNetworks/stem/commit/0f690053c698696fe7bfc860b4b7690c4fcf5c1f))
* **theme:** adopt botanical-earth surface palette (Phase 4) ([d82ae9d](https://github.com/MustardSeedNetworks/stem/commit/d82ae9d29a1f28b8d56dac4fc38746f9fae43549))
* **theme:** Apply 2026-05-22 brand audit — Stem becomes blue ([24576de](https://github.com/MustardSeedNetworks/stem/commit/24576de60478f062cd23430bfe21c18848d3ec91))
* **theme:** fix button contrast against constant brand anchor (Phase 7) ([901eb9b](https://github.com/MustardSeedNetworks/stem/commit/901eb9b04bb4797ddf9c96771102ace018b0505b))
* **theme:** identity shift — Stem becomes blue (Phase 5) ([0475681](https://github.com/MustardSeedNetworks/stem/commit/04756815530f0854c8a580003ce06c7ab33ac28a))
* **theme:** self-host Inter + JetBrains Mono via [@fontsource-variable](https://github.com/fontsource-variable) (Phase 2) ([78459f0](https://github.com/MustardSeedNetworks/stem/commit/78459f0e1eb58b146c4fb284dc66f23e246eb562))
* **theme:** swap status palette to canonical brand-tied anchors (Phase 1) ([40e298c](https://github.com/MustardSeedNetworks/stem/commit/40e298c63daa676d2c3d8b66b070d6e0dd8c9d48))
* tls by default + canonical port 8444 + http redirector + csrf fail-closed ([#232](https://github.com/MustardSeedNetworks/stem/issues/232)) ([406bc43](https://github.com/MustardSeedNetworks/stem/commit/406bc43d68675aa71b0828ec029523c385abe19e))
* **tui:** add Reflector TUI filter profiles and keyboard shortcuts ([4448f8c](https://github.com/MustardSeedNetworks/stem/commit/4448f8cec06153c03457cc50712d7931902acaf8)), closes [#82](https://github.com/MustardSeedNetworks/stem/issues/82)
* **types:** add profile types aligned with Seed patterns ([b9f2f59](https://github.com/MustardSeedNetworks/stem/commit/b9f2f599984ee0c4aab686512b8ee9b5881b4360))
* **ui,api:** Reflector platform-guard + E2E cleanup of imaginary-UI specs ([#70](https://github.com/MustardSeedNetworks/stem/issues/70) / [#64](https://github.com/MustardSeedNetworks/stem/issues/64)) ([#224](https://github.com/MustardSeedNetworks/stem/issues/224)) ([d765f62](https://github.com/MustardSeedNetworks/stem/commit/d765f6224a2e0e302b579a71b19b94a70621c6e3))
* **ui,api:** Wire RoleChip to backend mode-switch endpoint ([#74](https://github.com/MustardSeedNetworks/stem/issues/74)) ([#225](https://github.com/MustardSeedNetworks/stem/issues/225)) ([cf69a9d](https://github.com/MustardSeedNetworks/stem/commit/cf69a9d38feba0b8add742e8a808885dfa41f5e0))
* **ui:** add ARIA labels for accessibility ([#111](https://github.com/MustardSeedNetworks/stem/issues/111)) ([2610fdb](https://github.com/MustardSeedNetworks/stem/commit/2610fdba9f1446e4d1a1756a9f29d99a4c587d58))
* **ui:** add canonical help-drawer testids (Phase 3b) ([#354](https://github.com/MustardSeedNetworks/stem/issues/354)) ([f2a1518](https://github.com/MustardSeedNetworks/stem/commit/f2a15185740909e27b53e8b7bcb3ce851df9775f))
* **ui:** add focus traps to modals for accessibility ([#110](https://github.com/MustardSeedNetworks/stem/issues/110)) ([05810e6](https://github.com/MustardSeedNetworks/stem/commit/05810e683173e3646d5ecf16a56c9e4b86875135))
* **ui:** add ModuleCard component with per-test results display ([c61697a](https://github.com/MustardSeedNetworks/stem/commit/c61697a6baa43e48923f3a6ff3057e81c9f17e92))
* **ui:** add profile store, i18n test sections, and more stories ([8336c50](https://github.com/MustardSeedNetworks/stem/commit/8336c507599ba66166a9a34da916e3693f8ec59b))
* **ui:** add React Error Boundary for crash recovery ([#109](https://github.com/MustardSeedNetworks/stem/issues/109)) ([efbaf2a](https://github.com/MustardSeedNetworks/stem/commit/efbaf2a5a024ce77e7940d6d8a8e546bd5865061))
* **ui:** add React Query + migrate the interfaces read (W5.2a) ([#431](https://github.com/MustardSeedNetworks/stem/issues/431)) ([12e073c](https://github.com/MustardSeedNetworks/stem/commit/12e073ccda2ebe9a54afb2a4a75be22f39d7e3e6))
* **ui:** add types/generated directory for backend types ([18feccc](https://github.com/MustardSeedNetworks/stem/commit/18fecccc88642819c92f73c3c95980ff67dbad6a))
* **ui:** canonical shell — modernize Sidebar + PageHeader (Phase 1) ([#339](https://github.com/MustardSeedNetworks/stem/issues/339)) ([3983689](https://github.com/MustardSeedNetworks/stem/commit/39836892a591fd847c615df8616a181fb22ddd6c))
* **ui:** comprehensive tooltip parity — add ~42 tooltips for icon-only buttons + complex actions ([5a9ef39](https://github.com/MustardSeedNetworks/stem/commit/5a9ef39aa0482871c77bd3cdecb612cb6d81927e))
* **ui:** converge settings drawer shell — a11y + testids (Phase 3c) ([#353](https://github.com/MustardSeedNetworks/stem/issues/353)) ([6138e4b](https://github.com/MustardSeedNetworks/stem/commit/6138e4b372d015c29400a701c974858d16456ef8))
* **ui:** expand UI primitive barrel exports (Wave 5 / [#236](https://github.com/MustardSeedNetworks/stem/issues/236)) ([#240](https://github.com/MustardSeedNetworks/stem/issues/240)) ([798772b](https://github.com/MustardSeedNetworks/stem/commit/798772b96fa9c2d954d1eac2982070d2f4123df1))
* **ui:** extract auth to auth-store + harden authFetch (W5.3) ([#434](https://github.com/MustardSeedNetworks/stem/issues/434)) ([7075ea6](https://github.com/MustardSeedNetworks/stem/commit/7075ea670dd28b8b0b7fe865e13d17fb61005ae6))
* **ui:** Flat sidebar + header role-chip + slimmed Settings + valid-interface filter ([#210](https://github.com/MustardSeedNetworks/stem/issues/210)) ([1cb58bd](https://github.com/MustardSeedNetworks/stem/commit/1cb58bd04693f1cd72597a3a1a868ecd504c8e19))
* **ui:** generate TypeScript types from JSON Schemas ([#299](https://github.com/MustardSeedNetworks/stem/issues/299)) ([a4ea7d5](https://github.com/MustardSeedNetworks/stem/commit/a4ea7d56096ccbca464c31c942ec08b53ad2b8d0))
* **ui:** harden RoleContext with valibot schemas ([#295](https://github.com/MustardSeedNetworks/stem/issues/295)) ([ef93f25](https://github.com/MustardSeedNetworks/stem/commit/ef93f251f5a0dbd3dc0028f372b33412f5ef956a)), closes [#272](https://github.com/MustardSeedNetworks/stem/issues/272)
* **ui:** harmonize Card system with Seed patterns ([e11f23c](https://github.com/MustardSeedNetworks/stem/commit/e11f23c709fc8d73cc04a542bc3e57f05d1e66b4))
* **ui:** harmonize RecoveryForm with Seed patterns ([ea6ef1f](https://github.com/MustardSeedNetworks/stem/commit/ea6ef1fe76ce3d637445a242689ab5157a8be021))
* **ui:** harmonize SetupWizard with Seed patterns ([a24ab49](https://github.com/MustardSeedNetworks/stem/commit/a24ab497fcf625d1bc983f5285f5ff4df0572268))
* **ui:** harmonize UI components with Seed patterns ([#122](https://github.com/MustardSeedNetworks/stem/issues/122)-[#129](https://github.com/MustardSeedNetworks/stem/issues/129)) ([76a357f](https://github.com/MustardSeedNetworks/stem/commit/76a357f226fa7e20926f5c5fb407aa4d9be5566b))
* **ui:** lift primitive kit, add command palette, polish dark mode ([#206](https://github.com/MustardSeedNetworks/stem/issues/206)) ([b4339de](https://github.com/MustardSeedNetworks/stem/commit/b4339dee8b13f0bdec1db10b30a4309b238cfe49))
* **ui:** migrate the stats poll to React Query (W5.2b) ([#432](https://github.com/MustardSeedNetworks/stem/issues/432)) ([4f17830](https://github.com/MustardSeedNetworks/stem/commit/4f17830bd0b99a95482915cf87657eb4aca8a707))
* **ui:** phase A router + sidebar architecture (multi-page) ([207129b](https://github.com/MustardSeedNetworks/stem/commit/207129b802ebe8212d281ad29033bc9f01647b1c))
* **ui:** port useTheme hook from seed for cross-repo parity ([a6d7494](https://github.com/MustardSeedNetworks/stem/commit/a6d74945029ed4a9efc69d68edac5a013e29b2dd))
* **ui:** require auth and add live stats ([46b01a7](https://github.com/MustardSeedNetworks/stem/commit/46b01a7b6ab006ed3dff4e302dd1353a6cf6f9e6))
* **web:** route test execution through module system ([b2b9d22](https://github.com/MustardSeedNetworks/stem/commit/b2b9d225543d18826f31e4bdf0356d24eb576ff7))


### Bug Fixes

* add missing braces and fix data race in test executor ([aaf8b42](https://github.com/MustardSeedNetworks/stem/commit/aaf8b4224dc37a21d6f65820beba01725a80c405))
* add missing help-content data file and @types/node dependency ([f0e3d12](https://github.com/MustardSeedNetworks/stem/commit/f0e3d1221823fda34a302b80608e493d2cd7d72c))
* address golangci-lint issues across codebase ([26075d0](https://github.com/MustardSeedNetworks/stem/commit/26075d0e234f02a1eb145c5a5999bf7d2c5d924f))
* address remaining code quality issues from Codex review ([cb1997a](https://github.com/MustardSeedNetworks/stem/commit/cb1997a4bf8d989d67826ca1cde1fb36e08e9fdf))
* align dataplane MEF results and docs ([7e7286a](https://github.com/MustardSeedNetworks/stem/commit/7e7286ab03261be83299961e3a26d2babec61da1))
* **api:** add SPA fallback for client-side routes ([#214](https://github.com/MustardSeedNetworks/stem/issues/214)) ([ae5a51a](https://github.com/MustardSeedNetworks/stem/commit/ae5a51aae68002b0b83f7f7624a2e423d765bef0))
* **api:** authenticate reflector config/stats + pin govulncheck ([#398](https://github.com/MustardSeedNetworks/stem/issues/398)) ([7d9a4e3](https://github.com/MustardSeedNetworks/stem/commit/7d9a4e31b7f2d3c6d60b9899c067da5aa7f803c8))
* **api:** make RFC1918 CORS reflection opt-in (STEM_CORS_ALLOW_PRIVATE) ([#399](https://github.com/MustardSeedNetworks/stem/issues/399)) ([554fc76](https://github.com/MustardSeedNetworks/stem/commit/554fc763e90b0d5b546e9e60ddf85e1c5a551259))
* **api:** update fs.Sub subdir to "ui" to match embed glob ([058d44f](https://github.com/MustardSeedNetworks/stem/commit/058d44fdf297cb15b689eb3c5329260b98526460))
* **auth:** remove t.Parallel from HIBP tests that mutate shared endpoint ([#315](https://github.com/MustardSeedNetworks/stem/issues/315)) ([08ef121](https://github.com/MustardSeedNetworks/stem/commit/08ef121f47546d52e846afb741e51f58ab32dee9))
* **auth:** Serialise HIBP test seams behind a sync.RWMutex ([#235](https://github.com/MustardSeedNetworks/stem/issues/235)) ([5f87f35](https://github.com/MustardSeedNetworks/stem/commit/5f87f35a7f7e5358056e0adc9d7c54470df49cc1))
* **build:** expose linux feature APIs for c23 ([ef93e2a](https://github.com/MustardSeedNetworks/stem/commit/ef93e2ad74b7080d8a30e0e334c776bb7e0593d6))
* **build:** fix packaging targets and add build-darwin/build-linux ([faefd65](https://github.com/MustardSeedNetworks/stem/commit/faefd652111c2467948a9042994b1dc5bcb6038f))
* **ci:** add target_tag input to SLSA backfill ([#75](https://github.com/MustardSeedNetworks/stem/issues/75) follow-up) ([#228](https://github.com/MustardSeedNetworks/stem/issues/228)) ([6e00400](https://github.com/MustardSeedNetworks/stem/commit/6e0040087d2fdf81baddff14d5f544e2158ffa52))
* **ci:** align container and license validation ([655c917](https://github.com/MustardSeedNetworks/stem/commit/655c9171e8194e45c76d2a499a07353c638942e7))
* **ci:** allow gitleaks to inspect pull requests ([cd5728a](https://github.com/MustardSeedNetworks/stem/commit/cd5728a6ccf84af1c460a518186e8df59f1c15cd))
* **ci:** allow MPL npm dependencies ([5b03f31](https://github.com/MustardSeedNetworks/stem/commit/5b03f3139d72c6a18b6dd8efe202221c9c07821f))
* **ci:** auto-trigger release-please on CI completion (was manual-only) ([5334db2](https://github.com/MustardSeedNetworks/stem/commit/5334db21fa76875e2a7ded4a24e14a8a52f31147))
* **ci:** build browser test server without cgo ([46d3a3b](https://github.com/MustardSeedNetworks/stem/commit/46d3a3ba31a1bdd77d1fbc434f42f6b9f4767242))
* **ci:** build stem native library with clang ([59f46a0](https://github.com/MustardSeedNetworks/stem/commit/59f46a0fa7d6bef2a24e6f5558b27fd03b2c15ca))
* **ci:** build stem native test dependencies ([dfb6d45](https://github.com/MustardSeedNetworks/stem/commit/dfb6d45d0128dfc2f31aa38347dd4fddeb0e2818))
* **ci:** bump Dockerfile go-build to golang:1.26-bookworm ([032a37e](https://github.com/MustardSeedNetworks/stem/commit/032a37e2d50e3d774469132756532ee783eaae38))
* **ci:** correct artifact path + Docker [@locales](https://github.com/locales) copy ([b4902e4](https://github.com/MustardSeedNetworks/stem/commit/b4902e4ac2ae194aa06925c48fab173c33f74804))
* **ci:** fetch full history for security scans ([655c135](https://github.com/MustardSeedNetworks/stem/commit/655c135c05b9d7c025cc1138bbd1f3826932acb9))
* **ci:** handle missing dataplane contexts ([8736134](https://github.com/MustardSeedNetworks/stem/commit/8736134b10b1a8a23a23d9b2007bad41ed7dac2f))
* **ci:** inject UIBuildHash ldflag (Universal Build Contract) ([#282](https://github.com/MustardSeedNetworks/stem/issues/282)) ([ca443be](https://github.com/MustardSeedNetworks/stem/commit/ca443be20f491f7dcb36785e4eb485248006c2f6))
* **ci:** keep stem analysis advisory ([74f779e](https://github.com/MustardSeedNetworks/stem/commit/74f779e0de00fa7bd4c2fef92f0bed0cce4347ac))
* **ci:** link native dataplane tests ([b6da226](https://github.com/MustardSeedNetworks/stem/commit/b6da22688638460abb5b2279024cfcf1b00793b8))
* **ci:** point Lighthouse at the real served URLs ([#65](https://github.com/MustardSeedNetworks/stem/issues/65)) ([#220](https://github.com/MustardSeedNetworks/stem/issues/220)) ([cde7653](https://github.com/MustardSeedNetworks/stem/commit/cde7653e76c771bcc8f497c0cba8cdd419f974ed))
* **ci:** race detector needs C dataplane deps + serialize SSE tests ([#199](https://github.com/MustardSeedNetworks/stem/issues/199)) ([34fad0d](https://github.com/MustardSeedNetworks/stem/commit/34fad0d5337e9b1dc03315599d39c7dd4087d483))
* **ci:** rename status import to statusColor to avoid noShadow lint ([da4d3d9](https://github.com/MustardSeedNetworks/stem/commit/da4d3d9de1535eb94d7c030e6352f5ce8c703c8d))
* **ci:** repair buildpacks project metadata ([cdcb63f](https://github.com/MustardSeedNetworks/stem/commit/cdcb63f4965cc080cae68daa7b9be0fd7d0033f0))
* **ci:** repair label sync workflow ([7acb464](https://github.com/MustardSeedNetworks/stem/commit/7acb4647a4eb80d138f01a10a5a3b113bebaae40))
* **ci:** report stem analyzer findings ([d726b50](https://github.com/MustardSeedNetworks/stem/commit/d726b501d973ee8fbf1bda2975d9ed13ff7feb48))
* **ci:** resolve stem workflow blockers ([314785d](https://github.com/MustardSeedNetworks/stem/commit/314785d6c3f3a0f763e3758b3ba64fffdddf50c5))
* **ci:** restore stem validation pipeline ([c1a26b2](https://github.com/MustardSeedNetworks/stem/commit/c1a26b20afce1f59e5a0b694d263d62860b1c41f))
* **ci:** run stub unit tests without race ([6272714](https://github.com/MustardSeedNetworks/stem/commit/62727147bada8993d1ce1682e64925c09aee02b6))
* **ci:** satisfy servicetest lint ([ec275df](https://github.com/MustardSeedNetworks/stem/commit/ec275df79aa63360ee069f492469d13c6633fc70))
* **ci:** scope stem container and license checks ([d267154](https://github.com/MustardSeedNetworks/stem/commit/d2671547ae280830d09777768d5635d58721dfd6))
* **ci:** scope stem e2e smoke suite ([4ce2153](https://github.com/MustardSeedNetworks/stem/commit/4ce2153966bff419ad4fb47f75edbd336db2c9a9))
* **ci:** skip stem docker publish without dockerfile ([a5a9deb](https://github.com/MustardSeedNetworks/stem/commit/a5a9debc00402ef97b0ce79d052038ee5fe5116b))
* **ci:** split native compile from unit tests ([f1f8c82](https://github.com/MustardSeedNetworks/stem/commit/f1f8c82c6be3026e969a8917cffc075841eafeba))
* **ci:** stabilize automated validation ([76209fa](https://github.com/MustardSeedNetworks/stem/commit/76209faef490df7baa09d161222ec7fc5da838e8))
* **ci:** stabilize stem browser smoke gate ([7dc7655](https://github.com/MustardSeedNetworks/stem/commit/7dc765542a92fcd6465aa1f483e19aadea440ab1))
* **ci:** start stem web server in browser jobs ([2c9f44b](https://github.com/MustardSeedNetworks/stem/commit/2c9f44b0c29dc60aaf97345a781a9748355defac))
* **ci:** suppress biome noBarrelFile on intentional theme barrel ([ee76bd3](https://github.com/MustardSeedNetworks/stem/commit/ee76bd3ac7de18181a02386e1d30f38f39078b38))
* **ci:** trigger CodeQL on PR + push + weekly schedule ([#293](https://github.com/MustardSeedNetworks/stem/issues/293)) ([736fed4](https://github.com/MustardSeedNetworks/stem/commit/736fed43badd1852109845ee1011f4b452f8f539))
* **ci:** unblock main — race in sse, lighthouse cert, e2e selectors ([#345](https://github.com/MustardSeedNetworks/stem/issues/345)) ([533f7f8](https://github.com/MustardSeedNetworks/stem/commit/533f7f8de044cbf88222ea006ee47057d664d092))
* **ci:** unblock stem main — PageHeader testid, E2E selector, robots.txt ([#347](https://github.com/MustardSeedNetworks/stem/issues/347)) ([6a5de95](https://github.com/MustardSeedNetworks/stem/commit/6a5de95ca0a18b1ea93be20e07cc09553b6b758f))
* **ci:** unescape apostrophe in target_tag description ([#229](https://github.com/MustardSeedNetworks/stem/issues/229)) ([e0c3d16](https://github.com/MustardSeedNetworks/stem/commit/e0c3d16120d2265e050a1e5c5c7cbc31be5bc5c0))
* **ci:** use compatible labeler action ([99c9c57](https://github.com/MustardSeedNetworks/stem/commit/99c9c57eab8ee28c0a69d6a1570046cd6b49c596))
* **ci:** use current intel macos release runner ([715b7bf](https://github.com/MustardSeedNetworks/stem/commit/715b7bf229e49ec2c98615f999065d63ff7e8613))
* **ci:** use go-version-file and fix c lint ([cf34ecb](https://github.com/MustardSeedNetworks/stem/commit/cf34ecb9bed85ac54440e0ce1729b4e7ec9c2b62))
* **ci:** use hosted node setup in container workflow ([9023b15](https://github.com/MustardSeedNetworks/stem/commit/9023b15e74f79c4d929145aaca4dd1067da8b718))
* **ci:** use labeler yaml format ([8d68517](https://github.com/MustardSeedNetworks/stem/commit/8d6851793528dd8862dd6c5bd9fde29866b485b2))
* **ci:** verify UIBuildHash embedded in built binary ([#286](https://github.com/MustardSeedNetworks/stem/issues/286)) ([b35a6f4](https://github.com/MustardSeedNetworks/stem/commit/b35a6f4d1383f31a41fe8510a8e64396a9cf310a))
* comprehensive code quality improvements ([4a25f5a](https://github.com/MustardSeedNetworks/stem/commit/4a25f5ae58a524193c38ecbe05cccacd0533f565))
* comprehensive lint, format, and test cleanup ([e26662a](https://github.com/MustardSeedNetworks/stem/commit/e26662aebd2813650911d2a7e2732d9eec0653b7))
* correct API paths and TLS config in smoke test ([25c7b10](https://github.com/MustardSeedNetworks/stem/commit/25c7b108da887532c720a7d35d88494fc2119456))
* **dataplane:** replace interface{} with typed ConfigUpdate struct ([8e2ee57](https://github.com/MustardSeedNetworks/stem/commit/8e2ee57f2de4712be42ebce4574f8dea29bfcc76)), closes [#2](https://github.com/MustardSeedNetworks/stem/issues/2)
* **dataplane:** require full payload length in RFC2544/Y.1564 frame validators ([#410](https://github.com/MustardSeedNetworks/stem/issues/410)) ([1a2239c](https://github.com/MustardSeedNetworks/stem/commit/1a2239c7dded5978bcb449e0cfca12f76679487e))
* **dataplane:** return error on invalid OUI in UpdateConfig ([#11](https://github.com/MustardSeedNetworks/stem/issues/11)) ([499dcc3](https://github.com/MustardSeedNetworks/stem/commit/499dcc307d832a1d95dff7a8f5056a00e945d92b))
* **deps:** bump golang.org/x/net to v0.55.0 (GO-2026-5026) ([855f165](https://github.com/MustardSeedNetworks/stem/commit/855f1659df1b4ade02bde6b1678de9705070db32))
* **deps:** Bump golang.org/x/net to v0.55.0 (GO-2026-5026) ([4011ac4](https://github.com/MustardSeedNetworks/stem/commit/4011ac41a5598ce1268636d508ac224305c0e52d))
* **docs:** correct PR template 'cd web' -&gt; 'cd ui' ([#283](https://github.com/MustardSeedNetworks/stem/issues/283)) ([79c2782](https://github.com/MustardSeedNetworks/stem/commit/79c27821344623197fe24006c34b4736f0f05379))
* **e2e:** role-chip-test_master uses underscore not hyphen ([#388](https://github.com/MustardSeedNetworks/stem/issues/388)) ([8a3e683](https://github.com/MustardSeedNetworks/stem/commit/8a3e6833b0c56a7dbd67f841373c061d5a85c5e9))
* handle --help correctly in test command and smoke test runner ([6f7ad30](https://github.com/MustardSeedNetworks/stem/commit/6f7ad3007301bc4b14573b50fbf94adaa3af91e7))
* harden API security and resolve all open issues ([2e671ea](https://github.com/MustardSeedNetworks/stem/commit/2e671ea68f1a8ba3b9fcd1c201feac50262175d7))
* **i18n-es:** normalize accents and add missing diacritics (107 fixes) ([#321](https://github.com/MustardSeedNetworks/stem/issues/321)) ([c4e5870](https://github.com/MustardSeedNetworks/stem/commit/c4e58708a596b760447656261849a23e7b494abd))
* **i18n:** resolve 48 t() calls referencing missing EN locale keys ([#329](https://github.com/MustardSeedNetworks/stem/issues/329)) ([b0232aa](https://github.com/MustardSeedNetworks/stem/commit/b0232aaeee5362edd79c0dd3c42165a366b186bf))
* **i18n:** update document.lang on locale change for a11y ([#316](https://github.com/MustardSeedNetworks/stem/issues/316)) ([963bf12](https://github.com/MustardSeedNetworks/stem/commit/963bf128daa57e09a00d2da4414604922a5d5ce2))
* **license:** add RWMutex to Manager for safe concurrent access ([#312](https://github.com/MustardSeedNetworks/stem/issues/312)) ([cf8afe3](https://github.com/MustardSeedNetworks/stem/commit/cf8afe3c1559833a57bcba858431517b824e560a))
* **lint:** Clear gocognit, godoclint, nestif, tparallel ([#262](https://github.com/MustardSeedNetworks/stem/issues/262)) ([19a2234](https://github.com/MustardSeedNetworks/stem/commit/19a223489207abdcb3326a09190c9ef749301d7b))
* **lint:** Extract test-type + standard-name consts (partial goconst cleanup) ([#263](https://github.com/MustardSeedNetworks/stem/issues/263)) ([9bcb8aa](https://github.com/MustardSeedNetworks/stem/commit/9bcb8aac7feb4b53b19b1cabe673781db5b90698))
* **lint:** resolve all remaining lint issues ([fe4bac2](https://github.com/MustardSeedNetworks/stem/commit/fe4bac2f2b05785fcbbbc641e05513c93f3e89ba))
* **metrics:** serialize tests that share Prometheus counter labels ([3e413bc](https://github.com/MustardSeedNetworks/stem/commit/3e413bc196564221a31f5a4ced920cc446623e15))
* **modules:** resolve all golangci-lint warnings ([2fd1bad](https://github.com/MustardSeedNetworks/stem/commit/2fd1badd462b1ae5a997b925851e2a52c811ed18))
* prevent smoke test early exit on assertion failures ([7408218](https://github.com/MustardSeedNetworks/stem/commit/740821842f65a781c93667a5b92ca27ef09a14f1))
* **quality-gates:** honest exit codes, fmt-check, npm audit fix, CI/make parity ([#468](https://github.com/MustardSeedNetworks/stem/issues/468)) ([0993c8d](https://github.com/MustardSeedNetworks/stem/commit/0993c8d4246f07645e5e29de1e93da32154318b2))
* **reflector/web:** use writeJSON and secure CORS headers ([f121557](https://github.com/MustardSeedNetworks/stem/commit/f121557c5406fa4349bc9ef3fb5b9572e7fb977d))
* regenerate lockfiles to resolve esbuild vulnerability ([dbbcd7e](https://github.com/MustardSeedNetworks/stem/commit/dbbcd7e1d5cce153bf72c7a4a3e64cd0cfcc5e13))
* **release:** align stem release manifest with latest 0.x tag ([#500](https://github.com/MustardSeedNetworks/stem/issues/500)) ([7741954](https://github.com/MustardSeedNetworks/stem/commit/77419546ed29d6ca62e954676094ba8a74b3b46e))
* **release:** Replace broken SLSA generator with attest-build-provenance ([#208](https://github.com/MustardSeedNetworks/stem/issues/208)) ([4af33d0](https://github.com/MustardSeedNetworks/stem/commit/4af33d0d4b56bcb02da8cdcd9babce8b09550088))
* **release:** revert invalid 1.0.0 changelog ([#498](https://github.com/MustardSeedNetworks/stem/issues/498)) ([bc2d53b](https://github.com/MustardSeedNetworks/stem/commit/bc2d53b356cc74d27297cd85037be1b60049bbbf))
* resolve 15 frontend-backend integration issues ([1560a2f](https://github.com/MustardSeedNetworks/stem/commit/1560a2f54369b65b7bb063151b3341c856896295))
* resolve all lint warnings and errors ([134b8ab](https://github.com/MustardSeedNetworks/stem/commit/134b8abffbe04324bf02779ee18304dd4de16a54))
* resolve concurrent request hang in smoke test ([22b9083](https://github.com/MustardSeedNetworks/stem/commit/22b9083c44ea27a95c49751f5edb46e9ea8f5851))
* resolve errcheck and gocritic lint warnings ([7eefef9](https://github.com/MustardSeedNetworks/stem/commit/7eefef9c11b899147d427dda2d49771f9df89b86))
* resolve golangci-lint warnings and wire remaining module executors ([a997f25](https://github.com/MustardSeedNetworks/stem/commit/a997f2507d7d3ef56013c9d9e813ac5d0f7a2bd5))
* **rpm:** add user/group Provides for Fedora compatibility ([a539c93](https://github.com/MustardSeedNetworks/stem/commit/a539c93f32b401441c7fb7cb723bcb726a187a76))
* **rpm:** auto-enable and start stem service on install ([a662684](https://github.com/MustardSeedNetworks/stem/commit/a662684d2f3e03d7cf77018e690baa3402f51f87))
* **scripts:** clean up all shellcheck warnings + pin severity=warning ([#307](https://github.com/MustardSeedNetworks/stem/issues/307)) ([43c2d62](https://github.com/MustardSeedNetworks/stem/commit/43c2d62f64ffaad5636343adfa51568f53105a8b))
* **scripts:** deploy-validate default scheme=https + port=8444 ([#292](https://github.com/MustardSeedNetworks/stem/issues/292)) ([acdf2a6](https://github.com/MustardSeedNetworks/stem/commit/acdf2a684758a25810a972e11582a55f388fd797))
* **security:** add TLS, httpOnly cookies, and security headers ([04082fb](https://github.com/MustardSeedNetworks/stem/commit/04082fbd8743ad4872e141250a84e5949896d090))
* **security:** auth-gate /mode /settings /interfaces /stats + first-run hardening ([#340](https://github.com/MustardSeedNetworks/stem/issues/340), [#356](https://github.com/MustardSeedNetworks/stem/issues/356)) ([#357](https://github.com/MustardSeedNetworks/stem/issues/357)) ([5cd6489](https://github.com/MustardSeedNetworks/stem/commit/5cd64895992b02166f7566e44047bc6020a90cfa))
* **security:** bump Go 1.26.5 and gate Trivy SARIF ([#496](https://github.com/MustardSeedNetworks/stem/issues/496)) ([a7aa351](https://github.com/MustardSeedNetworks/stem/commit/a7aa351157e59a1dec259cdb00ee4a4bc0de06d6))
* **security:** complete Phase 1 critical security fixes ([9f42f6c](https://github.com/MustardSeedNetworks/stem/commit/9f42f6cd2dbf86466c205ec96f097cb5d3ce0e06))
* **security:** Hardcode HTTPS-only auth, cipher overflow safety, fixture renames ([#1070](https://github.com/MustardSeedNetworks/stem/issues/1070)) ([#260](https://github.com/MustardSeedNetworks/stem/issues/260)) ([de2d166](https://github.com/MustardSeedNetworks/stem/commit/de2d16677f25ac2a5fc1f9200a7f00703fcbec13))
* **security:** prevent rate limiter memory exhaustion DoS ([#112](https://github.com/MustardSeedNetworks/stem/issues/112)) ([d32c8be](https://github.com/MustardSeedNetworks/stem/commit/d32c8be217001a25af347d4d43c9c25bc49d1bd0))
* **security:** resolve gosec issues and improve auth test coverage ([2b342d3](https://github.com/MustardSeedNetworks/stem/commit/2b342d3f5ebd1117493fac14f38500ceb43e17d9))
* **security:** scope generated TLS certificate writes ([83f6cef](https://github.com/MustardSeedNetworks/stem/commit/83f6cef51e216a8c2a9b7c6e713fc064541de697))
* **server:** clean lint issues ([652af1c](https://github.com/MustardSeedNetworks/stem/commit/652af1c8df5c86e1dcd0b1d7320cb453a8e73e2e))
* set vite outDir to internal/api/dist for Go embed ([500f676](https://github.com/MustardSeedNetworks/stem/commit/500f6765ecd8001738b917234e173a0fc40fb389))
* smoke test handles missing C dataplane gracefully ([6f34d6f](https://github.com/MustardSeedNetworks/stem/commit/6f34d6f261839b56f2cb8ec65af46b7e9a8220e5))
* **smoke:** annotate STEM_AUTH_PASSWORD assignment as gitleaks:allow ([#389](https://github.com/MustardSeedNetworks/stem/issues/389)) ([579774e](https://github.com/MustardSeedNetworks/stem/commit/579774ebc81bb196f1da381f9ba7904831d7ccb8))
* **systemd:** remove invalid --config flag from service file ([5e0f2c0](https://github.com/MustardSeedNetworks/stem/commit/5e0f2c04404101188f104a376d6eea4372b778f2))
* **tests:** add captureStdout helper for error-checked output capture ([d139a30](https://github.com/MustardSeedNetworks/stem/commit/d139a308d930ee118c4bf9be9f02e40ed08b2497))
* **tests:** fix errcheck warnings in test files ([#7](https://github.com/MustardSeedNetworks/stem/issues/7)) ([3161abd](https://github.com/MustardSeedNetworks/stem/commit/3161abd9d2d99f3f9875ba7ffb531aebc2d7b639))
* **tests:** gate remaining measure tests under -short ([#201](https://github.com/MustardSeedNetworks/stem/issues/201)) ([b0fc1be](https://github.com/MustardSeedNetworks/stem/commit/b0fc1be9382e540c9ae252445de392db22e7a696))
* **tests:** improve test coverage and fix all lint issues ([29652b9](https://github.com/MustardSeedNetworks/stem/commit/29652b9148526d2736e67f3fa0082e961234c473))
* **tests:** make race detector pass on Linux + CGO ([#200](https://github.com/MustardSeedNetworks/stem/issues/200)) ([23cb945](https://github.com/MustardSeedNetworks/stem/commit/23cb9458dd5328361591743b2ccb1de468308597))
* **tests:** resolve all exhaustruct lint issues in test files ([c6a30c8](https://github.com/MustardSeedNetworks/stem/commit/c6a30c8c881538f9525669e58c39c050bb976182))
* **tests:** resolve goroutine leaks and auth test failures ([c7c56e5](https://github.com/MustardSeedNetworks/stem/commit/c7c56e5fdd0c069ce483bbbd2a9b055e682d27a7))
* **tests:** use SetOutput for proper stdout capture in help tests ([53cd204](https://github.com/MustardSeedNetworks/stem/commit/53cd20465f3085363c5fb249e5593cd3e2b00765))
* **ui,api:** replace hardcoded "0.1.0" with /__version + add the endpoint ([#212](https://github.com/MustardSeedNetworks/stem/issues/212)) ([69fe359](https://github.com/MustardSeedNetworks/stem/commit/69fe359dbaffcaf7f8a5fd73bd62a175ed9c0948))
* **ui:** add explicit type annotations in HelpDrawer ([c263cff](https://github.com/MustardSeedNetworks/stem/commit/c263cff2dd4f58f39d72d7dd3bf5a175c373b06a))
* **ui:** add logging for silently ignored errors ([#113](https://github.com/MustardSeedNetworks/stem/issues/113)) ([0b8b110](https://github.com/MustardSeedNetworks/stem/commit/0b8b1109903db62b0cddf9ad6837434a8d5d2f19))
* **ui:** add role=alert to stem login error display ([#385](https://github.com/MustardSeedNetworks/stem/issues/385)) ([d4a76a7](https://github.com/MustardSeedNetworks/stem/commit/d4a76a7b8124ebc0cc2fe3ef68a2c4acc3a483cb))
* **ui:** close token-gate escapes in themeComponents + narrow the gate (S5) ([#425](https://github.com/MustardSeedNetworks/stem/issues/425)) ([483b7b6](https://github.com/MustardSeedNetworks/stem/commit/483b7b6f72e535a04e33cea8e01faa4468b896e0))
* **ui:** complete httpOnly cookie auth migration ([1c987f0](https://github.com/MustardSeedNetworks/stem/commit/1c987f0b8cbe46546e27915a8ee4b79b41d8ad82))
* **ui:** correct import casing and add explicit type annotations ([4e59556](https://github.com/MustardSeedNetworks/stem/commit/4e595568261d9a0453311b272d0cff17bafa4650))
* **ui:** correct import path casing for RFC/TSN components ([96d056c](https://github.com/MustardSeedNetworks/stem/commit/96d056c78030f84dcb35d8fcfbd8c2c68002cc82))
* **ui:** drop sidebar-*-button testids on the mobile aside copy ([#381](https://github.com/MustardSeedNetworks/stem/issues/381)) ([b91ce7e](https://github.com/MustardSeedNetworks/stem/commit/b91ce7e9d425ab964bfcb57083c6c3ee88e69d9b))
* **ui:** enable erasableSyntaxOnly + refactor ApiError TS-only syntax ([#290](https://github.com/MustardSeedNetworks/stem/issues/290)) ([1cc7d6a](https://github.com/MustardSeedNetworks/stem/commit/1cc7d6a5f3b19437659ee3d7c7aa2bbdcf1842e9)), closes [#285](https://github.com/MustardSeedNetworks/stem/issues/285)
* **ui:** give module-certify a distinct violet hue (S7) ([#430](https://github.com/MustardSeedNetworks/stem/issues/430)) ([5ab0891](https://github.com/MustardSeedNetworks/stem/commit/5ab089133cd7aa6e110f1b47d0cc50f156e3173a))
* **ui:** i18n the Account nav group label (S3) ([#419](https://github.com/MustardSeedNetworks/stem/issues/419)) ([4922212](https://github.com/MustardSeedNetworks/stem/commit/49222125166f66f6404a81e8a47e31c137211429))
* **ui:** render sidebar product name from t('app.title'), not hardcoded ([#350](https://github.com/MustardSeedNetworks/stem/issues/350)) ([446b639](https://github.com/MustardSeedNetworks/stem/commit/446b6390f4353378d0b8b55c2d6721a39d1807b2))
* **ui:** repair token-discipline guard and close white/black color leaks ([#361](https://github.com/MustardSeedNetworks/stem/issues/361)) ([e33fd7f](https://github.com/MustardSeedNetworks/stem/commit/e33fd7fe0df7067a0f35a0349a76c49ec870c884))
* **ui:** replace undefined status-danger token with status-error ([#366](https://github.com/MustardSeedNetworks/stem/issues/366)) ([69e1e30](https://github.com/MustardSeedNetworks/stem/commit/69e1e30b78244226307d42d2ee46870190de9b18))
* **ui:** stop emitted vite.config.js shadowing vite.config.ts ([#416](https://github.com/MustardSeedNetworks/stem/issues/416)) ([65421cd](https://github.com/MustardSeedNetworks/stem/commit/65421cd7697de8ebfd4a1eb7f8317e0aae6e30a2))
* **ui:** suppress node dep0205 build warning ([2ffcaf8](https://github.com/MustardSeedNetworks/stem/commit/2ffcaf88819ec55016133e03e76255d9aff74484))
* update Logf function callers in main.go ([1fc6c9a](https://github.com/MustardSeedNetworks/stem/commit/1fc6c9a002e333ee8387174977a8a30108324a1e))
* update test types to use module-prefixed names ([3b6d7a7](https://github.com/MustardSeedNetworks/stem/commit/3b6d7a7eca89d6f00cc1de491b15e411c50f7c0f))
* **vite:** stop inlining font assets as data: URLs (CSP fix) ([2f3099f](https://github.com/MustardSeedNetworks/stem/commit/2f3099fef8ed508bfc1fe1651a31aafa639d90c4))
* **vite:** Stop inlining font assets as data: URLs (CSP fix) ([96b4b8a](https://github.com/MustardSeedNetworks/stem/commit/96b4b8a812dcaacb79907df73cc017755949e0c2))
* **web:** add interface validation to handleSettings ([#4](https://github.com/MustardSeedNetworks/stem/issues/4)) ([71578cc](https://github.com/MustardSeedNetworks/stem/commit/71578ccf927acb7958745900e6eb8729f508c1a9))
* **web:** add observability for config update operations ([#5](https://github.com/MustardSeedNetworks/stem/issues/5)) ([a9eea18](https://github.com/MustardSeedNetworks/stem/commit/a9eea186447f2acef2f585d428790a018f25cab9))
* **web:** add writeJSON error handling and HTTP server timeouts ([c6a65e5](https://github.com/MustardSeedNetworks/stem/commit/c6a65e5e70a55ca7ff9cb64e6907bd4f0fcf82a9))
* **web:** implement actual test execution via module executors ([#9](https://github.com/MustardSeedNetworks/stem/issues/9)) ([b23eb48](https://github.com/MustardSeedNetworks/stem/commit/b23eb48b877b6b8975b12e7391ed2e65b9a3eecf))
* wire reflector API config/stats to dataplane and add stop support ([fc95ea0](https://github.com/MustardSeedNetworks/stem/commit/fc95ea09cc7af43a3f62f3024fe2c57bbbbe1ca3))


### Performance Improvements

* **e2e:** bump CI workers 1-&gt;2 and retries 2-&gt;1 ([#255](https://github.com/MustardSeedNetworks/stem/issues/255)) ([6b8c658](https://github.com/MustardSeedNetworks/stem/commit/6b8c65891f95b62c0a6b9200b22c3dc61739d5ef))
* **tests:** optimize test performance with caching and fast bcrypt ([e4364e7](https://github.com/MustardSeedNetworks/stem/commit/e4364e71ab0d8cf43042680516903ec0fb7e3582))
* **ui:** split vendor chunks, add bundle analyzer ([#417](https://github.com/MustardSeedNetworks/stem/issues/417)) ([cd30725](https://github.com/MustardSeedNetworks/stem/commit/cd3072579de14cfbf5a35fbc2076ad8ddcb2e82b))

## [0.23.1](https://github.com/krisarmstrong/stem/compare/v0.23.0...v0.23.1) (2026-05-30)


### Bug Fixes

* **e2e:** role-chip-test_master uses underscore not hyphen ([#388](https://github.com/krisarmstrong/stem/issues/388)) ([8a3e683](https://github.com/krisarmstrong/stem/commit/8a3e6833b0c56a7dbd67f841373c061d5a85c5e9))
* **smoke:** annotate STEM_AUTH_PASSWORD assignment as gitleaks:allow ([#389](https://github.com/krisarmstrong/stem/issues/389)) ([579774e](https://github.com/krisarmstrong/stem/commit/579774ebc81bb196f1da381f9ba7904831d7ccb8))
* **ui:** add role=alert to stem login error display ([#385](https://github.com/krisarmstrong/stem/issues/385)) ([d4a76a7](https://github.com/krisarmstrong/stem/commit/d4a76a7b8124ebc0cc2fe3ef68a2c4acc3a483cb))
* **ui:** drop sidebar-*-button testids on the mobile aside copy ([#381](https://github.com/krisarmstrong/stem/issues/381)) ([b91ce7e](https://github.com/krisarmstrong/stem/commit/b91ce7e9d425ab964bfcb57083c6c3ee88e69d9b))

## [0.23.0](https://github.com/krisarmstrong/stem/compare/v0.22.0...v0.23.0) (2026-05-29)


### Features

* **a11y:** axe-core test harness for the Button primitives ([#370](https://github.com/krisarmstrong/stem/issues/370)) ([991c92f](https://github.com/krisarmstrong/stem/commit/991c92f9fbf572a6d2278e1315a3995f0a1cb29b))
* **cli:** help entries for tui/list-tests/install-ca + completeness test ([#371](https://github.com/krisarmstrong/stem/issues/371)) ([85d56a1](https://github.com/krisarmstrong/stem/commit/85d56a1b00ff71e5693906eecca4bfb866414b52))
* **help:** surface version in HelpDrawer header (in-app About) ([#372](https://github.com/krisarmstrong/stem/issues/372)) ([86b76b5](https://github.com/krisarmstrong/stem/commit/86b76b5669d022b8dfe95a087d19421f727312f5))

## [0.22.0](https://github.com/krisarmstrong/stem/compare/v0.21.4...v0.22.0) (2026-05-29)


### Features

* **i18n:** en/es key parity + DNT compliance test (stem) ([#373](https://github.com/krisarmstrong/stem/issues/373)) ([f36d954](https://github.com/krisarmstrong/stem/commit/f36d95410515398467e17f9d72c23bb5b2bde15b))

## [0.21.4](https://github.com/krisarmstrong/stem/compare/v0.21.3...v0.21.4) (2026-05-29)


### Bug Fixes

* **ui:** replace undefined status-danger token with status-error ([#366](https://github.com/krisarmstrong/stem/issues/366)) ([69e1e30](https://github.com/krisarmstrong/stem/commit/69e1e30b78244226307d42d2ee46870190de9b18))

## [0.21.3](https://github.com/krisarmstrong/stem/compare/v0.21.2...v0.21.3) (2026-05-29)


### Bug Fixes

* **ui:** repair token-discipline guard and close white/black color leaks ([#361](https://github.com/krisarmstrong/stem/issues/361)) ([e33fd7f](https://github.com/krisarmstrong/stem/commit/e33fd7fe0df7067a0f35a0349a76c49ec870c884))

## [0.21.2](https://github.com/krisarmstrong/stem/compare/v0.21.1...v0.21.2) (2026-05-29)


### Bug Fixes

* **security:** auth-gate /mode /settings /interfaces /stats + first-run hardening ([#340](https://github.com/krisarmstrong/stem/issues/340), [#356](https://github.com/krisarmstrong/stem/issues/356)) ([#357](https://github.com/krisarmstrong/stem/issues/357)) ([5cd6489](https://github.com/krisarmstrong/stem/commit/5cd64895992b02166f7566e44047bc6020a90cfa))

## [0.21.1](https://github.com/krisarmstrong/stem/compare/v0.21.0...v0.21.1) (2026-05-28)


### Bug Fixes

* **ui:** render sidebar product name from t('app.title'), not hardcoded ([#350](https://github.com/krisarmstrong/stem/issues/350)) ([446b639](https://github.com/krisarmstrong/stem/commit/446b6390f4353378d0b8b55c2d6721a39d1807b2))

## [0.21.0](https://github.com/krisarmstrong/stem/compare/v0.20.0...v0.21.0) (2026-05-28)


### Features

* **ui:** canonical shell — modernize Sidebar + PageHeader (Phase 1) ([#339](https://github.com/krisarmstrong/stem/issues/339)) ([3983689](https://github.com/krisarmstrong/stem/commit/39836892a591fd847c615df8616a181fb22ddd6c))


### Bug Fixes

* **ci:** unblock main — race in sse, lighthouse cert, e2e selectors ([#345](https://github.com/krisarmstrong/stem/issues/345)) ([533f7f8](https://github.com/krisarmstrong/stem/commit/533f7f8de044cbf88222ea006ee47057d664d092))
* **ci:** unblock stem main — PageHeader testid, E2E selector, robots.txt ([#347](https://github.com/krisarmstrong/stem/issues/347)) ([6a5de95](https://github.com/krisarmstrong/stem/commit/6a5de95ca0a18b1ea93be20e07cc09553b6b758f))

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
