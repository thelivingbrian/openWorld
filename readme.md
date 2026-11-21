## Introduction 

A webserver (./server) and asset manager (./tools) - used to build/deploy/host 2D web based multiplayer worlds. 

Players view the world as rendered HTML/css after connecting over HTTP/WebSocket - Modest server hardware should support hundreds of players.

Check out: https://bloopworld.co - For live demo


## Build 
    # Server 
        - Must have mongo-db to connect to named "bloopdb"
        - Must have "./data/areas.json" starting asset file (produced using tools)
        - Compile executable with go & run
    # Tools 
        - Compile executable with go & run 
        - Angular SPA editor (new):
            - Build client for hosted server path: `cd tools/main/spa && npm install && npm run build -- --base-href /design/`
            - Start tools server from `tools/main` and open `http://localhost:4444``
        - Deploy changes:
            -Web: visit localhost:4444 with application running, view chosen collection and click 'deploy' at top
            -linux: go build && ./main deploy [collection name]
            -powershell: go build; .\main.exe deploy [collection name]
        - Track changes using git 


## Snapshots 
This project uses: https://github.com/gkampitakis/go-snaps

To update snapshots once (Powershell) use:

        $env:UPDATE_SNAPS = 'true'; go test; Remove-Item Env:\UPDATE_SNAPS

## Hosted world platform

Production runs the server in controller mode. The controller serves `/design/`, the public `/worlds` directory, JSON APIs, and path-routed runtimes under `/w/{world-id-or-slug}`. Published worlds run as isolated child processes using immutable release artifacts cached beneath `WORLD_CACHE_DIR`.

The packaged systemd service supplies the standard paths. For a local controller, set:

    OPENWORLD_MODE=controller
    WORLD_PLATFORM_ENABLED=TRUE
    WORLD_SEED_DIR=../../tools/main/data/collections
    WORLD_DESIGN_DIR=../../tools/main/spa/dist/spa/browser
    WORLD_CACHE_DIR=./data/world-cache

Optional runtime limits are `WORLD_RUNTIME_MEMORY_LIMIT` (Go memory-limit syntax such as `512MiB`) and `WORLD_RUNTIME_GOMAXPROCS`.

## World editor workflow

Sign in with an account listed in `ADMIN_IDENTIFIERS` (comma-separated provider identifiers), then open `/admin/worlds`. Only administrators can create, read drafts, edit, publish, launch, or stop hosted worlds. The console lists unpublished and published worlds, their deployment mode and live player count, with links to the editor and each running world's player/stage console.

1. Create a world from Bloop or Escape, or open an existing world.
2. Edit terrain, prototypes, fragments, interactables and colors in the World and asset tabs. Use **Save all** (`Ctrl/Cmd+S`) to save changed resources together. Paint operations support **Undo paint** (`Ctrl/Cmd+Z`), **Redo** (`Ctrl/Cmd+Shift+Z`), and zoom. A failed save keeps local edits and prevents publishing.
3. In **Settings**, choose the player entry, teams, respawn locations and player limit. Locations must reference walkable tiles. Existing player records whose locations no longer exist fall back to the configured spawn.
4. In **NPCs**, create an NPC and arrange a repeating program. Actions include cardinal movement, wandering, chasing nearby enemy players, attacking and waiting. Each step can run always, below half health, or when an enemy player is nearby. Tick count and interval control timing; attacks specify radius and damage.
5. In **Spawns**, place NPCs, money, boosts or power-ups on a stage, at fixed coordinates or random walkable tiles. Rules run once per stage instance or on player entry with a cooldown. NPC rules have a shared capacity per stage; programmed NPCs also have a world-wide cap of 500 and a maximum lifetime of one hour. Existing area spawn presets are available in **Edit Details**; select **None (custom rules only)** to disable the preset.
6. In **Achievements**, define exploration or statistic milestones. Awards persist in each player's world profile and appear in the accomplishments menu. Keep achievement IDs stable when changing their names or descriptions.

**Save all** updates the editable draft. **Save version** in **Versions & deployment** validates and stores a named, immutable snapshot without selecting it. **Publish** saves pending edits, validates the world and selects a new version. **Launch** applies the selected version and deployment mode; it does not publish unsaved changes. Changing the live version reconnects players.

To roll back, select an older version and launch it. **Restore draft** replaces the editable source from that version without changing the running world. Save a version first to preserve work you intend to keep. Restores use a generation check and an atomic draft switch; conflicting edits produce an error instead of silently overwriting another editor's resource. Versions are retained for explicit rollback and are not automatically pruned.

Choose **Keep running until shut down** for persistent deployment. The controller restores the deployed version after restart, even if another version has since been selected. **Shut down** disconnects players and keeps the world off across controller restarts until an administrator launches it again. The other modes run while the owner is present (with a 60-second grace period), or until the world empties after the owner leaves.

The offline tools server listens on `127.0.0.1:4444`. NPCs, spawns and achievements are saved in each collection's `manifest.json` and included in local compilation/deployment. Hosted version and runtime controls are available through the controller; offline sources can continue to use Git for version history.


## Willow Frontier review

The Bloop seed now has a standalone multiplayer loop: **spawn, collect, explore/fight, extend a streak, die, and quickly return**. Health decreases through combat and hazards; only death restores it. There is no automatic drain, kill healing, or streak banking. Safe spaces preserve current health and inventory. Player and NPC kill counters remain separate; the existing streak includes hostile defeats of either type.

Run the real compiler and game server locally without MongoDB or sign-in:

```sh
cd server/main
go run . --preview
```

Open **http://127.0.0.1:4445/review** to choose Sky Blue or Fuchsia in separate tabs and inspect live population counts/countdowns. Direct links are `http://127.0.0.1:4445/?team=sky-blue` and `http://127.0.0.1:4445/?team=fuchsia`. Preview players are temporary; reloading creates a new test player. Production profiles keep their existing persistence behavior. The preview never publishes or changes database records. Set `OPENWORLD_PREVIEW_PORT` to use another loopback port if needed. Restart after Go server changes; refresh the browser after renderer changes. It compiles the local Bloop collection, palette, and manifest into memory at startup, then runs the same gameplay server, WebSocket updates, and renderer used by normal worlds; restart it to load saved editor changes. The browser tab title “Willow Reach · Bloopworld” is hardcoded in `preview.go`. The world name “Bloopworld · Willow Frontier” comes from the saved manifest and is editable in Settings; region names in the content are separate.

