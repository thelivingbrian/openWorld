# Todo List

## Engine
- [ ] Admin screen
 - [-] Player / Team count
 - [ ] Stage list 
 - [ ] Most Dangerous stats 
 - [ ] Observe stage / player 
- [-] Minimum streak for most dangerous (Possibly just for award but possibly for inclusion in heap as well)
  - [ ] Do not award new most dangerous on logout? - No Ties but legitmate person may get overlooked even with continued steeak
- [ ] Canvas based interactive/realtime stage map? 

## Interactables and puzzles
- [ ] mutable 'state' property for interactables
  - [ ] can be set in design workspace 
  - [ ] can be used as 'reactsWith' gate (e.g. state is/is not/contains)
- [ ] transmit push movement
  - [ ] reaction with nil
  - [ ] sends push to other interactables by state or type
  - [ ] allow for rotation or scale of push vector
- [ ] sticky blocks 
  - [ ] will stick together
  - [ ] requires polyomino based pushing logic
    - [ ] pushable interactables? 
- [ ] new puzzle
  - [ ] new space in collection:escape demonstrating the new functionality 

## Integration 
- [ ] Python rewrite
- [ ] Bot AI
  - [ ] Use boosts
  - [ ] Move in line
  - [ ] Open menus
  - [ ] Hallucinate

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
- [ ] BUG: Non-number in mongo breaks HS list for everyone 

## Accomplishments 
- [-] Add accomplishments list

## Performance 
- [-] Load test database cluster
- [ ] Load Test NPC
  - [-] Max count ~2000 cpu ~43.4%
- [ ] Load test smaller server

## Design Workspace
- [ ] place NPCs
- [ ] diff based save
- [ ] Matrix for space 
  - [ ] Apply prototype via matrix 
- [ ] Random
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
      - [ ] is this still true? test
- [ ] Editor color for interactable


## Workspace Bugs 
- [ ] New areas are always "unsafe"


## Mobile
  - [-] Controls
    - [-] Vertical
    - [-] Horizontal
      - [ ] Full screen? (screen resizes)
      - [ ] menu button overlap
  - [ ] Pause continued support? 
  - [ ] clicking triggers highlighting

## Bottom text
 - [-] Trigger
 - [ ] Display as "!" Notification in mobile instead of on screen
 - [-] Fade with time
 - [ ] Deault bottom text on load
 - [ ] Game tips

## Metrics
  

## Testing


## Tutorial
- [ ] Prevent interactable overlaps teleport 
  - [-] Can prevent manually with "pass-all" interactable
  - [-] is applied in tutorial 
  - [ ] Apply to rest of world
  - [ ] Include automatically via engine? 

## Bugs

## Transformation syntax:
layerXCss : "static {transformationType:value} string"


If item spawns on stage on same initial tile as you - do you pick it up? - Think maybe no.


Events/Triggers - Similar to ineractables but stationary - stackable - have state - only react to nil  


Interactable machine for teleporting across a boundary:

 [ ] <-  push nil after to reset other side 
 ---------
 [ ]
Enter here

One way or two way.

-goals (new): 
  - add tests for camera
    - Count number of updates? 
  x choose license
    x AGPLv3 
  - modularize interactions
  x better workspace css 
  - Highscore from menu
  x Horizontal mobile

-goals:

 - [ ] Broadcast group for area descriptions
  - [-] Capture
  - [ ] Utilize 
 - [ ] Weather type for area descriptions
   - [-] Static
   - [ ] Dynamic

 - [-] Clean up, consolidate todo list? 

 - [ ] Spawn NPCs 
   - [-] Basic
   - [ ] Clean up old spawns (e.g. method signature changes)
   - [ ] from design workspace
 - [ ] Programmable interactable state 

type []byte(update) 100 times

[]byte(update) 
I will type []byte(update) 56 more times 
I will type []byte(update) 55 more times
