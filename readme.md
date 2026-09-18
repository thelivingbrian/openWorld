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


## AI orchestration

- Stable instructions: `.ai/AGENTS.md`
- Architecture reference: `.ai/architecture.md`

Suggested flow:
1. Read `.ai/AGENTS.md`, then `.ai/architecture.md`
2. Complete the active task
3. Update `.ai/architecture.md` when a lasting project property changes