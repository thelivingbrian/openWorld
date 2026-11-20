# Todo List

## Current in progress — A living multiplayer frontier

### Description
  Turn the visual review into a self-contained, playable Bloopworld slice. Preserve the pressure of life drain and competing for kill streaks, give teams reasons to meet in a larger world, and introduce a dark cooperative puzzle zone. Deliver an unpublished, reproducible review build before tuning or expanding the whole world.

### Design direction
  - Keep cute geometric bloops, rounded silhouettes, restrained palettes, and minimal faces. Scenery supports navigation and play; functional objects remain abstract and immediately recognizable. - Bloopworlds simplicity is part of its charm 
  - Separate the new world from legacy Bloopworld. Remove review portals in both directions and restore the legacy source tile. Legacy content stays available as its own world/version.
  - A large authored world does not require a large active combat area. Begin with two team refuges around a compact shared frontier; open space must have routes, resources, or meaningful choices. - I suppose this is judicious. I like the idea of an open combat world but it is difficult from a planning perspective and references are scarce. I do like for bloop death to be lightweight and fundamentally acceptable though. Tradeoffs to consider 
  - Proposed first loop: leave refuge → collect powers/boosts → fight roaming threats or rival players → use a limited recovery opportunity → decide whether to push a streak or take a puzzle expedition → die and return to team refuge. CTF is outside this slice.
  - Life drain creates expedition pressure. Camping at a refuge, reconnecting, or changing stages must not preserve a competitive streak indefinitely.
  - Keep player kills, NPC kills, team objectives, and puzzle completions distinct. No new generic currency or crafting system is needed to test this loop.

### Findings from the visual review
  - Willow Reach currently has no pickup rules: its manifest only spawns wandering NPCs. Powers and boosts need explicit placement and reliable replenishment.
  - Current spawn triggers are once/entry; staying within a large region does not trigger recurring activity. NPC capacity is tracked, but pickup capacity/ownership and respawn scheduling need explicit design.
  - Scenery is a coded procedural catalog using Floor2 tokens, not general uploaded sprites. Each drawing can extend beyond its ground anchor; native tiles/fragments still define collision. Catalog edits currently require code changes, and a four-tile visibility margin limits supported visual extent.
  - Three-quarter scenery fixtures keep a fixed viewing direction. Extend the editor with explicit anchor, footprint, and visual-bounds previews rather than implying arbitrary scenery rotation or sprite import already exists.
  - Confirm the exact live life-drain behavior before replacing/tuning it: its ticking/recovery rules are not evident in the current checkout's ordinary player damage path. Preserve the intended pressure, while documenting any live/check-out discrepancy.

### Proposed first slice and tuning hypotheses
  - Two teams, two protected arrival/refuge areas, one shared woodland/river frontier, and one small dark puzzle annex. Separate team spawns, several viable paths, and avoid a single unavoidable bridge camp. +up
  - Try a base expedition budget around 3 minutes. This is a playtest hypothesis, not an established balance value. Tune using measured traversal time and actual kill/health recovery rules.
  - Provide a visible power and boost within roughly 10–15 seconds of a normal departure. Use multiple public pickup sites, some safe and predictable, others contested. Supplies are neutral after departure. +up agree must always have simple fast obvious gameplay
  - Preserve kill-streak survival as the main competitive challenge. choose one bounded recovery rule after that audit. Recovery must extend a good run without allowing an uncontested infinite run or easy teammate/alt farming. Right now health is strictly decreasing through envionmental npc and pvp harm do we need to consider recovery
  - Use one announced, rotating supply/recovery hotspot to draw both teams toward a changing meeting point. Prototype this with an existing item/recovery mechanic before adding a capture system. Local supplies must still make the game viable when another team holds the hotspot. + sounds good in theory 
  - A neutral safe puzzle annex pauses drain only during a separate expedition: entering explicitly banks/ends the active combat streak. Return through team refuge with a fresh run; no combat-health/power farming through puzzle entry or resets. Review this proposed tradeoff in playtesting.
  - One solo-solvable introductory chamber, followed by one clearly marked two-player chamber. Any teams can cooperate inside the annex. Neither chamber blocks ordinary combat progression. - This is important. Bloopworld *is* full live multiplayer which is part of what makes it technically cool, howver bloopworld is early stage in gaining an audience and needs solo playability 

### 1. Specify the rules and preserve the existing challenge

  - [ ] Separate tutorial/refuge preparation from competitive runs. Leaving refuge starts a run; returning ends/banks it and clears combat advantages according to a documented policy. Do not allow safe-area streak camping.
      - [ ] Should there be a dedicated tutorial like there is today? or should play be ultra-simple. Playtesting of bloopworld indicated many players did not complete the tutorial 
  - [ ] Preserve existing leaderboard semantics where possible. Add separate run/puzzle/team metrics only if the slice requires them; no PvP kill credit for drain expiry, NPC deaths, friendly kills, or puzzle hazards.
  - [ ] Define multiplayer targeting/team damage, kill attribution, simultaneous-death handling, and rules against repeated cooperative/alt farming before adding recovery rewards.
]

