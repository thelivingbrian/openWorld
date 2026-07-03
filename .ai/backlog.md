# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: Operationalize design workspace
  Idea: Present design workspace is used to produce a compiled artifact using the repository data. The disadvantage is that it makes doing level editing a heavy local process. Additionally it could be nice to eventually support admin or user created bloopworlds.

  There is inherent complexity with this shift though. For one thing the function of the highscores and the behavior of tutorials would need to be made more generic. The structure of "players" within the database would also need to change in some way to support multiple worlds with different teams/rules/locations. How to manage deployments and resources (compute / urls). What implications there are to performance and security. Risk of malicious db usage, how many worlds can one server reliably support, potential future load balancing. I would like enough assets to exist locally that a cloned copy of the repo would still function.

  Plan:
  - [x] Add Mongo-backed, revisioned world drafts; deterministic immutable release archives; GridFS storage; rollback; quotas; and per-world palettes/base maps.
  - [x] Add controller/runtime modes, path routing, supervised world processes, artifact caching, persistent/owner-present/until-empty lifecycle, and graceful shutdown.
  - [x] Scope player profiles, events, sessions, teams, onboarding metadata, and highscores by world while migrating legacy records.
  - [x] Host the authenticated design workspace, seed player worlds from repository content, expose public directory/lifecycle/moderation APIs, and package it for deployment.

## Backlog
- [ ] Player created worlds
    - [x] From world select screen give option to edit or launch world
    - [x] One-n Collection(s) per player stored in mongo
        - [x] Enforce space and running-world limits
        - [x] Retain immutable release versions and rollback
    - [x] "Edit" option opens user's collection in the design workspace
    - [x] "Launch" starts an isolated world with owner-present or until-empty lifecycle
      - [x] Admins can mark a world as "persistent" meaning it will stay live after the creator signs out
    - [ ] world "watch" mode
      - [ ] some method of picking the most interesting player / stage - shows them
      - [ ] more efficient for public launch / high observer count? 
- [ ] Workflow deploy enhancements
  - [ ] Install the repository's `deploy-openworld.sh` into `/opt/openworld/bin` during deployment so the installed copy cannot become stale.
  - [ ] Add a manually triggered configuration-only workflow that writes production configuration atomically and restarts the service without deploying a new binary.
    - [ ] Either need variables / secrets in GH or can use one multiline secret "env file" 
  - [ ] Make deployment fail when the systemd service is missing or fails to restart, and remove status-output pipelines that can fail under `pipefail`.
  - [ ] Verify application health after restart and preserve or restore the previous configuration when validation fails.

## Blocked / Questions

