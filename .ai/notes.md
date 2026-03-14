# AI Notes (Mutable)

Working memory for AI agents. This file is intentionally editable by agents.

## Session Summary
- Date: 2026-03-14
- Branch: `ai-orchestration`
- Focus: establish AI collaboration scaffolding

## Repo Quick Map
- `server/main`: gameplay server runtime
- `tools/main`: authoring/deployment tooling
- `integration/main`: integration runner
- `deploy/droplet`: deployment scripts/service

## Current Decisions
- Use `AGENTS.md` for stable instructions.
- Use `AI_BACKLOG.md` as shared source for queued and active work.
- Keep one primary in-progress task to reduce collisions.

## Known Risks / Watchouts
- Existing `todo.md` is broad and mixed; avoid rewriting it automatically.
- Keep AI tracking files lightweight to prevent maintenance overhead.

## Update Log
- 2026-03-14: Initialized AI orchestration notes file.
- 2026-03-14: Converted `todo.md` section `Interactables and puzzles` into `AI_BACKLOG.md` items AI-003 through AI-013.
- 2026-03-14: Completed AI-003 by adding interactable `state` to server runtime + JSON description load path and added focused test coverage.
- 2026-03-14: Completed AI-004 by adding tools+SPA support to set/edit interactable `state` and persist it through interactable set save/load.
- 2026-03-14: Completed AI-014 by making interactable states full configuration sets, adding per-tile state selection in grid place/view, and resolving selected/default state during compile + server stage creation.
- 2026-03-14: Completed AI-005 by adding `reactsWith` gates `interactableStateIs`, `interactableStateIsNot`, and `interactableStateContains` in server runtime + rule registries, with focused tests and editor registry entries.
- 2026-03-14: Completed AI-006 by adding `transmitPushAll` reaction support via offset-aware reaction wiring, exposing it in the SPA reaction registry, and adding focused server tests for registry resolution and push transmission behavior.
- 2026-03-14: Completed AI-015 by fixing Interactables Add Set UX in the SPA editor: moved New Set + Add Set controls to the right of Save Set/Add Interactable, added explicit blank/duplicate set-name status messages, and added focused component tests for add-set success and validation paths.
- 2026-03-14: Completed AI-016 by fixing a stale SPA test expectation in `grid-engine.spec.ts` (`normalizeTile` now includes `interactableState`); verified with targeted and full Jest runs (all suites passing).
- 2026-03-14: Completed AI-017 by fixing SPA interactable set/edit refresh regressions: `touchBootstrap` now clones current collection maps to trigger signal recomputation after nested mutations, and `addInteractable` now re-syncs edited state selection so new interactables remain editable; validated with focused Jest component tests.
- 2026-03-14: Completed AI-018 by fixing first-add visibility in brand-new interactable sets: `addInteractable` now assigns a new array (`[...]`) instead of mutating with `push`, ensuring immediate template updates without switching sets; added focused component regression test.
- 2026-03-14: Completed AI-007 by fixing `transmitPushAll` over-push behavior (right/down): it now snapshots and direction-sorts source tiles before pushing so moved interactables are not processed twice in one sweep; added focused server tests for right/down single-step transmission.
- 2026-03-14: Completed AI-019 by adding concurrency verification for `transmitPushAll` via a timeout-based simultaneous multi-direction push test; targeted `go test -run TransmitPushAll` passes. Race-detector execution is blocked in this environment because `go test -race` needs `gcc` (CGO C compiler) on PATH.
