# Todo List

## Engine
- [ ] Update player view refactor
  - [ ] Empty boost swaps 
- [ ] Admin screen
- [ ] Test client
- [ ] Constant special area names in tests and game 
- [ ] Boosta not spawning?

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
- [ ] Space Edit
  - [ ] Default tile color control
  - [ ] view map
    - [ ] Area -> image 
    - [ ] Absolute (for plane/torus)
    - [ ] Relative? (some topologies may not project simply into a map)

## Mobile Controls
  - [-] Cleanup current branch
    - [-] Add missing test and square stages
      - [-] Square 
        - [-] 4x4 with center river (Looks and plays bad should offset) 
        - [-] 5x5 with river
      - [-] Test
  - [ ] Mobile controls
    - [ ] Detect Touch Screen
    - [ ] Display controls
    - [ ] Send events on tap 

## World map
- [-] Add map color to prototype
 - [-] generate map colors automatically
- [ ] Area to map (png?)
 - [ ] png to svg so map can be resized? 
- [ ] Space data model 
 - [ ] topology
 - [ ] map
 - [ ] map for subarea 
 - [ ] save space as json 
  - [ ] compile space changes (if any) 
- [ ] Map for spaces (Torus and Plane only?)

  


### Transformation syntax:
layerXCss : "static {transformationType:value} string"


 
