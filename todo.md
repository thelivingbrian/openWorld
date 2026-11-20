# Todo List

## Current in progress -

### Description:
  Bring Bloopworld's overworld and play experience to a reviewable standard before releasing the new design workspace, database, and world release workflow. Build a playable, polished overworld region with a consistent three-quarter perspective, improved HUD, and fullscreen-first mobile controls. Carry the work through implementation and validation, with a reproducible preview for review.

### Agreed direction:
  - Three-quarter view on the existing square movement grid: ground seen from above, with visible fronts and tops on tall objects.
  - Preserve Bloopworld's rounded geometric shapes and simple graphics. Keep the winter-tree silhouette and bring other scenery into the same perspective.
  - Following camera by default; fixed cameras only for intentionally composed rooms and puzzles.
  - Treat world dimensions, camera viewport dimensions, and on-screen tile size as separate decisions. A larger world alone will not fix wasted screen space.
  - Make the world the primary screen content. Account for mobile browser chrome, safe areas, and orientation changes as well as true fullscreen.

### Todo:

#### 1. Establish the visual and spatial rules
  - [ ] Build a representative scene with a house and entrance, deciduous and winter trees, water, paths, an interactable, and moving characters.
  - [ ] Give tall objects consistent ground anchors, upward visual extent, scale, and contact shadows. Replace centered-trunk overhead trees with trees that follow the agreed perspective.
  - [ ] Separate each scenery object's ground anchor, collision footprint, and visual extent. Canopies and roof overhangs must not implicitly block every tile they cover.
  - [ ] Implement front/behind ordering based on ground position, with explicit handling for roofs and overhead scenery. Verify players can visibly pass in front of and behind tall objects.
  - [ ] Keep obscured players readable with an outline or selective scenery fading; keep entrances, hazards, and interaction targets legible.
  - [ ] Define a restrained palette, shadow direction, corner treatment, and scale guide that future assets can follow.

#### 2. Make camera, fullscreen, and mobile layout work together
  - [ ] Replace the narrow centered play area with a layout that uses the available viewport in portrait and landscape, including short landscape screens.
  - [ ] Add a clear user-initiated enter/exit fullscreen control where supported; keep its state correct after browser-driven exit or failure.
  - [ ] Make browser mode fully usable when fullscreen is unavailable or declined. Respond to the visible viewport as browser bars expand/collapse and respect safe-area insets.
  - [ ] Support rectangular camera views and choose visible tile counts and tile size for readability. Preserve square tiles rather than stretching the world to fit.
  - [ ] Trial a larger overworld region and choose a new authoring size default based on traversal and composition, independently of camera size.
  - [ ] Tune camera following and boundaries; verify resizing, rotation, and fullscreen transitions do not lose the player, show stale tiles, or break input.
  - [ ] Check gameplay implications of different visible areas, including mobile visibility of threats and interactions.

#### 3. Improve HUD and controls
  - [ ] Replace the large two-row heading with a compact HUD: prioritize health and useful gameplay state; reduce the prominence of the player name and secondary stats.
  - [ ] Keep HUD, menu, notifications, and fullscreen controls readable without reserving large empty bands around the world.
  - [ ] Replace oversized keyboard-labeled touch buttons with clear directional and action controls, using labels that describe the actions.
  - [ ] Position touch controls for comfortable reach in both orientations; minimize obstruction of the player and nearby gameplay while preserving usable touch targets.
  - [ ] Support simultaneous movement/action touches and reliable release/cancel behavior. Prevent stuck inputs, accidental text selection, and page gestures during play.
  - [ ] Preserve keyboard controls and make menus usable on short screens, with clear focus and dismissal behavior.

#### 4. Build the new overworld region and reusable authoring assets
  - [ ] Expand the representative scene into a coherent playable region with a recognizable arrival point, village, wooded area, water crossing, and connections to existing gameplay.
  - [ ] Compose landmarks, paths, clearings, and transitions deliberately; maintain visual interest while keeping navigation and collisions understandable.
  - [ ] Replace inconsistent scenery throughout the review region, including older overhead trees and front-only buildings.
  - [ ] Package complete reusable trees, buildings, and scenery groups with their anchors, footprints, and overlap behavior in the design workspace.
  - [ ] Keep editor previews, saved data, compiled releases, and runtime rendering consistent; preserve existing worlds when extending the content format.
  - [ ] Add restrained environmental motion and effects that complement the simple graphics and preserve gameplay readability.
  - [ ] Evaluate whether agents can author independent level sections through the same editor interface, with clear boundaries and validation before combining changes.

#### 5. Verify and deliver a reviewable implementation
  - [ ] Provide a reproducible local preview or unpublished draft/version of the new region, with a clear entry point and instructions for playing it.
  - [ ] Walk through arrival, exploration, a building entrance, an interaction, and a return route; verify connectivity and collision footprints.
  - [ ] Exercise front/behind ordering, canopy occlusion, multiple characters, camera edges, and transitions to existing areas.
  - [ ] Review desktop, mobile portrait, and mobile landscape, including the supplied 667 x 375 reference and reduced visible height from browser chrome.
  - [ ] Check both browser mode and fullscreen, browser-bar changes, safe areas, orientation changes, menus, and touch controls on actual mobile browsers where available; record any unverified cases.
  - [ ] Capture repeatable screenshots of the same locations and a short movement walkthrough demonstrating depth, camera behavior, HUD, and controls.
  - [ ] Add focused regression checks for changed rendering, camera, input, and content behavior; verify editor-to-runtime consistency and performance on a representative mobile device.
  - [ ] Record the preview location, validation results, known limitations, and remaining polish so the implementation can be reviewed before release.

#### 6. Extend the established foundation
  - [ ] Add a real-time day/night and weather cycle with an explicit world clock policy and controllable preview states in the editor.
  - [ ] Recheck player, path, entrance, and hazard readability across lighting/weather states; provide reduced-effects behavior and verify mobile performance.
  - [ ] Apply the reusable assets and review checklist to subsequent regions until the overworld consistently meets the same visual and interaction standard.


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
  - [ ] can occur with juke 
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
