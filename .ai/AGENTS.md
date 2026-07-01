# AGENTS.md

Stable operating instructions for AI agents working in this repository.

## Purpose
- Keep this file stable and high-signal.
- Put temporary findings and evolving context in `.ai/notes.md`.
- Track shared execution work in `.ai/backlog.md`.
- Track evolving properties, architectural decisions, and constraints in `.ai/architecture.md`.
- Track known issues that provide context but should not be fixed unless explicitly requested in `.ai/known-issues.md`.

## Priority and source of truth
1. User request in the active session.
2. System/developer runtime instructions.
3. This file (`.ai/AGENTS.md`).
4. Mutable collaboration files (`.ai/notes.md`, `.ai/backlog.md`, `.ai/architecture.md`, `.ai/known-issues.md`).
5. Existing code/tests/docs.

If instructions conflict, follow the higher-priority source.

## Working agreements
- Make targeted changes with terse, clever, readable code.
- Fix root causes when practical; avoid cosmetic churn.
- Preserve existing naming/style unless asked to refactor.
- Define new go functions on lines ~below~ where they are first referenced 
- Do not modify unrelated files.
- Validate changes with focused tests when available.

## Collaboration protocol
- Before coding: read `.ai/backlog.md`, `.ai/notes.md`, `.ai/known-issues.md`, and `.ai/architecture.md`.
- Observe the currently in progress task for instructions
- If an encountered issue appears in `.ai/known-issues.md`, do not fix it unless the active user request explicitly includes that issue.
- After coding: check completed tasks off, leave a note for items that cannot be completed and leave them unchecked. Append notes you believe will aid future agents to `.ai/notes.md`, and update `.ai/architecture.md` when new properties/decisions/constraints are discovered.

## File ownership model
- `.ai/AGENTS.md`: stable policy; change only when workflow rules truly change.
- `.ai/notes.md`: agent-editable working memory (safe to rewrite sections).
- `.ai/backlog.md`: shared work queue and execution state.
- `.ai/architecture.md`: evolving architecture reference (known properties, decisions, constraints).
- `.ai/known-issues.md`: issue context and explicit do-not-fix list unless requested.

## Definition of done
- Code changes compile or tests pass for the touched scope (when feasible).
- In progress Backlog item(s) all are checked or include relevant note why they cannot be complete
- Relevant learning captured in `.ai/notes.md`.
