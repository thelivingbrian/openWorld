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
- [ ] Player still on stage with closed channel
- [ ] Improved goal scoring
- [ ] Improved bottom text 
  - [ ] Tutorial text no longer viewable 
  - [ ] color code most dangerous 
  - [ ] Notify goal scores and team in lead 

## Integration 
- [ ] Bot AI
  - [ ] Use boosts
  - [ ] Move in line
  - [ ] Open menus
  - [ ] Hallucinate
- [-] All players in tutorial 
- [ ] With DB Writes 

## Stats 
- [ ] Boosts 
  - [ ] used
  - [ ] collected
- [ ] Total money 
- [ ] Goals scored 

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
  - [ ] Blueprint page for fragment is broken 
      - [ ] Fragment can only view the modify window and blueprint is loading for the parent area
  - [ ] Instruction human readable name
- [ ] Space Enhancements
  - [ ] Default tile color control
  - [-] view map
    - [-] Area -> image 
    - [-] Absolute (for plane/torus)
  - [ ] Matrix for space 
    - [ ] Apply prototype via matrix 
- [ ] Random
  - [ ] Interactable "select" tool  does not work on main grid
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
    - [ ] Package executable in with tools? soft-deploy and run?
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
- [ ] level player? 


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
- [ ] Highlights
  - [ ] Same stage teleport highlights
    - [ ] Overlap excluded
    - [ ] Sometimes entire highlight is removed at once shortly after displaying ?

## Bottom text
 - [-] Trigger
 - [ ] Display as "!" Notification in mobile instead of on screen

## Kill streak
 - [-] User Streak
 - [ ] DB stuff 
   - [-] Total kills
   - [-] Total Deaths
   - [ ] Highest Streak
   - [ ] Time alive/In Danger ?
- [ ] Test conconcurrency
- [ ] Add trim reward

## Stats / Metrics

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


## Transformation syntax:
layerXCss : "static {transformationType:value} string"



-goals:

x - Expose endpoint that returns world player count and player count by team 

x - Tidy Up sandy world 

improve goal scoring experience 

type []byte(update) 100 times

 
