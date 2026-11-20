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

## Operational World Platform
- The deployed server starts in `controller` mode. It hosts authentication, `/design/`, world APIs, the public directory, and `/w/{id-or-slug}` reverse proxying.
- Each running world is a child process of the controller and loads exactly one immutable release from a read-only, content-addressed local cache. A world remains the scheduling and failure-isolation unit.
- Editable source is stored as revisioned `worldResources`. Publishing validates a consistent draft, creates one deterministic ZIP containing `areas.json`, `manifest.json`, `maps.json`, per-world CSS, and one base PNG per tiled space, and stores source/artifact bundles in GridFS.
- `worldPlayerProfiles` are represented by player records keyed/indexed with `worldId` and stable `userId`; legacy records are assigned to the reserved `legacy` world during platform initialization.
- Runtime paths are same-origin under `/w/...`; textual runtime responses are base-path rewritten by the controller so existing HTMX/WebSocket/static URLs remain isolated.
- Runtime lifecycle values are `owner-present` (60-second grace), `until-empty`, and admin-only `persistent`. Runtime heartbeats publish player count and owner presence through TTL-indexed Mongo records.
- Repository collections and palettes are packaged as seeds, preserving local/offline authoring while allowing hosted worlds to fork known-good content.
- Hosted design pages, all draft APIs, creation and deployment require an administrator. `/admin/worlds` provides controller-level world navigation; `/w/{id}/admin` remains the live player/stage console. World API writes are audited through `adminActions`.
- `deploymentEnabled` persists explicit launch/shutdown intent; `deployedReleaseId` is independent of the selected `publishedReleaseId`. Persistent restoration uses the deployed version. Runtime start/stop operations are serialized, re-read desired state, and stop restarting once controller shutdown begins.
- Named releases may be saved without selection. Version selection leaves the live process alone until launch. Releases are retained rather than automatically pruned. Draft restoration writes a fresh resource namespace and atomically updates `draftResourceId`; revisions advance past the old generation to reject stale saves. HTTP operations are serialized per world within the single controller.
- `manifest.json` defines player/team spawns, NPC programs, stage-entry spawn rules and world achievements. Publishing validates locations, references and limits. Programmed NPCs execute bounded action sequences with conditions, per-stage spawn capacity, a global cap of 500, and finite lifetimes. Achievement IDs remain stable across display-name changes and awards use the existing world-scoped player profile.
- The editor saves changed resources before publication, keeps failed edits in memory, and separates metadata refresh from draft revision loading. Local authoring persists the same manifest in each collection and exports it with compiled content. The unauthenticated offline tool binds only to loopback.
- Draft palettes drive editor previews and client-generated minimaps. Gameplay and spectator canvas rendering import each world's published CSS palette, including animation colors, and reset palette overrides when switching worlds.

## Operational Notes
- Keep design changes in tools and runtime behavior changes in server aligned.
- Server is used by multiple players concurrently, changes must prioritize stability then performance.
- Prefer focused tests near changed packages before broad test runs.
- Droplet releases are activated by `/opt/openworld/bin/deploy-openworld.sh`, which restarts the configured systemd unit after updating `/opt/openworld/current`; changes to the repository copy must also be installed into `/opt/openworld/bin`.

## Admin Console Notes
- Admin console is server-rendered in `server/main` and exposed via `/admin` on game server instances.
- Admin access is identifier-based (`ADMIN_IDENTIFIERS` env var, comma-separated).
- Admin write actions persist audit records in Mongo collection `adminActions`.
- User bans are stored on user records and enforced at login (`banReason`, `bannedBy`, `banExpiresAt`).

### Live preview
- Admin-only watch pages use `/admin/watch`; differential canvas updates use the authenticated `/admin/watch/screen` WebSocket.
- Player watching mirrors the selected player's canvas viewport and includes highlights.
- HUD and menu forwarding are independently configurable per watch session and default to off.
- When the selected player logs out, the watch session remains open and renders a dark logged-out state with explanatory text.
- Stage watching uses a read-only camera at a random valid viewport on the selected stage.
- A stage camera jumps to another random viewport after 30 seconds without an update visible in its current viewport; a manual Next camera action triggers the same behavior.
- Stage camera jumps stay within the selected stage. Free panning is deferred.
- Spectator camera output is bounded and non-blocking. Queue overflow requests a full viewport resync so a slow spectator cannot block camera-zone gameplay updates.
- Non-admin observer access is deferred; the route-level authorization boundary is the future extension point.

## Rendering Notes
- Hosted compilation preserves a prototype’s explicit `cssColor` over blueprint ground paint, matching the editor and offline compiler. Ground status and corner styling apply only when the prototype has no explicit ground color.
- The editor brand bubble is green when the running release matches the selected `publishedReleaseId`, yellow when another release is running, and gray otherwise.
- Editor palette rules target `.box` with enough specificity to override Angular-scoped tile border defaults; built-in colors remain available when draft palettes are incomplete.
- Bloop and Escape airlocks use explicit door (`closed`/`open`) and close-switch (`disarmed`/`armed`) states with connected-airlock reaction rules. Keep these seed definitions aligned; legacy stateless airlock reactions cannot drive the state transitions. Existing stateless airlocks are not migrated; the proof-of-concept world will use the corrected seeds.
- Canvas rendering lives in `server/main/assets/canvas.js` with per-layer stage buffers and immediate per-tile updates from websocket quick-swap messages.
- Weather tint and dynamic weather effects share the `Lw1` layer:
	- Static weather remains class-driven (`blue trsp20`, etc.) via existing color/transparency token parsing.
	- Dynamic weather modes are token-driven extensions in the same weather string (example: `raining`) and rendered client-side by mode-specific frame functions.
	- Current implementation redraws only `Lw1` during weather animation frames to avoid touching gameplay/entity layers.
