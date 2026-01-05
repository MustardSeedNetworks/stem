# Documentation reference

All canonical documentation for The Stem now lives under the [MustardSeedNetworks](../MustardSeedNetworks) workspace. The files that used to live in this repository are now maintained there; make your changes in MustardSeedNetworks and keep this folder as a lightweight pointer so we never drift between copies.

| Former `stem/docs/` file | Canonical location in MustardSeedNetworks |
|--------------------------|------------------------------------------|
| `API_REFERENCE.md`       | `05-Engineering/API_REFERENCE.md` (general API reference) and `03-The-Stem/THE_STEM_API_REFERENCE.md` (product-specific notes) |
| `MODULE_STATUS.md`       | `03-The-Stem/THE_STEM_IMPLEMENTATION_SPEC.md` covers current module/test status; `03-The-Stem/THE_STEM_ARCHITECTURE_DIAGRAMS.md` captures the system block view |
| `IMPLEMENTATION_PLAN.md` | `03-The-Stem/THE_STEM_IMPLEMENTATION_SPEC.md` is the plan of record for ongoing implementation work |
| `AI_README.md`           | `02-The-Seed/AI_README.md` (Seed-specific AI guidance) |
| `AI_MCP_SURVEY_INTEGRATION_PLAN.md` | `02-The-Seed/AI_MCP_SURVEY_INTEGRATION_PLAN.md` |
| `DEFECT_REMEDIATION_PLAN.md` | `05-Engineering/DEFECT_REMEDIATION_PLAN.md` |
| `DEVELOPMENT.md`         | `05-Engineering/DEVELOPMENT.md` (seed dev onboarding) |
| `DOCUMENTATION_STRUCTURE.md` | `05-Engineering/DOCUMENTATION_STRUCTURE.md` |
| `HARDWARE.md`            | `05-Engineering/HARDWARE_COMPATIBILITY.md` and the tiered views under `05-Engineering/HARDWARE_TIERS_SEED.md`/`_STEM.md` |
| `REFACTOR_PLAN.md`       | `05-Engineering/REFACTOR_PLAN.md` |
| `STYLE_GUIDE.md`         | `05-Engineering/STYLE_GUIDE.md` (coding style and prose conventions) |
| `TESTING.md`             | `05-Engineering/TESTING.md` plus `05-Engineering/TESTING_STRATEGY.md`/`03-The-Seed/THE_SEED_TESTING_STRATEGY.md` |
| `SURVEY_COMPLETION_PLAN.md` | `05-Engineering/SURVEY_COMPLETION_PLAN.md` |
| `WIKI_CONTENT.md`        | `05-Engineering/WIKI_CONTENT.md` |

If new documentation is needed for The Stem, add it under MustardSeedNetworks (use the `03-The-Stem` and `05-Engineering` folders as appropriate) and then update this table to point to the new file so readers here know where to look.

## Additional moved content

- Seed-specific AI planning, tooling, and QA notes now live in `02-The-Seed/` (files prefixed with `AI_*`).  
- Internal product planning (brand, marketing, sales, support, testing, hardware plans, etc.) can be found in the respective directories under MustardSeedNetworks (`01-Strategy`, `04-Brand-Marketing`, `05-Engineering`, `06-Sales`, `07-Support-TAC`, `09-Marketing`).  
- Reference material such as API types, CI tooling analysis, and templates now live under `05-Engineering/reference/` and `05-Engineering/templates/`, while the technical wiki has been copied into `05-Engineering/wiki/`.  
- NetAlly AirMapper capture files moved into `02-The-Seed/NetAllyAirMapper/` for safekeeping.  

Always update the canonical workspace first; this repository should not reintroduce conflicting copies.
