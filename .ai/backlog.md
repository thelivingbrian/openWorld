# AI Backlog

Shared queue for human + agent execution.

Notes: AI agents see .ai/AGENTS.md for high level overview of project and working process

## Current In-Progress 
  Feature: 
  Idea: 

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
- [ ] Workflow deploy enhancements
  - [ ] Install the repository's `deploy-openworld.sh` into `/opt/openworld/bin` during deployment so the installed copy cannot become stale.
  - [ ] Add a manually triggered configuration-only workflow that writes production configuration atomically and restarts the service without deploying a new binary.
    - [ ] Either need variables / secrets in GH or can use one multiline secret "env file" 
  - [ ] Make deployment fail when the systemd service is missing or fails to restart, and remove status-output pipelines that can fail under `pipefail`.
  - [ ] Verify application health after restart and preserve or restore the previous configuration when validation fails.

## Blocked / Questions

