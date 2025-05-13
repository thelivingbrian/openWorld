# Todo List

## Engine
- [ ] Update player view refactor
  - [-] Empty boost swaps 
- [ ] Admin screen
 - [-] Player / Team count
 - [ ] Stage list 
 - [ ] Most Dangerous stats 
 - [ ] Observe stage / player 
- [-] Test client
  - [-] integration client
- [ ] Constant special area names in tests and game
  - [ ] complete? e.g. clinic & test stages? 
- [ ] Constant predefined interactables 
- [-] Boosts not spawning at same rate?
- [-] Instant kill button or key
- [-] Shift prevents stage changes even with no boosts 
- [-] No player detail update on respawn 
- [-] green dot of money is invisible after killing other player
- [-] relative border radius
- [-] Score goal
- [-] Minimum streak for most dangerous (Possibly just for award but possibly for inclusion in heap as well)
  - [ ] Do not award new most dangerous on logout? - No Ties but legitmate person may get overlooked even with continued steeak
- [-] Player zombies
  - [-] up to 1,000 concurrent player deaths
  - [-] Automated test
- [-] Disable respawn from tutorial
- [-] Player still on stage with closed channel
- [-] Improved goal scoring
- [-] Improved bottom text 
  - [-] Tutorial text no longer viewable 
  - [-] color code most dangerous 
  - [-] Notify goal scores and team in lead 
  - [-] Goal shows score of each team
- [-] Seperate homepage from game server
- [-] Add sound fx
- [ ] Canvas based interactive/realtime stage map? 

## Integration 
- [ ] Bot AI
  - [ ] Use boosts
  - [ ] Move in line
  - [ ] Open menus
  - [ ] Hallucinate
- [-] All players in tutorial 
- [-] With DB Writes 

## Stats 
- [ ] Boosts 
  - [ ] used
  - [x] collected - nah?
- [ ] Money
  - [ ] Total
  - [-] Current
  - [-] Peak 
- [-] Goals scored
- [ ] Games won
- [-] NPC Kills
- [-] Peak Killstreak 

## Highscores
- [-] Current Money / Kills / Goals Scored
- [-] Peak kill streak
  - [-] player/npc kills as side stats - tried but no, too busy
  - [-] KD (Player + NPC) / Deaths as side stat 
- [-] Peak money 
  - [-] Current money becomes side stat
  - [-] Total money as side stat?
  - [-] Too busy - probably 2 stats tops - prefer 1. 
- [ ] Games won as side stat for Goals Scored
- [ ] BUG: Non-number in mongo breaks HS list for everyone 

## Accomplishments 
- [-] Add accomplishments list

## Performance 
- [-] Performance from Ec2 is degraded vs localhost - 400 Websocket users / 8 stages very stable
  - [-] Client overwhelmed Potentially?
    - [-] Relation to integration test with no reader?
      - [-] Run PProf on Ec2  
      - [-] Run PProf locally with no reader
      - [-] Lack of reader can be solved by timeout on socket send
        - [-] Logout is unclean, can leave channels closed before writes ? 
        - [-] Performance still degrades significantly - Only with mass timeout ? 
    - [-] Websocketstream? 
      - [-] No - only available on chrome
- [-] Highlight code not working with buffered sends
  - [-] Initial screen load no longer requires button press? 
- [-] Highlight code broken for overlapping after teleport
- [-] Result of removing buffer with the new timeout deadlines (Should still get overwhelmed)
  - [-] Not as good
- [-] Load test database cluster
- [ ] Load Test NPC
  - [-] Max count ~2000 cpu ~43.4%

## Design Workspace
- [-] Rotations
  - [-] New collection
  - [-] Convert materials to protos
    - [-] Create default protos
    - [-] Unique id 
    - [-] Edit Proto
      - [-] Do not transform props (4 recv funcs?)
    - [-] New Prototype set
    - [-] define rotate(proto)
  - [-] Update fragment schema to have transformations 
    - [-] New Fragment Set
    - [-] New Fragment
    - [-] Fragment has protos
    - [-] Fragment applies transformations
  - [-] Modify Transformations
    - [-] Fragment Transform Proto
    - [-] Area transform proto? 
  - [-] Blueprint page 
    - [-] Place fragment or proto on blueprint
    - [-] Do not place empty cells
    - [-] Transform fragment 
    - [-] remove/reorder instructions
    - [-] modify instructions 
      - [-] blueprint rotate removes previous 
  - [-] Compile Collection 
    - [-] DefaultFragements([][]proto) -> areas + materials
  - [-] Cleanup 
    - [-] Restore tests
    - [-] Default areas and protos
