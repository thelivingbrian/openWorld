# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: Admin console
  Idea: There needs to be a way for bloopworld admins to see what is going on with the live server and databases, currently the only insight is via mongodb compass / console logs / actually signing into the game. The admin console will improve and consolidate all of these processes 

  Tasks:
- [ ] Admin Panel
  - [x] View player info 
    - [x] modify stats
    - [x] Ban Player 
  - [ ] World console 
    - [x] Actively logged in players 
      - [x] links to player info
    - [x] Active stages 
    - [x] Can observe by stage or player. E.g. see the same game screen / canvas
        - [x] Player watch mirrors the player's camera and includes highlights
        - [x] Player watch has independent HUD and menu visibility options, both off by default
        - [x] Player logout leaves a dark canvas with a logged-out message
        - [x] Stage watch starts at a random valid viewport and rotates to another viewport after 30 seconds without visible activity
        - [x] Stage watch has a Next camera control that jumps to another random viewport on the same stage; free panning is deferred
        - [ ] Should be extendable into a non-admin observation deck for the site 
    - [x] View session information 
    - [x] Improve stylistic appearance
      - [x] Player info appears in modal popup
        - [x] Player info has "watch" link for live preview
      - [x] remove actions -> view column
        - [x] player name is link to player info
        - [x] stage name is link to stage info modal which has live view link and live info
      - [x] Produce quick write up on performance impact of polling for updates on console screens instead of manual refresh



Notes:
- 2026-07-02 live-preview implementation:
  - Added admin-authenticated `/admin/watch` and `/admin/watch/screen` routes for read-only player and stage canvas streaming.
  - Player watches mirror canvas updates and highlights from the watched player's outbound stream. HUD and menu forwarding are independent query options and default off.
  - Stage watches use non-blocking read-only cameras, random viewport selection, 30-second visible-inactivity rotation, and a manual Next camera action.
  - Slow watcher queues request a full viewport resync instead of blocking gameplay camera-zone updates.
  - Observer-role access remains deferred; watch routes currently require admin access.
- 2026-07-02 live-preview planning decisions:
  - Player preview includes the target player's highlight layer. HUD and menu visibility are separate preview settings and both default off.
  - A player logout keeps the preview open but replaces the canvas with a dark state and an explanatory message.
  - Stage preview chooses a random valid viewport initially. After 30 seconds with no updates visible in that viewport, it chooses another random viewport on the same stage.
  - Stage preview has a manual Next camera action using the same random-view selection. Arbitrary panning is deferred.
- 2026-07-02 UI update:
  - Restyled the admin console as responsive panels and tables.
  - Player names now open player details/editing in a modal; the redundant actions column was removed.
  - Stage names now open a live-info modal. Player/stage watch affordances were subsequently connected to the live-preview routes.
- 2026-03-28 implementation update:
  - Added server-rendered admin console routes on the game server (`/admin`, `/admin/player/update`, `/admin/player/ban`).
  - Added admin access gating via `ADMIN_IDENTIFIERS` env var (comma-separated provider identifiers).
  - Added player stat/location/team/accomplishment editing workflow with live-online player update + persistence.
  - Added ban flow with immediate kick and optional duration in minutes (blank/0 = permanent).
  - Added login-time ban enforcement for authorized users.
  - Added admin audit trail persistence in new Mongo collection `adminActions`.
  - Observation canvas parity is still pending and remains unchecked.

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
      - [ ] Admins can mark a world as "persistent" meaning it will stay live after the creator signs out
    - [ ] world "watch" mode
      - [ ] some method of picking the most interesting player / stage - shows them
      - [ ] more efficient for public launch / high observer count? 

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

