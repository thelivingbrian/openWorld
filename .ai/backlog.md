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
  - [ ] Dynamic Tiles  
    - [ ] Add abilities for tiles or their borders to cycle between two colors 
    - [ ] Add a rainbow cycling mode for tile or its border 
    - [ ] Add a tile that looks like water
    - [ ] Tiles that can sparkle 

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

Date - Description - Notes

