# AI Backlog

Shared queue for human + agent execution.

Status values: `TODO` | `IN_PROGRESS` | `BLOCKED` | `DONE`

## Current In-Progress (single primary)

| Task ID | Title | Owner | Started | Notes |
|---|---|---|---|---|
| _none_ | _none_ | _unassigned_ | _n/a_ | Last complete: AI-015 (2026-03-14). Set next task to `IN_PROGRESS` and mirror it here. |

## Backlog

| Task ID | Status | Priority | Area | Title | Owner | Updated | Notes |
|---|---|---|---|---|---|---|---|
| AI-001 | DONE | P1 | Repo | Add AI orchestration docs (`AGENTS.md`, `.ai/notes.md`, `AI_BACKLOG.md`) | agent | 2026-03-14 | Initial scaffolding created. |
| AI-002 | TODO | P2 | Process | Add first real engineering task for this workflow | unassigned | 2026-03-14 | Promote to `IN_PROGRESS` when started. |
| AI-003 | DONE | P1 | Interactables | Add mutable `state` property for interactables | copilot | 2026-03-14 | Added server model+persistence support and test `TestCreateStageFromAreaLoadsMutableInteractableState`. |
| AI-004 | DONE | P1 | Tools | Support setting interactable `state` in design workspace | copilot | 2026-03-14 | Added tools model + SPA editing/normalization support for `state`. |
| AI-005 | DONE | P1 | Interactables | Use `state` as `reactsWith` gate (`is`/`is not`/`contains`) | copilot | 2026-03-14 | Added new state gate functions, registry wiring, design registry entries, and focused tests. |
| AI-006 | DONE | P1 | Interactables | Implement transmit push movement | copilot | 2026-03-14 | Added `transmitPushAll` reaction with offset-aware runtime wiring, editor registry entry, and focused tests. |
| AI-007 | TODO | P2 | Interactables | Add transmit push reaction with `nil` | unassigned | 2026-03-14 | Clarify reset behavior on no reaction target. |
| AI-008 | TODO | P1 | Interactables | Send push to other interactables by `state` or `type` | unassigned | 2026-03-14 | Requires target selection strategy. |
| AI-009 | TODO | P2 | Interactables | Allow rotation/scale of transmitted push vector | unassigned | 2026-03-14 | Decide transform syntax compatibility. |
| AI-010 | TODO | P1 | Physics | Implement sticky blocks that stick together | unassigned | 2026-03-14 | Parent item for grouped movement. |
| AI-011 | TODO | P1 | Physics | Add polyomino-based pushing logic | unassigned | 2026-03-14 | Core requirement for sticky groups. |
| AI-012 | TODO | P2 | Interactables | Decide/implement pushable interactables behavior | unassigned | 2026-03-14 | Scoped as sub-task of polyomino push logic. |
| AI-013 | TODO | P2 | Content | Create new puzzle space in `collection:escape` demonstrating features | unassigned | 2026-03-14 | Should showcase new state/transmit/sticky mechanics. |
| AI-014 | DONE | P1 | Interactables | Interactable state as full configuration + per-tile state selection | copilot | 2026-03-14 | Added state-config model, per-tile state selection in editor grid, compile resolution, and server runtime application.
| AI-015 | DONE | P1 | Tools | Fix interactable add-set UX and validation | copilot | 2026-03-14 | Reordered controls and added explicit blank/duplicate name validation + success feedback; added focused editor component tests. |
| AI-016 | DONE | P2 | Tools | Fix stale grid-engine normalizeTile test expectation | copilot | 2026-03-14 | Updated spec to expect `interactableState` in normalized tile output; full SPA Jest suite passes. |

## Protocol
- Pick one `TODO` item and set to `IN_PROGRESS`.
- Copy that row into `Current In-Progress`.
- If blocked, set status to `BLOCKED` and add unblock condition.
- On completion, set to `DONE`, add short outcome note, and clear `Current In-Progress`.
