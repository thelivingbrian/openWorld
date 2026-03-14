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
