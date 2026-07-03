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
- 2026-07-02: Added deployment-hardening follow-ups to the backlog: synchronize the installed deploy script, support configuration-only deployments, fail on restart errors, and health-check configuration changes.
- 2026-07-02: Fixed deployment service detection under `set -o pipefail`; the deploy script now uses `systemctl cat` instead of a `systemctl list-unit-files | grep -q` pipeline that could misinterpret SIGPIPE as a missing service and skip restarts.
- 2026-07-02: Implemented admin live preview by player and stage.
	- Added authenticated watch page/WebSocket routes and connected the player/stage modal watch links.
	- Player preview taps source updates before player socket batching, preserving independent canvas/HUD/menu filtering; highlights are included in initial and live canvas state.
	- Stage preview owns a non-blocking camera registered with existing camera zones and supports random initial view, 30-second visible-inactivity rotation, and manual Next camera.
	- Player logout swaps in a persistent dark status state. Bounded queues use full-snapshot recovery on overflow.
	- Added focused tests for filtering, mode-specific controls, highlights, logout state, backpressure, and visible stage activity.
- 2026-07-02: Recorded live-preview behavior decisions.
	- Player preview includes highlights; HUD and menu visibility are independent settings, initially off.
	- Player logout retains the preview in a dark logged-out state rather than closing or freezing the last frame.
	- Stage preview begins at a random valid viewport and jumps within the same stage after 30 seconds without visible updates.
	- A manual Next camera action performs the same random viewport jump; free panning is deferred.
- 2026-07-02: Refined the admin console UI in `server/main`.
	- Added responsive panel/table styling and removed template inline styles.
	- Player and stage names are now direct links to query-addressable detail modals (`?player=` and `?stage=`).
	- Watch controls were initially rendered as disabled affordances, then connected when the live-preview transport was implemented.
	- Stage modal data comes from the same in-memory stage snapshot used by the table.
	- Polling note: a 3-5 second metadata refresh is reasonable at current scale if it uses a small dedicated endpoint, pauses in hidden tabs, and backs off on errors. Re-fetching the entire admin page repeatedly would waste HTML bandwidth, repeat Mongo reads while a player modal is open, and add avoidable world/stage lock traffic. Canvas preview should use differential WebSocket updates rather than polling full snapshots.
- 2026-03-28: Admin console follow-up hardening + UX refinement.
	- Fixed duplicate logout panic (`close of closed channel`) by making logout completion idempotent using per-player atomic guards.
	- Added regression test `TestCompleteLogout_Idempotent` in `server/main/world_test.go` to prevent double-logout crashes.
	- Updated admin controls to one combined kick/ban section:
		- Kick accepts minutes (`0` allowed).
		- Kick with minutes > 0 applies temporary lockout (stored as timed ban fields) when an authorized user record exists.
		- Ban is always permanent and still kicks immediately.
- 2026-03-28: Implemented Admin Console MVP in `server/main`.
	- Added admin routes: `/admin`, `/admin/player/update`, `/admin/player/ban`.
	- Added admin-gate config via `ADMIN_IDENTIFIERS` (comma-separated user identifiers like `google:123...`).
	- Added server-rendered admin template with:
		- live logged-in players list + linkable player details,
		- active stages list,
		- live in-memory session summary,
		- stat editing (money, health, accomplishments, team, location),
		- ban action (immediate kick + optional duration).
	- Added Mongo audit trail collection `adminActions` and logging for admin update/ban actions.
	- Added authorized-user ban fields and login-time ban enforcement.
	- Full `go test ./...` is environment-blocked when Mongo is unavailable; compile validation succeeded with `go test -run '^$' ./...`.
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
- 2026-03-26: Completed Cooler Visuals / Dynamic Tiles in `server/main/assets/canvas.js` and `server/main/assets/ws.js`.
	- New tile tokens:
		- `cycle(colorA,colorB)` animated fill cycling between two palette colors.
		- `cycle-b(colorA,colorB)` animated border cycling between two palette colors.
		- `rainbow` animated rainbow fill.
		- `rainbow-b` animated rainbow border.
		- `water` blue wave-like tile shading with soft border.
		- `sparkle` procedural twinkle overlay.
	- Design decisions:
		- Reused class-string token model so new visuals can be authored without schema or websocket payload changes.
		- Dynamic tiles and weather share a single animation loop so both effects can coexist and stay in sync.
		- Redraw scope during animation is restricted to currently visible dynamic tiles (plus weather layer when active) for performance.
	- Current limitations:
		- No per-token speed/intensity parameters yet (fixed animation timings).
		- `cycle(...)` and `cycle-b(...)` only accept named colors present in `COLOR_MAP`.
		- Sparkles are deterministic per tile but use a simple procedural pattern; no authored sparkle masks yet.
	- Future expansion possibilities:
		- Add optional speed/intensity args (examples: `rainbow@slow`, `sparkle(heavy)`, `water(choppy)`).
		- Add directional flow tokens for water (`water-east`, `water-west`) and foam edge modes.
		- Extend dynamic token parsing into tools editor previews so authors can see animation while editing.
- 2026-03-26: Added new escape prototype set `tools/main/data/collections/escape/prototypes/dynamic-tiles.json` with sample assets for all new dynamic tile tokens (`cycle`, `cycle-b`, `rainbow`, `rainbow-b`, `water`, `sparkle`) plus a combined `water-sparkle` tile.
- 2026-03-26: Updated dynamic water behavior + content pass:
	- `water` token in `server/main/assets/canvas.js` no longer auto-adds a dark border; water is now borderless unless an explicit border class is provided.
	- Added `water-shimmer-corner` and `water-sparkle-corner` prototypes to `escape/prototypes/dynamic-tiles.json` and made `water-sparkle` borderless/non-walkable.
	- Redesigned `escape/fragments/water.json` fragment `pond-bend` to use dynamic shimmer/sparkle water prototypes and new rounded-corner variants.
