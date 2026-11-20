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


## Willow Reach review

The Bloop seed now opens in **Willow Reach**, a 48×48 woodland village with three enterable cottages, a river bridge, a pine/snow grove, a village fountain, and a pushable seed garden. The old-world portal has a separate confirmed return tile. Existing published releases are unchanged until a new version is published and launched.

Run the real compiler and game server locally without MongoDB or sign-in:

```sh
cd server/main
go run . --preview
```

Open **http://127.0.0.1:4445**. Preview progress is temporary. Reload to start again; open another tab to test multiple players. Use WASD/arrows to move, Shift + direction to boost when stocked, Space to use a collected power, and M/Escape for the menu. Touch controls support holding a direction and using Power simultaneously; Boost toggles boosted movement. The upper-right fullscreen button enters/exits fullscreen where the browser permits it; browser mode uses the visible viewport and safe-area insets.

Review route: start at `willow-reach:0-0` (y23, x17), visit the cottage northwest at (19,13), leave through its south door, walk behind its roof, then cross the bridge east to the pine grove. The seed garden is southwest; push the golden seed toward the ring. The Old Road portal is at (44,17); in the old world, the return tile is at (9,10), next to the arrival tile. These coordinates are for locating review cases in the editor.

The source is ordinary Bloop collection data in the `willow-reach` prototype/fragment sets and `willow-reach` / `willow-cottage` spaces. The editor uses the same `server/main/assets/scenery.js` catalog as gameplay. Rebuild the editor with `npm run build -- --base-href /design/` from `tools/main/spa`. If another tools server is already running, start this checkout with `TOOLS_PORT=4446 go run .` from `tools/main` and open http://127.0.0.1:4446/design/.

Scenery authoring rules:

- The anchor is the center-bottom of its ground tile. Artwork extends upward; the catalog records visual width/height in tiles. Light comes from the upper left, with short shadows toward the lower right.
- Place complete fragments to preserve collision: a tree blocks one trunk tile; cottages use a 5×2 footprint with a walkable bottom-center doorway; the fountain uses a 3×2 footprint. Add a transport at each new cottage doorway and a distinct return tile outside it.
- Keep the three-quarter viewing direction fixed. The editor disables rotation of scenery fixtures. Use rounded silhouettes, restrained `willow-*` colors, and paths that remain legible under foliage. New scenery uses `scenery-*` tokens in Floor2; legacy ceiling layers keep their existing behavior.
- The deterministic `python3 tools/worlds/build_willow_reach.py` script rebuilds this authored seed and verifies connectivity. It replaces Willow assets and updates the old-world return tile; preserve editor changes before regenerating.

Review validation: all 97 Angular tests and 11 snapshots, the tools Go suite, four canvas/shell tests, and focused server/compiler/camera tests with the race detector passed. Browser checks covered desktop, 375×667 portrait, 667×375 and 667×300 landscape (reduced height for browser chrome), rotation, simultaneous touch input and cancellation, menu focus, cottage entry/return, garden-seed pushing, old-world travel in both directions, and fullscreen entry/exit. Local captures and a movement video are in `.cache/review/`. The full server suite is not green: `TestWorldAuthoringRequiresAdmin` panics in the unchanged baseline as well. The editor build also retains its existing stylesheet-budget warning.

Before release: test physical iOS/Android browser bars, notches, touch comfort, and frame rate; review competitive sight ranges with the responsive viewport. Day/night, weather cycles, and conversion of the remaining old regions are follow-up work. The review scene remains bright and readable without those effects.


## AI orchestration

- Stable instructions: `.ai/AGENTS.md`
- Architecture reference: `.ai/architecture.md`

Suggested flow:
1. Read `.ai/AGENTS.md`, then `.ai/architecture.md`
2. Complete the active task
3. Update `.ai/architecture.md` when a lasting project property changes
