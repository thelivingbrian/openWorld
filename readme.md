## Introduction 

This webserver (./server) and asset manager (./tools) are used to build/deploy/host 2d web based game worlds. 

Players view the world as rendered HTML/css after connecting over HTTP/WebSocket - Modest server Hardware should support hundred of players.

Check out: https://bloopworld.co For live demo


## Build 
    # Server 
        - Must have mongo-db to connect to named "bloopdb"
        - Must have "./data/areas.json" starting asset file (produced using tools)
        - Compile executable with go
    # Tools 
        - Compile executable with go
        - Can track changes in git 


## Snapshots 
This project uses: https://github.com/gkampitakis/go-snaps

To update snapshot once (Powershell) use: 
$env:UPDATE_SNAPS = 'true'; go test; Remove-Item Env:\UPDATE_SNAPS