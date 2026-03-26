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
- Use `.ai/AGENTS.md` for stable instructions.
- Use `.ai/backlog.md` as shared source for queued and active work.
- Keep one primary in-progress task to reduce collisions.

## Known Risks / Watchouts
- Existing `todo.md` is broad and mixed; avoid rewriting it automatically.
- Keep AI tracking files lightweight to prevent maintenance overhead.

## Update Log
- 2026-03-14: Initialized AI orchestration notes file.
- 2026-03-14: Converted `todo.md` section `Interactables and puzzles` into `.ai/backlog.md` items AI-003 through AI-013.
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
- 2026-03-14: Completed AI-008 by adding filtered transmit reactions `transmitPushByState` and `transmitPushByType` (argument-driven rule actions) while keeping direction-safe processing to avoid duplicate pushes; updated SPA reaction registry entries and added focused server tests for rule resolution + selective movement.
- 2026-03-14: Follow-up naming cleanup for AI-008: renamed filtered name-target reaction consistently to `transmitPushByName` across server registry/runtime, tests, and SPA reaction registry labels/args.
- 2026-03-14: Completed AI-020 by fixing SPA interactable state editing UX: added explicit state rename input/button flow and fixed cross-state reaction rule mutation by deep-cloning state configs/rule arrays when adding states; verified via focused editor component tests.
- 2026-03-14: Completed AI-021 by fixing compile-time interactable state precedence in `tools/main/context.go` so blank tile state resolves to `defaultState` (instead of persisted base `state`), with new regression tests in `tools/main/context_test.go` covering default and explicit tile-state behavior.
- 2026-03-14: Completed AI-010 by adding `sticky` bool field across all layers: server Go structs (Interactable, InteractableState, InteractableDescription, InteractableStateDescription), stage creation wiring, tile walkability; tools Go structs and compile pipeline; SPA TypeScript interfaces, editor component, and template (checkbox + preview).
- 2026-03-14: Completed AI-011 by implementing `pushStickyGroup()` in character.go — BFS-based polyomino group discovery using TryLock for concurrency safety, with atomic multi-tile movement. Integrated sticky checks into Player.push()/NonPlayer.push(). Fixed TryLock contention bug where locked neighbors (from push chains) were causing false group-push failures by switching BFS to skip-on-failure instead of bail-on-failure. Added 10 focused tests: pairs, L-shapes, squares, walls, edges, occupied tiles, ball-into-group, non-sticky isolation, and state-switch scenarios.
- 2026-03-14: Completed AI-012 with design decision: fragile blocks in sticky groups do NOT break on failed group push; they only break from damage reactions. All group members must be able to move for push to succeed. Documented in pushStickyGroup code comments.
- 2026-03-14: Completed AI-013 by creating puzzle content: added `sticky-teal` and `sticky-purple` interactable definitions to `tools/main/data/collections/escape/interactables/puzzles.json`; created new `sticky.json` space in `tools/main/data/collections/escape/spaces/` with 3 rooms — sticky-pair (basic pair movement), sticky-l-group (L-shaped group with interior walls), sticky-transmit (transmitPushAll + sticky pair with corridor barriers).
- 2026-03-26: Completed Cooler Visuals / Dynamic weather. Added tokenized dynamic weather mode support in `server/main/assets/canvas.js` with first implemented mode `raining`.
	- Design decisions:
		- Dynamic weather is client-rendered on the existing `Lw1` weather canvas layer; no server schema/message changes were required.
		- Weather mode selection is token-based from existing weather class strings (example: `blue trsp20 raining`), preserving compatibility with color/transparency classes.
		- Unknown weather tokens are ignored by the renderer so content can safely carry future mode tags before client support exists.
	- Current limitations:
		- Mode detection reads only currently visible `Lw1` tiles; mixed per-tile weather modes in a single camera view are not blended (first detected mode wins).
		- Raining uses a single global drop style and fixed animation profile; there is no intensity token yet.
		- Weather animation redraws only `Lw1` each frame for correctness, which is acceptable at current 16x16 view size but should be monitored if view area grows.
	- Future expansion possibilities:
		- Add tokens like `rain-light`, `rain-heavy`, `wind-left`, `wind-right` to parameterize density/speed/drift.
		- Add additional mode renderers (`snowing`, `foggy`, `ash`, `sandstorm`) in the same token-renderer map.
		- Extend mode resolution to support deterministic priority or blending when multiple mode tokens are present.
