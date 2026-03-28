# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: 
  Idea: 

  Tasks:


## Backlog
- [ ] Admin Panel
  - [ ] View player info 
    - [ ] modify stats
    - [ ] Ban Player 
  - [ ] World console 
    - [ ] Actively logged in players 
      - [ ] links to player info
    - [ ] Active stages 
    - [ ] Can observe by stage or player. E.g. see the same game screen / canvas
        - [ ] Should be extendable into a non-admin observation deck for the site 
    - [ ] View session information 
- [ ] Player created worlds
    - [ ] From world select screen give option to edit or launch world
    - [ ] One Collection per player stored in mongo
        - [ ] need space limit 
    - [ ] "Edit" option opens user's collection in the design workspace
    - [ ] "Launch" will start the collection as a new world that will continue to run until the owning player signs out

## Blocked / Questions

## Done 
- [x] Visual animations
  - [x] Dynamic weather
    - [x] Notes: weather strings now accept mode tokens (example: "blue trsp20 raining"). Unknown tokens are ignored, so old weather values still render the same.
  - [x] Dynamic Tiles  
    - [x] Notes: supported tokens are `cycle(colorA,colorB)`, `cycle-b(colorA,colorB)`, `rainbow`, `rainbow-b`, `water`, and `sparkle`.
- [x] Design workspace dynamic tile previews
  - [x] Add shared dynamic-style token parser in SPA editor for `cycle(...)`, `cycle-b(...)`, `rainbow`, `rainbow-b`, `water`, and `sparkle`
  - [x] Add preview animation loop (`requestAnimationFrame`) in editor state to drive time-based style updates
  - [x] Wire computed dynamic styles into main grid tile rendering (all relevant visual layers)
  - [x] Wire computed dynamic styles into fixture preview rendering so selected assets preview animated behavior
  - [x] Ensure deterministic per-cell phase offsets (based on tile coordinates) so previews look stable while panning/editing
  - [x] Add fallback behavior for unknown/invalid tokens to keep editor rendering safe and non-breaking
  - [x] Add/update SPA tests for parser behavior and dynamic preview rendering paths
  - [x] Document editor preview support and any differences vs in-game runtime rendering
    - [x] Notes: `sparkle` uses CSS brightness shimmer (canvas glint dots not available in CSS). All other tokens produce identical colour values to the runtime. Unknown tokens are ignored.

