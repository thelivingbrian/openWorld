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

## Admin Console Notes
- Admin console is server-rendered in `server/main` and exposed via `/admin` on game server instances.
- Admin access is identifier-based (`ADMIN_IDENTIFIERS` env var, comma-separated).
- Admin write actions persist audit records in Mongo collection `adminActions`.
- User bans are stored on user records and enforced at login (`banReason`, `bannedBy`, `banExpiresAt`).

## Rendering Notes
- Canvas rendering lives in `server/main/assets/canvas.js` with per-layer stage buffers and immediate per-tile updates from websocket quick-swap messages.
- Weather tint and dynamic weather effects share the `Lw1` layer:
	- Static weather remains class-driven (`blue trsp20`, etc.) via existing color/transparency token parsing.
	- Dynamic weather modes are token-driven extensions in the same weather string (example: `raining`) and rendered client-side by mode-specific frame functions.
	- Current implementation redraws only `Lw1` during weather animation frames to avoid touching gameplay/entity layers.
- Dynamic tile visuals are also class-token driven in `canvas.js` and rendered client-side:
	- Tokens: `cycle(colorA,colorB)`, `cycle-b(colorA,colorB)`, `rainbow`, `rainbow-b`, `water`, `sparkle`.
	- Runtime scans visible tiles for dynamic tokens and redraws only affected visible layers each frame.
	- WebSocket immediate draws now pass world coordinates to tile draw calls so animated phase offsets remain stable on quick-swaps.
