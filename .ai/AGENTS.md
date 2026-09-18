# AGENTS.md

Stable operating instructions for AI agents working in this repository.

## Purpose
- Keep this file stable and high-signal.
- Track evolving properties, architectural decisions, and constraints in `.ai/architecture.md`.

## Priority and source of truth
1. User request in the active session.
2. System/developer runtime instructions.
3. This file (`.ai/AGENTS.md`).
4. `.ai/architecture.md`.
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
- Before coding: read `.ai/architecture.md`.
- Observe the currently in progress task for instructions.
- After coding: update `.ai/architecture.md` when new properties, decisions, or constraints are discovered.

## File ownership model
- `.ai/AGENTS.md`: stable policy; change only when workflow rules truly change.
- `.ai/architecture.md`: evolving architecture reference (known properties, decisions, constraints).

## Definition of done
- Code changes compile or tests pass for the touched scope (when feasible).
