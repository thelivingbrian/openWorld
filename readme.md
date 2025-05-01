## Introduction 

A webserver (./server) and asset manager (./tools) - used to build/deploy/host 2D web based multiplayer worlds. 

Players view the world as rendered HTML/css after connecting over HTTP/WebSocket - Modest server Hardware should support hundred of players.

Check out: https://bloopworld.co - For live demo


## Build 
    # Server 
        - Must have mongo-db to connect to named "bloopdb"
        - Must have "./data/areas.json" starting asset file (produced using tools)
        - Compile executable with go & run
    # Tools 
        - Compile executable with go & run 
        - Track changes in git 
        - Deploy changes 
            -linux: go build && ./main deploy bloop
            -powershell: go build; .\main.exe deploy bloop


## Snapshots 
This project uses: https://github.com/gkampitakis/go-snaps

To update snapshot once (Powershell) use: 
$env:UPDATE_SNAPS = 'true'; go test; Remove-Item Env:\UPDATE_SNAPS