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
            - Build client: `cd tools/main/spa && npm install && npm run build`
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


## AI orchestration

- Stable instructions: `.ai/AGENTS.md`
- Mutable agent notes: `.ai/notes.md`
- Shared execution queue: `.ai/backlog.md`

Suggested flow:
1. Read `.ai/AGENTS.md`, then `.ai/backlog.md`, then `.ai/notes.md`
2. Claim one task by setting it to `IN_PROGRESS` and filling owner/date
3. Complete work, set task to `DONE`, and add key learnings to `.ai/notes.md`
