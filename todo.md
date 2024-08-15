# Todo List

## Engine
- [ ] Update player view refactor
  - [-] Empty boost swaps 
- [ ] Admin screen
- [ ] Test client
- [ ] Constant special area names in tests and game 
- [-] Boosts not spawning at same rate?
- [ ] Instant kill button or key
- [-] Shift prevents stage changes even with no boosts 
- [-] No player detail update on respawn 
- [-] green dot of money is invisible after killing other player
- [-] relative border radius

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
  - [ ] Grid updates from blueprint window (Challenging because need to get/pass screenId)
    - [ ] rotate updates grid
    - [ ] oob highlight
      - [ ] fresh grid
      - [ ] Select corner 
    - [ ] Blueprint page for fragment is broken 
- [ ] Random
  - [ ] Space Topologies
    - [-] Plane
    - [-] Disconnected
    - [ ] Resize
    - [ ] Fractal 
      - [ ] Can implement in a "south zooms out" manner etc. (All 4 directions from center root square?)
  - [ ] Test Play 
    - [ ] Package executable in with tools? soft-deploy and run?
  - [ ] Save All/Everything button 
    - [ ] Cannot compile without save
  - [ ] NSEW buttons on sides of area display 
  - [ ] Clean up 
    - [ ] Remove concept of materials? 
  - [ ] relative border radius
- [ ] Space Enhancements
  - [ ] Default tile color control
  - [ ] view map
    - [-] Area -> image 
    - [-] Absolute (for plane/torus)
    - [ ] Relative? (some topologies may not project simply into a map)
- [ ] Edit Space Page
  - [-] Links to page
  - [ ] details
    - [-] details component
    - [ ] working link on page 
  - [-] generate png for simply tiled space 
  - [ ] modify blueprint
    - [-] Select by clicking on area 
    - [ ] Add Fragment/Prototype button
    - [ ] Show area (image?) 
      - [ ] Update on rotation / deletion / addition of instruction 
    - [ ] Set X and Y 
    - [ ] Highlight selected instruction 

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

## Bottom text
 - [ ] Trigger
 - [ ] Display as "!" Notification in mobile instead of on screen

## Kill streak
 - [-] User Streak
 - [ ] DB stuff 
   - [ ] Total kills
   - [ ] Total Deaths
   - [ ] Highest Streak
   - [ ] Time alive/In Danger ?

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
  

## Testing
- [ ] Unit testing 
  - [-] WebSocket
    - [-] Move Player via websocket in unit test
    - [-] Interface DB? or testing database....
      - [-] testing db works and is arguably better?
    - [ ] Test fails due to race condition 
  - [-] Most Dangerous
- [ ] Load testing
  - [ ] Selenium? 


## Transformation syntax:
layerXCss : "static {transformationType:value} string"


 
