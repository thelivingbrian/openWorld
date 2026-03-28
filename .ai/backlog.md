# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: Admin console
  Idea: There needs to be a way for bloopworld admins to see what is going on with the live server and databases, currently the only insight is via mongodb compass / console logs / actually signing into the game. The admin console will improve and consolidate all of these processes 

  Tasks:
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

Notes:
- Planning session decisions (2026-03-28):
  - Roles for MVP: `admin` and `observer`.
  - Observer access: no observers for now (observer-facing features deferred).
  - Admin console UI location: server-rendered page in the game server.
  - Ban behavior: include immediate kick + optional soft ban duration chosen at action time.
  - Stat editing MVP fields: money, health, accomplishments, team, and location.
  - Admin audit trail: required; add a new Mongo collection for admin action records.
    - Capture at minimum: action type, acting admin identifier, target player identifier (if any), payload/delta, timestamp, and result/status.
  - Session info scope (MVP): current live in-memory session information only.
  - Historical session views: deferred to a later phase.
- Discovery notes from planning:
  - Current codebase has no existing role/permission enforcement layer yet.
  - No existing admin routes/pages currently exist.
  - Existing telemetry endpoints are limited; admin-specific APIs/views will need to be added.
  - live spectator view canvas is available only in the admin screen for now
  - view can piggyback a logged in player or lock on a given stage 

## Backlog
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
    - [x] Notes: `sparkle` uses CSS brightness shimmer (canvas glint dots not available in CSS). All other tokens produce identical colour values to the runtime. Unknown tokens are ignored.
  - [x] Add preview animation loop (`requestAnimationFrame`) in editor state to drive time-based style updates

