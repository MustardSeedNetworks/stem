<!--
CHECKPOINT RULES (from session-management.md):
- Quick update: After any todo completion
- Full checkpoint: After ~20 tool calls or decisions
- Archive: End of session or major feature complete

After each task, ask: Decision made? >10 tool calls? Feature done?
-->

# Current Session State

*Last updated: 2025-01-18*

## Active Task
No active task

## Current Status
- **Phase**: idle
- **Progress**: N/A
- **Blocking Issues**: None

## Context Summary
The Stem is a network performance testing tool with Go backend and React/TypeScript frontend.
Module-based architecture for different testing standards (RFC 2544, Y.1564, Y.1731, etc.).

## Files Being Modified
| File | Status | Notes |
|------|--------|-------|
| N/A | - | - |

## Next Steps
1. [ ] Resume development work

## Key Context to Preserve
- Go 1.25.5 backend, Node 25.2.1 frontend
- Uses Biome for linting (not ESLint/Prettier)
- Module structure: Benchmark/ServiceTest/TrafficGen/Measure/Certify
- Port 8443 for HTTPS, 8080 for dev
- C23 dataplane for packet processing

## Resume Instructions
To continue this work:
1. Check _project_specs/todos/active.md for pending tasks
2. Review CLAUDE.md for project conventions
