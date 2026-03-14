# AGENTS.md

Stable operating instructions for AI agents working in this repository.

## Purpose
- Keep this file stable and high-signal.
- Put temporary findings and evolving context in `.ai/notes.md`.
- Track shared execution work in `AI_BACKLOG.md`.

## Priority and source of truth
1. User request in the active session.
2. System/developer runtime instructions.
3. This file (`AGENTS.md`).
4. Mutable collaboration files (`.ai/notes.md`, `AI_BACKLOG.md`).
5. Existing code/tests/docs.

If instructions conflict, follow the higher-priority source.

## Working agreements
- Make minimal, targeted changes.
- Fix root causes when practical; avoid cosmetic churn.
- Preserve existing naming/style unless asked to refactor.
- Do not modify unrelated files.
- Validate changes with focused tests when available.

## Collaboration protocol
- Before coding: read `AI_BACKLOG.md` and `.ai/notes.md`.
- Claim one backlog item by setting it to `IN_PROGRESS` and filling owner/date.
- Keep exactly one primary in-progress task at a time in the `Current In-Progress` block.
- After coding: update backlog status and append notes to `.ai/notes.md`.

## File ownership model
- `AGENTS.md`: stable policy; change only when workflow rules truly change.
- `.ai/notes.md`: agent-editable working memory (safe to rewrite sections).
- `AI_BACKLOG.md`: shared work queue and execution state.

## Definition of done
- Code changes compile or tests pass for the touched scope (when feasible).
- Backlog item moved to `DONE` with a short outcome note.
- Relevant learning captured in `.ai/notes.md`.
