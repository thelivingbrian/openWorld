# Todo List

## Current in progress — A living frontier 

### Description
  New visual language lacks full-world cohesivenesss. Decide how to create a living world that is enjoyable at variety of player counts. 

### Open Questions
  - [ ] Is the design workspace intended to be the first class medium for world design, is it capable/functional/convienient with the new level of graphics 
    - [ ] Is admin console still fully functional
    - [ ] Does preview script duplicate functionality of start_dev.py script? Will it be removed prior to release 
    - [ ] Commit to long term world release change strategy, like that the target decouples world changes from code changes 
      - [ ] Impact on horizontal scalability 
  - [ ] Grindy exp gameplayfacet and/or economy 
    - [ ] Cut tree and other actions 
    - [ ] Type of puzzle we can proceedurally generate 
      - [ ] Try multiple candidates 
    - [ ] How does this balance with high risk of death / hard to regenerate health / lightweight death mechanic
    - [ ] Reward risk taking
  - [ ] Live day / night 
  - [ ] Different biomes 
    - [ ] Not too many but enough for a distinctive cool world 
    - [ ] Connected feel, extensible but complete 
    - [ ] Thinking creative and large scale
      - [ ] Bloop universe should feel fun/diverse/abstract enticing to explore 
      - [ ] Teleporting and visual screen hallucinations are a component ("Press G" - today)
      - [ ] The bloopworld has at least two teams - how does that influence the gameplay / design 
      - [ ] Bloopworld can seemlessly integrate muliplayer and singleplayer areas, influence on gameplay / design 
      - [ ] Think big. swarms of bloops, bloop armies, cinematic sequences, simple abstract world should be easier to push further
  - [ ] What can be worked on in parallel ?


### Previous iteration findings / review
  - Do new wandering NPCs seem less dangerous? 
  - must menu have reset/agree option always even when unusable?
  - map is not displaying correctly 

### Previous Design directions
  - Keep cute geometric bloops, rounded silhouettes, restrained palettes, and minimal faces. Scenery supports navigation and play; functional objects remain abstract and immediately recognizable. Bloopworld’s simplicity is part of its charm.
  - Separate the new world from legacy Bloopworld. Remove review portals in both directions and restore the legacy source tile. Legacy content stays available as long as not actively hindering future design
  - Keep an open combat world. Death is lightweight, expected, and followed by a quick return to play.
  - Health decreases through environmental hazards, aggressive NPCs, and PvP damage. Test the existing survival challenge before adding recovery. Evaluate gameplay consequences. 
  - Rely on simplicity, bloopworld should be simple, abstract, fun and emergent

### -01. Previous todo uncompleted

  - [ ] Verify database-backed reconnect timing under rapid refresh before release; preview sessions are intentionally ephemeral.

  - [ ] Define a repeated-victim/alt-farming policy before public competitive release.

  - [ ] Bound gameplay visibility/reach independently of desktop/mobile camera proportions where needed. Verify that offscreen attacks, enemy discovery, and hotspot cues remain fair on short mobile viewports.


### 6. Reviewable completion criteria
  - [ ] Measure encounter spacing, death-to-next-action time, and worst supply drought in human playtests across team sizes.
  - [ ] Check physical iOS/Android layout, browser bars, safe areas, touch comfort, and frame rate. Keep desktop/mobile sight-range balance as an explicit review item.
  - [ ] Confirm no "Ghost Line" bug on any devices


### Follow-up scope
  - Resolve the known baseline authorization-test failure before production release; keep it distinct from this gameplay plan's regression checks. (AI said this but I don't know what test it refers to, delete this?)
  - Highscore from menu 
  - Create a test environment? easiest way?
  - Release / v1 replacement plan 





---- AI Agents ignore below ----


# Full Backlog -

## Engine
- [ ] Do not award new most dangerous on logout? - No Ties but legitmate person may get overlooked even with continued steeak
- [ ] Smaller than 16 x 16 stage/area occupies only part of screen
- [ ] thin lines behind tiles with ground on some screens 
- [ ] Server tests are coupled to workspace output
  - [ ] Not inherently bad, indirectly validates the workspace deploy and game playability. Could be confusing

## Interactables and puzzles
- [ ] Group reaction
  - [ ] Can have group reactions occur for entire stuck group passing the group size in.
    - [ ] e.g. group member touches blackhole - all gone 
    - [ ] group of N batteries provides N charge 

## Integration 
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
- [-] Load Test NPC
  - [-] Max count ~2000 cpu ~43.4%
- [-] Load test smaller server
  - [-] 256 Players -> 32% cpu, 15MBs outbound, minimal memory
  - [-] Can invoke through github actions

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
- [ ] Can not choose new fragment size 

## Mobile
  - [-] Controls
    - [-] Vertical
    - [-] Horizontal
      - [-] Full screen? (screen resizes)
      - [ ] menu button overlap
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
  - [ ] can occur with juke 
  - [ ] Include automatically via engine? 
- [ ] Remove tutorial? too much time investment / play tests indicate people don't finish 

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
  x modularize interactions
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
 - [-] Programmable interactable state 

type []byte(update) 100 times

[]byte(update) 
I will type []byte(update) 45 more times
I will type []byte(update) 44 more times
