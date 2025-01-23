package main

import (
	"fmt"
	"math/rand"
)

type Interactable struct {
	name           string
	pushable       bool
	walkable       bool
	cssClass       string
	fragile        bool
	reactions      []InteractableReaction // Lowest index match wins
	rejectTeleport bool
}

type InteractableReaction struct {
	ReactsWith func(incoming *Interactable, initiatior *Player) bool
	Reaction   func(incoming *Interactable, initiatior *Player, location *Tile) (outgoing *Interactable, push bool) // rotate ?
}

var interactableReactions = map[string][]InteractableReaction{
	"black-hole": []InteractableReaction{InteractableReaction{ReactsWith: everything, Reaction: eat}},
	"fuchsia-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: interactableHasName("fuchsia-ball"), Reaction: scoreGoalForTeam("sky-blue")},
		InteractableReaction{ReactsWith: interactableHasName("sky-blue-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
	"sky-blue-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: interactableHasName("sky-blue-ball"), Reaction: scoreGoalForTeam("fuchsia")},
		InteractableReaction{ReactsWith: interactableHasName("fuchsia-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
	"tutorial-goal-sky-blue": []InteractableReaction{
		InteractableReaction{ReactsWith: playerTeamAndBallNameMatch("sky-blue"), Reaction: destroyEveryotherInteractable},
		InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("sky-blue"), Reaction: notify("Try using the matching ball.")},
	},
	"tutorial-goal-fuchsia": []InteractableReaction{
		InteractableReaction{ReactsWith: playerTeamAndBallNameMatch("fuchsia"), Reaction: destroyEveryotherInteractable},
		InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("fuchsia"), Reaction: notify("Try using the matching ball.")},
	},
	"gold-target": []InteractableReaction{
		InteractableReaction{ReactsWith: interactableHasName("ball-gold"), Reaction: destroyEveryotherInteractable},
	},
}

func (source *Interactable) React(incoming *Interactable, initiator *Player, location *Tile, yOff, xOff int) bool {
	if source.reactions == nil {
		return false
	}
	for i := range source.reactions {
		if source.reactions[i].ReactsWith != nil && source.reactions[i].ReactsWith(incoming, initiator) {
			outgoing, push := source.reactions[i].Reaction(incoming, initiator, location)
			if push {
				nextTile := getRelativeTile(location, yOff, xOff, initiator)
				if nextTile == nil {
					return false
				}
				return initiator.push(nextTile, outgoing, yOff, xOff)
			}
			return true
		}
	}
	return false
}

// Gates
func everything(*Interactable, *Player) bool {
	return true
}

func matchesCssClass(cssClass string) func(*Interactable) bool {
	return func(i *Interactable) bool {
		return i.cssClass == cssClass
	}
}

func interactableHasName(name string) func(*Interactable, *Player) bool {
	return func(i *Interactable, _ *Player) bool {
		if i == nil {
			return false
		}
		return i.name == name
	}
}
func playerTeamAndBallNameMatch(team string) func(*Interactable, *Player) bool {
	return func(i *Interactable, p *Player) bool {
		if i == nil {
			return false
		}
		fmt.Println("HEYO", i.name, team, p.getTeamNameSync())
		return i.name == "ball-"+team && team == p.getTeamNameSync()
	}
}

func PlayerAndTeamMatchButDifferentBall(team string) func(*Interactable, *Player) bool {
	return func(i *Interactable, p *Player) bool {
		if i == nil {
			return false
		}
		if len(i.name) < 5 {
			return false
		}
		return i.name[0:5] == "ball-" && team == p.getTeamNameSync()
	}
}

// Actions
func eat(*Interactable, *Player, *Tile) (*Interactable, bool) {
	// incoming interactable is discarded
	return nil, false
}

func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) (outgoing *Interactable, ok bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		p.world.leaderBoard.scoreboard.Increment(team)
		if team == p.getTeamNameSync() {
			// Otherwise you have scored a goal for a different team
			p.incrementGoalsScored()
			p.updateRecord()
		}
		fmt.Println(p.world.leaderBoard.scoreboard.GetScore(team))
		return nil, false
	}
}

func scoreGoalForPlayer(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	team := p.getTeamNameSync()
	p.world.leaderBoard.scoreboard.Increment(team)
	p.incrementGoalsScored()
	p.updateRecord()
	// need to  give a trim
	fmt.Println(p.world.leaderBoard.scoreboard.GetScore(team))
	return nil, false
}

func spawnMoney(amounts []int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		tiles := walkableTiles(t.stage.tiles)
		count := len(tiles)
		if count == 0 {
			return nil, false
		}
		for i := range amounts {
			randn := rand.Intn(count)
			tiles[randn].addMoneyAndNotifyAll(amounts[i])
		}
		return nil, false
	}
}

func finishTutorial2(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	t.stage.tiles[7][7].teleport.destStage = getStageFromTeam(p.getTeamNameSync())
	t.stage.tiles[8][7].teleport.destStage = getStageFromTeam(p.getTeamNameSync())
	return destroyEveryotherInteractable(i, p, t)
}

func getStageFromTeam(s string) string {
	if s == "fuchsia" {
		return "team-fuchsia:4-3"
	} else if s == "sky-blue" {
		return "team-blue:3-4"
	} else {
		return "team-blue:0-0"
	}
}

func destroyEveryotherInteractable(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	tiles := everyOtherTileOnStage(t)
	for i := range tiles {
		go destroyInteractable(tiles[i], p)
	}
	return nil, false
}

// Create an always false auth that notifies to prevent consuming incoming
func notify(notification string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		p.updateBottomText(notification)
		return i, true
	}
}