### 2. Build a sustainable population and pickup system
  - [ ] Add server-owned recurring spawn rules independent of stage entry. Activate on occupied stages, stop on unload/shutdown, and prevent duplicate schedulers after re-entry/reconnect.
  - [ ] Author spawn groups/regions, permitted terrain, time intervals, caps, retry behavior, and minimum player distance. Never materialize threats on a player, inside a refuge, or in an inaccessible pocket.
  - [ ] Track NPCs and pickups per rule; replenish missing supplies without stacking an unlimited pile on one tile. Account for collection, consumption, despawn, lifetime expiry, and stage reset.
  - [ ] Seed a starter supply immediately, then maintain a predictable baseline while a player stays in the same region. Make powers, boosts, and recovery visually distinct from decorations and puzzle objects.
  - [ ] Add a small bounded enemy population with readable behavior and replenishment at low and higher occupancy. Maintain solo activity without filling the region with offscreen enemies.
  - [ ] Add the rotating public hotspot with clear location/timing cues. Avoid stage-wide random drops as the only way to find supplies.
  - [ ] Expose spawn regions, intervals, caps, live counts/next spawn, and rule validation in the designer and local review tooling. Keep publish/offline compilation aligned.

### 3. Compose a standalone multiplayer world
  - [ ] Create separate team refuges with readable exits and protected respawn positions. Add brief spawn protection that ends on departure/attack and cannot be used to attack safely.
  - [ ] Remove the old-world connection and add the new frontier/puzzle entry without requiring access to legacy areas. Update the manifest, review route, and generator together.
  - [ ] Rework paths around two or more approach/escape routes, pickup pockets, sightlines, and the rotating hotspot. Tune travel times against the life budget, not only tile dimensions.
  - [ ] Verify the loop with one player, two allies, opposing players, and uneven teams. Provide ordinary solo activity and keep optional multiplayer puzzles clearly marked.
  - [ ] Bound gameplay visibility/reach independently of desktop/mobile camera proportions where needed. Verify that offscreen attacks, enemy discovery, and hotspot cues remain fair on short mobile viewports.

### 4. Strengthen the art and authoring contract
  - [ ] Polish the winter grove: remove green squares under tree anchors, blend snow/grass/path borders, give pines coherent rounded tiers and snow accents, and match shadows to snow. Remove arbitrary patchwork and floating foliage.
  - [ ] Keep scenery intentionally simple: a few flat color masses, one restrained shadow direction, small highlights, no realistic textures, heavy outlines, or detailed humanoid characters. Reserve stronger contrast/glow for gameplay signals.
  - [ ] Add editor toggles for anchor, blocked footprint, interaction/entrance tile, and full visual bounds. Fixture placement previews should show the complete object before committing it.
  - [ ] Make palette selection offer complete labeled fixtures with thumbnails. Clearly distinguish placing a full cottage/fountain from placing only an anchor prototype.
  - [ ] Validate extent against supported camera margins and reject unsupported scenery tokens. Document fixed-facing fixtures and prevent footprint-only rotation through all authoring paths.
  - [ ] Preserve shared renderer parity; new scenery must be verifiable in editor, compiled runtime, local preview, and saved releases.

### 5. Build the dark puzzle annex
  - [ ] Establish a second environment in the same visual language: dark navy/charcoal void, low slate platforms/walls, restrained cyan/lilac illumination, cute bloops, clear walkable edges. Keep foreground height low enough to read puzzle state and adjacency.
  - [ ] Preserve abstract mechanics: plain pushable blocks, sockets, switches, gates, and distinct active/inactive shapes. Decorative glows must never resemble pickups or obscure hitboxes.
  - [ ] Build one solo teaching puzzle using existing push/group/airlock mechanics, then a two-player chamber with a shared observable goal and clear cooperative roles. 
  - [ ] Define participant scope, occupancy, simultaneous switch evaluation, and reset authority. Reuse existing state/reaction systems where sufficient; add only the minimal missing group trigger behavior.
  - [ ] Make puzzles recoverable after a disconnect, death, extra player, or abandoned block. Provide an obvious leave/reset flow; prevent griefing the only exit and duplicate completion rewards.
  - [ ] Isolate puzzle state per chamber/session where needed. Show required participant count before entry and allow waiting players to leave freely.
  - [ ] Keep puzzle completion recognition separate from competitive streak/health supplies. Test whether the drain-free expedition boundary feels coherent rather than splitting the game awkwardly.

### 6. Reviewable completion criteria
  - [ ] Deliver one reproducible preview/version with two team arrivals, the shared frontier, visible collectible powers/boosts, sustained enemy/supply spawning, explicit life pressure, and both puzzle chambers.
  - [ ] Demonstrate a continuous 10–15 minute session in one large area without crossing a boundary to replenish supplies or enemies. Observe at least several complete drain/death/respawn cycles.
  - [ ] Demonstrate two opposing players racing for supplies, fighting, and recovering/losing a streak; demonstrate allies solving the two-player puzzle and recovering when one disconnects.
  - [ ] Validate duplicate scheduler prevention, capacity bounds, occupied spawn exclusion, item collection races, time/drain transitions, score attribution, and puzzle reset/reward idempotency with focused tests.
  - [ ] Capture short solo, opposing-team, and cooperative walkthroughs, plus polished winter and dark-zone screenshots. Record first-pickup time, time between encounters, life-budget/travel ratio, and longest observed supply drought.
  - [ ] Check physical iOS/Android layout, browser bars, safe areas, touch comfort, and frame rate. Keep desktop/mobile sight-range balance as an explicit review item.
  - [ ] Update review docs and todo with implemented rules, tunable defaults, results, and known limitations. Keep the build unpublished until the gameplay slice has been reviewed.

### Follow-up scope
  - Day/night and weather remain deferred until the new combat/puzzle loop and visual readability are validated.
  - Broader biome expansion, additional teams/modes, general sprite import/custom scenery editing, and concurrent agent authoring through the editor follow this slice.
  - Resolve the known baseline authorization-test failure before production release; keep it distinct from this gameplay plan's regression checks.


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
