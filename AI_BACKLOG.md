# AI Backlog

Shared queue for human + agent execution.

Status values: `TODO` | `IN_PROGRESS` | `BLOCKED` | `DONE`

## Current In-Progress (single primary)

| Task ID | Title | Owner | Started | Notes |
|---|---|---|---|---|
| _none_ | _none_ | _unassigned_ | _n/a_ | Last complete: AI-020 (2026-03-14). Set next task to `IN_PROGRESS` and mirror it here. |

## Backlog

| Task ID | Status | Priority | Area | Title | Owner | Updated | Notes |
|---|---|---|---|---|---|---|---|
| AI-001 | DONE | P1 | Repo | Add AI orchestration docs (`AGENTS.md`, `.ai/notes.md`, `AI_BACKLOG.md`) | agent | 2026-03-14 | Initial scaffolding created. |
| AI-002 | TODO | P2 | Process | Add first real engineering task for this workflow | unassigned | 2026-03-14 | Promote to `IN_PROGRESS` when started. |
| AI-003 | DONE | P1 | Interactables | Add mutable `state` property for interactables | copilot | 2026-03-14 | Added server model+persistence support and test `TestCreateStageFromAreaLoadsMutableInteractableState`. |
| AI-004 | DONE | P1 | Tools | Support setting interactable `state` in design workspace | copilot | 2026-03-14 | Added tools model + SPA editing/normalization support for `state`. |
| AI-005 | DONE | P1 | Interactables | Use `state` as `reactsWith` gate (`is`/`is not`/`contains`) | copilot | 2026-03-14 | Added new state gate functions, registry wiring, design registry entries, and focused tests. |
| AI-006 | DONE | P1 | Interactables | Implement transmit push movement | copilot | 2026-03-14 | Added `transmitPushAll` reaction with offset-aware runtime wiring, editor registry entry, and focused tests. |
| AI-007 | DONE | P2 | Interactables | Add transmit push reaction with `nil` | copilot | 2026-03-14 | Confirmed `interactableIsNil -> transmitPushAll` wiring and fixed right/down over-push bug by processing tiles in direction-aware order; added focused regressions. |
| AI-008 | DONE | P1 | Interactables | Send push to other interactables by `state` or `name` | copilot | 2026-03-14 | Added `transmitPushByState` and `transmitPushByName` reactions (direction-safe ordering retained), wired SPA registry entries, and added focused server tests for resolution and selective movement. |
| AI-009 | TODO | P2 | Interactables | Allow rotation/scale of transmitted push vector | unassigned | 2026-03-14 | Decide transform syntax compatibility. |
| AI-010 | TODO | P1 | Physics | Implement sticky blocks that stick together | unassigned | 2026-03-14 | Parent item for grouped movement. |
| AI-011 | TODO | P1 | Physics | Add polyomino-based pushing logic | unassigned | 2026-03-14 | Core requirement for sticky groups. |
| AI-012 | TODO | P2 | Physics | Decide/implement pushable fragment behavior | unassigned | 2026-03-14 | Scoped as sub-task of polyomino push logic. |
| AI-013 | TODO | P2 | Content | Create new puzzle space in `collection:escape` demonstrating features | unassigned | 2026-03-14 | Should showcase new state/transmit/sticky mechanics. |
| AI-014 | DONE | P1 | Interactables | Interactable state as full configuration + per-tile state selection | copilot | 2026-03-14 | Added state-config model, per-tile state selection in editor grid, compile resolution, and server runtime application.
| AI-015 | DONE | P1 | Tools | Fix interactable add-set UX and validation | copilot | 2026-03-14 | Reordered controls and added explicit blank/duplicate name validation + success feedback; added focused editor component tests. |
| AI-016 | DONE | P2 | Tools | Fix stale grid-engine normalizeTile test expectation | copilot | 2026-03-14 | Updated spec to expect `interactableState` in normalized tile output; full SPA Jest suite passes. |
| AI-017 | DONE | P1 | Tools | Fix SPA interactable set dropdown/update and new interactable editability | copilot | 2026-03-14 | Fixed `touchBootstrap` to clone current collection/maps so computed lists refresh, and fixed `addInteractable` to re-sync editable state selection; added focused editor component regressions. |
| AI-018 | DONE | P1 | Tools | Fix first interactable visibility in brand-new set | copilot | 2026-03-14 | Changed `addInteractable` to replace set array immutably (instead of in-place push), so first interactable appears immediately; added focused regression test. |
| AI-019 | DONE | P2 | Interactables | Verify transmitPushAll concurrency safety | copilot | 2026-03-14 | Added concurrent deadlock-focused `transmitPushAll` test (simultaneous opposite directions) and targeted transmit tests pass; `-race` requires local `gcc` toolchain (not available in current env). |
| AI-020 | DONE | P1 | Tools | Fix SPA interactable state editing UX | copilot | 2026-03-14 | Added state rename control/action in editor and fixed cross-state reaction rule coupling by deep-cloning state config/rule arrays when creating new states; added focused SPA regression tests. |

## Protocol
- Pick one `TODO` item and set to `IN_PROGRESS`.
- Copy that row into `Current In-Progress`.
- If blocked, set status to `BLOCKED` and add unblock condition.
- On completion, set to `DONE`, add short outcome note, and clear `Current In-Progress`.