WASD/arrows or the direction pad move; Space/Power uses a collected red power; Shift + direction or the Boost toggle uses blue boosts. M/Escape opens the menu. Optional **How to play** help replaces a mandatory tutorial. Team symbols supplement color. The compact HUD shows health, powers, boosts, and zone safety; the fullscreen button and browser-mode layout support short mobile viewports.

Review route:

1. Each team starts in its own 12×12 safe refuge at (Y7,X6). Walk north to collect a boost and a power, then leave at (3,6). Supplies replenish while people remain in the room.
2. Explore the 48×48 shared frontier. Orange wild bloops chase and attack; allies cannot hurt one another. Defeat enemies with powers, collect new supplies, and continue without crossing an area boundary to trigger spawning. A brief 1.5-second arrival shield after leaving safety ends immediately when attacking.
3. Both the village bridge and north crossing connect the riverbanks. Baseline power/boost caches remain available while the announced **North Crossing / Riverbank / Seed Garden** bonus site rotates every minute. Health does not regenerate when visiting a cottage or refuge.
4. Visit the snowy northeast grove to review rounded snow boundaries, snow pine silhouettes, and scenery bases that match the ground.
5. Enter the dark annex from the southern frontier portal at (44,17). The foyer's left portal (5,4) leads to the solo room; its right portal (5,11) leads to the optional two-player relay. Foyer exits (13,6) and (13,9) lead to the labeled team refuges without healing or changing a streak.
6. In the solo room, push the block north onto (9,5), then cross the gate at row 6 to the goal at (4,8). In the duo room, two players of any teams stand on (9,5) and (9,11) together. Completion gives recognition, not health, supplies, money, or competitive score. Gates close when plates are released; they never close on an occupied gate tile. Exits remain usable with gates closed.
7. **Menu → Reset chamber** requires agreement from everyone currently in that room; empty chambers reset automatically. Ordinary turns do not juke blocks in puzzle rooms. Boost becomes a single step there without consuming charges, so it cannot bypass a gate. Disconnect a partner to check recovery and gate closure.