- [ ] Blueprint enhancements
  - [ ] Grid updates from blueprint window
    - [ ] Area updates with instruction  
      - [ ] Update on rotation / deletion / addition of instruction 
      - [ ] reload area-edit
      - [-] modular area-edit 
        - [-] Follow current page style with reload (blueprint etc) 
          - [-] Tiny nav only
    - [-] oob highlight
      - [-] Select corner 
  - [-] Blueprint page for fragment is broken 
      - [-] Fragment can only view the modify window and blueprint is loading for the parent area
  - [ ] Instruction human readable name
- [ ] Space Enhancements
  - [-] Default tile color control
  - [-] view map
    - [-] Area -> image 
    - [-] Absolute (for plane/torus)
  - [ ] Matrix for space 
    - [ ] Apply prototype via matrix 
- [ ] Random
  - [ ] Interactable "select" tool  does not work on main grid
    = [ ] is umimpemented in general - could indicated selected ?
  - [ ] Space Topologies
    - [-] Plane
    - [-] Disconnected
    - [ ] Resize
    - [ ] Fractal 
      - [ ] Can implement in a "south zooms out" manner etc. (All 4 directions from center root square?)
    - [ ] Cube and/or higher torus
    - [ ] Maps for non-simple tilings?
      - [ ] Relative to current area
  - [ ] Test Play 
    - [ ] Package executable in with tools? soft-deploy and run
    - [ ] level player (e.g. live stage demo) ^ same as above
  - [ ] Save All/Everything button 
    - [ ] Cannot compile without save
    - [-] Save space 
  - [-] NSEW buttons on sides of area display 
  - [-] Clean up 
    - [-] Remove concept of materials 
  - [-] relative border radius
- [-] Edit Space Page
  - [-] Links to page
  - [-] details
    - [-] details component
    - [-] alternate links on page 
  - [-] generate png for simply tiled space 
  - [-] modify blueprint
    - [-] Select by clicking on area 
    - [-] Set X and Y 
    - [-] Set rotation
    - [-] Highlight selected instruction
- [-] generate structure
  - [-] Generate new 
  - [-] Place at coord 
  - [-] Remove structure 
    - [-] at coord 
  - [-] Regenerate 
  - [-] Delete
- [-] version control
  - [-] workspace deploys via build command
    - [-] clean up collection logic
    - [-] add cli  
  - [-] get rid of proc folder?
- [ ] Bugs: 
  - [ ] New color will output to local file but deploying requires application restart.
  - [ ] New areas are always "unsafe"


## Mobile
  - [-] Controls
    - [-] Cleanup current branch
      - [-] Add missing test and square stages
        - [-] Square 
          - [-] 4x4 with center river (Looks and plays bad should offset) 
          - [-] 5x5 with river
        - [-] Test
    - [-] Mobile controls
      - [-] Detect Touch Screen
        - [-] detect landscape/protrait 
        - [-] show buttons only on mobile
      - [-] Display controls
      - [-] Send events on tap 
      - [-] Test on device
      - [-] menus
        - [-] touch input vs adding controls for mobile? (Try touch first)
        - [-] touch may be better for menu 
      - [-] shift 
      - [-] desktop cenetering
        - [-] info div
      - [-] Resize grid squares
  - [-] fix display of map

## Display 
- [-] Highlights
  - [-] Same stage teleport highlights
    - [-] Overlap excluded
    - [-] Sometimes entire highlight is removed at once shortly after displaying ?
  - [ ] Disable buttons e.g. on laptop with touch screen 

