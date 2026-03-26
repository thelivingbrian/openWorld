# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: Cooler visuals 
  Idea: Now we have switched from HTML based to Canvas based graphics lets leverage this and add some more advanced visual effects 

  Tasks:
  - [x] Dynamic weather
    - [x] Currently stages have a weather string that applies that solid color to the "w" layer of canvas. e.g "blue trsp20" to create an overcast rainy effect
    - [x] Let's extend that to allow for specific additional weather modes 
    - [x] As a proof of concept lets add a "raining" weather which will use the w layered canvas but besides an overcast effect additionally add falling rain drops 
    - [x] Notes: weather strings now accept mode tokens (example: "blue trsp20 raining"). Unknown tokens are ignored, so old weather values still render the same.
  - [x] Dynamic Tiles  
    - [x] Add abilities for tiles or their borders to cycle between two colors 
    - [x] Add a rainbow cycling mode for tile or its border 
    - [x] Add a tile that looks like water
    - [x] Tiles that can sparkle 
    - [x] Notes: supported tokens are `cycle(colorA,colorB)`, `cycle-b(colorA,colorB)`, `rainbow`, `rainbow-b`, `water`, and `sparkle`.

## Backlog
- [ ] Design workspace dynamic tile previews
  - [ ] Add shared dynamic-style token parser in SPA editor for `cycle(...)`, `cycle-b(...)`, `rainbow`, `rainbow-b`, `water`, and `sparkle`
  - [ ] Add preview animation loop (`requestAnimationFrame`) in editor state to drive time-based style updates
  - [ ] Wire computed dynamic styles into main grid tile rendering (all relevant visual layers)
  - [ ] Wire computed dynamic styles into fixture preview rendering so selected assets preview animated behavior
  - [ ] Ensure deterministic per-cell phase offsets (based on tile coordinates) so previews look stable while panning/editing
  - [ ] Add fallback behavior for unknown/invalid tokens to keep editor rendering safe and non-breaking
  - [ ] Add/update SPA tests for parser behavior and dynamic preview rendering paths
  - [ ] Document editor preview support and any differences vs in-game runtime rendering
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

Date - Description - Notes

