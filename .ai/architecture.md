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
- Server is used by multiple players concurrently, changes must prioritize stability then performance.
- Prefer focused tests near changed packages before broad test runs.

## Rendering Notes
- Canvas rendering lives in `server/main/assets/canvas.js` with per-layer stage buffers and immediate per-tile updates from websocket quick-swap messages.
- Weather tint and dynamic weather effects share the `Lw1` layer:
	- Static weather remains class-driven (`blue trsp20`, etc.) via existing color/transparency token parsing.
	- Dynamic weather modes are token-driven extensions in the same weather string (example: `raining`) and rendered client-side by mode-specific frame functions.
	- Current implementation redraws only `Lw1` during weather animation frames to avoid touching gameplay/entity layers.