## Bottom text
 - [-] Trigger
 - [ ] Display as "!" Notification in mobile instead of on screen
 - [ ] Fade with time
 - [ ] Deault bottom text on load
 - [ ] Game tips

## Kill streak
 - [-] User Streak
 - [ ] DB stuff 
   - [-] Total kills
   - [-] Total Deaths
   - [ ] Highest Streak
   - [ ] Time alive/In Danger ?
- [ ] Test conconcurrency
- [-] Add trim reward

## Metrics

## World map
- [-] Basic world maps
  - [-] Add map color to prototype
    - [-] generate map colors automatically
  - [-] Area to map (png?)
    - [-] png to svg so map can be resized? - No resize with css
  - [-] Serve an image
  - [-] Space data model 
    - [-] topology
    - [-] map
    - [-] map for subarea (do this with highlighting below)
  - [-] save space as json 
    - [-] compile space changes (if any) 
  - [-] Map for spaces (Torus and Plane only?)
    - [-] Map for fixed size plane/torus
  - [-] "You are here" highlighting? 
    - [-] 1 map per area generated (May be easy to extend this to other topologies like cube also)
  - [-] Serve maps from game
    - [-] Deploy 
    - [-] Map window opens 
    - [-] Serve image by uuid
    - [-] Scvale image inside map window 
- [-] Additions
  - [-] Prototype edit map color (And generate automatically?)
  - [-] Map has wrong size on wide monitor 
- [-] Infirmary
- [ ] Show weather on map 
  

## Testing
- [-] Unit testing 
  - [-] WebSocket
    - [-] Move Player via websocket in unit test
    - [-] Interface DB? or testing database....
      - [-] testing db works and is arguably better?
    - [-] Test fails due to race condition 
  - [-] Most Dangerous
  - [-] Precomputed seed disagrees with current result (on linux)
- [-] Load testing
  - [ ] Selenium? 
- [-] Benchmarks 
  - [-] Benchmark slowness caused by test: TestDamageABunchOfPlayers. MoveAllTwice went from ~17ms to ~30ms
    - [-] close routines
    - [-] introduced via commit 90a3043177f78f90fb651c2cc1e427031c888e33

## Tutorial
- [ ] New Tutorial
  - [ ] Layout
  - [ ] Item Spawning
  - [ ] NPC Spawn
  - [ ] Teleport Home
- [ ] Skip tutorial option

## Bugs
 - [-] test remove damage tangibility check
 - [-] prevent infitine interactable spawn bug (technically still possible for game balls)

## Transformation syntax:
layerXCss : "static {transformationType:value} string"




Interactable machine for teleporting across a boundary:

 [ ] <-  push nil after to reset other side 
 ---------
 [ ]
Enter here

One way or two way.


-goals:

 - [ ] Broadcast group for area descriptions
  - [-] Capture
  - [ ] Utilize 
 - [ ] Weather type for area descriptions
   - [-] Static
   - [ ] Dynamic
 - [ ] Ground pattern for area description
  - [-] Grid for editing ground pattern
    - [-] toggle 
    - [-] toggle between / toggle fill 
    - [-] view ground from area edit
    - [-] return to area edit not working
  - [-] Ground is visible from map
  - [-] Blueprint color1 and color2 
    - [-] When creating space
    - [-] Update for area
    - [-] update for space 
  - [ ] Add ground pattern using structure window 
  - [ ] Add additional states for Cell ? 
    - [ ] would need to extend smoothness algorithm 
 - [-] remove global variables
 - [-] update Area output to have materials by value
   - [-] remove material output?  
   - [-] compile / load successfully w/ ground
   - [-] compile tests
     - [-] Snapshots for grid actions 
   - [ ] http writer - keep in handler instead of action funcs 

 - [-] Toroidal Woods - 12x12

 - [ ] Clean up, consolidate todo list? 

 - [ ] Spawn NPCs 
   - [-] Basic
   - [ ] Clean up old spawns (e.g. method signature changes)
 - [ ] Programmable interactable state 

type []byte(update) 100 times

[]byte(update)
I will type []byte(update) 70 more times
I will type []byte(update) 69 more times 
I will type []byte(update) 68 more times