The new world has no links to legacy Bloopworld. Legacy seed areas remain available in source for separate worlds/releases. Existing published versions are unchanged until explicitly published and launched.

Population defaults are editable in **Spawns**: recurring/entry/once triggers, regions, minimum player distance, counts/caps, lifetimes, and rotation labels/groups. There are 21 recurring rules. Hostiles replenish every 18 seconds with a maximum of three per region (nine total), a two-minute lifetime, and a five-tile minimum spawn distance. Starter supplies return every 12–15 seconds; baseline supplies return every 18–24 seconds. NPCs spawn on unoccupied connected terrain, outside safe stages and transports. Pickup expiry preserves unrelated dropped items. One world-owned scheduler services occupied stages and stops on shutdown.

The editor's **Settings → Safe puzzle chambers** controls plates, gates, goals, and player requirements. Use safe source stages and blocked gate tiles. The **Anchors & visual bounds** and **Blocked footprint** toggles explain scenery placement; Edit Transports marks entrances. The fragment shelf provides labeled thumbnails of complete fixtures. Gameplay and editor share `server/main/assets/scenery.js`; fixtures keep a fixed viewing direction. Unknown scenery tokens and rotated scenery anchors fail compilation. Current catalog extents fit the four-tile camera margin.

To review the editor, build with `npm run build -- --base-href /design/` from `tools/main/spa`, then start `TOOLS_PORT=4446 go run .` from `tools/main` and open http://127.0.0.1:4446/design/. Select the `bloop` collection and `willow-*` spaces/sets. Rebuild the deterministic native source with `python3 tools/worlds/build_willow_reach.py`; preserve manual editor changes before regenerating.

A ten-minute automated two-team session observed first pickup at one second, 19 death/respawn cycles, at most nine hostile NPCs and twelve rule pickups, with no session errors. This confirms bounded replenishment and repeatable play, not final difficulty balance.

Validation and captures are in `.cache/frontier-review/`: real browser routes for both teams, solo/duo completion, partner disconnect, desktop/mobile dark rooms, winter, editor controls, and population session telemetry. Focused Go race tests cover replenishment over fifteen simulated minutes, caps/exclusion, safe rooms, puzzle reset/completion, gate bypass prevention, and release serialization. All 98 Angular tests and 11 snapshots pass, as do the tools Go suite and four canvas/shell tests. The editor production build passes with initial-bundle and stylesheet-budget warnings. The previous authorization-test failure was fixed by denying unmatched admin identifiers.

Before release, test physical iOS/Android browser bars, safe areas, touch comfort and frame rate; assess team balance, repeated-victim farming, and competitive sight ranges. These remain playtest/release checks rather than claims established by automation. Day/night, weather, broader biome expansion, and general sprite import remain follow-ups.


## AI orchestration

- Stable instructions: `.ai/AGENTS.md`
- Architecture reference: `.ai/architecture.md`

Suggested flow:
1. Read `.ai/AGENTS.md`, then `.ai/architecture.md`
2. Complete the active task
3. Update `.ai/architecture.md` when a lasting project property changes
