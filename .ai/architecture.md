# Architecture Notes

High-level architecture reference for agents and contributors.

## Components
- `server/main`: Go gameplay server runtime (HTTP/WebSocket, world logic).
- `tools/main`: Go authoring/deployment toolchain with Angular SPA editor.
- `integration/main`: integration runner and scripts.
- `deploy/droplet`: deployment scripts and service definitions.

## Data Flow (high level)
1. World content is authored in `tools/main` collections.
2. Compiled/exported data is consumed by the server runtime.
3. Clients connect via HTTP/WebSocket and render with assets in `server/main/assets`.

## Operational Notes
- Keep design changes in tools and runtime behavior changes in server aligned.
- Prefer focused tests near changed packages before broad test runs.