- Dynamic tile visuals are also class-token driven in `canvas.js` and rendered client-side:
	- Tokens: `cycle(colorA,colorB)`, `cycle-b(colorA,colorB)`, `rainbow`, `rainbow-b`, `water`, `sparkle`.
	- Runtime scans visible tiles for dynamic tokens and redraws only affected visible layers each frame.
	- WebSocket immediate draws now pass world coordinates to tile draw calls so animated phase offsets remain stable on quick-swaps.


## Three-quarter overworld and responsive play
- `server/main/assets/scenery.js` is the shared procedural scenery catalog. Gameplay loads it directly; Angular packages the same file under its own `assets/` and uses `SceneryDirective`. New `scenery-*` Floor2 tokens specify center-bottom anchors with visual extents in tile units. Collision remains native source tiles/fragments; no release-schema migration is required.
- Gameplay composites scenery and characters in ground-Y order on `Lp1`, with selective fading when a character is behind tall scenery. Legacy ceiling layers retain their previous ordering. Websocket changes batch drawing into animation frames. Named Willow terrain receives deterministic detail and subtle water ripples; reduced-motion users skip animated effects.
- Cameras negotiate bounded rectangular dimensions (8–32 per axis) through the `viewport` event. A `view` quick-swap carries camera and stage dimensions before `set`. Snapshots include a four-tile scenery margin; stage broadcast zones use the maximum supported viewport size. Camera snapshots are batched to avoid filling outbound queues during resizes. Spectator snapshots include the same dimensions and margin.
- `game-shell.js` and `play.css` own visible-viewport sizing, fullscreen state/fallback, floating HUD, safe areas, and touch layout. Pointer capture and lifecycle cancellation prevent held controls from sticking. Initial and refreshed HUDs share the `player-hud` template. Authored transport floor artwork is preserved; unstyled transports retain the legacy pink marker.
- The Bloop seed manifest enters separate `willow-refuge:0-0` / `:0-1` safe team stages at (7,6). The shared frontier is unsafe and 48×48, with three safe cottage interiors and two river crossings. All links to legacy areas have been removed; optional dark annex/solo/duo stages are separate shared safe rooms. New editor spaces default to a single 48×48 area. Scenery fixtures retain a fixed viewing direction.
- `go run . --preview` from `server/main` compiles the seed in memory and starts an ephemeral, database-free runtime on loopback port 4445. It uses the real compiler, templates, websocket movement, collisions and transports. It never publishes or alters database records. `TOOLS_PORT` selects another loopback port for the offline editor when 4444 is already occupied.


## Living frontier rules
- `WorldSpawnRule` supports `interval` alongside legacy `entry`/`once`, optional bounded `region`, minimum player distance, pickup lifetime, and rotating groups with a public label/duration. One cancellable world-owned 500ms clock services occupied stages; recurring work does not depend on stage transitions. Spawn state is per stage/rule, shared across entrants, capped for NPCs and pickups. Rules reject invalid regions/capacities/durations. Candidate tiles exclude occupied/interactable/transport sites; random sites must share walkable terrain with an occupant. NPCs prefer safe-distance positions near occupants when their region permits it.
- Rule-owned pickup tracking distinguishes collection, expiry and rotation changes. Expiring a rule pickup preserves unrelated dropped items. Programmed NPC lifetimes and the world cap remain bounded; hostile defaults use 50 health, short-range attacks and no healing rewards.
- Stage `Safe` is now enforced for both target and attacker. Power use in safe stages leaves inventory unchanged. A 1.5-second protection window after exiting a safe stage ends on Power use. Combat health/streak behavior otherwise retains the existing rules; PvP and NPC counters stay separate. Power and movement inputs have independent cooldowns.
- `WorldManifest.puzzles` describes shared safe chambers with plates, gates, goal, minimum participants, and optional players-only plates. `puzzles.go` evaluates block/player positions under canonical tile-lock ordering. Gate collision uses atomic overrides, while source tiles stay immutable. Gates do not close on occupants. Completion is recognition only, with no economic/health/combat reward. Reset requires all current participants or vacancy and restores authored interactables. Always-accessible transports provide exits. Juke/Boost cannot displace solved blocks or bypass gates inside chambers.
- Designer Settings exposes puzzle definitions; Spawns exposes recurring regions, capacity, distance, expiry and rotation. Overlay toggles mark anchors/visual bounds and collision footprints; the fixture shelf previews complete fragments. Native and hosted compilers reject unknown or rotated scenery tokens. Snow variants preserve their underlying biome and rounded transitions.
- Local preview supports `?team=sky-blue` / `?team=fuchsia`; `/review` and `/review/status` expose local-only player/population counts and countdowns. Production profiles are not involved. New preview loads intentionally create fresh ephemeral test players.
- Authorization checks now deny unmatched identifiers rather than returning true as a fallback; `TestWorldAuthoringRequiresAdmin` passes.
