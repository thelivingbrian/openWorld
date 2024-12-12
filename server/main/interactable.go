package main

import (
	"fmt"
	"math/rand"
)

type Interactable struct {
	name      string
	pushable  bool
	cssClass  string
	fragile   bool
	reactions []InteractableReaction
}

type InteractableReaction struct {
	ReactsWith func(*Interactable) bool
	Reaction   func(incoming *Interactable, initiatior *Player, location *Tile)
}

var interactableReactions = map[string][]InteractableReaction{
	"black-hole": []InteractableReaction{InteractableReaction{ReactsWith: everything, Reaction: eat}},
	"fuchsia-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: matchesName("fuchsia-ball"), Reaction: scoreGoalForTeam("sky-blue")},
		InteractableReaction{ReactsWith: matchesName("sky-blue-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
	"sky-blue-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: matchesName("sky-blue-ball"), Reaction: scoreGoalForTeam("fuchsia")},
		InteractableReaction{ReactsWith: matchesName("fuchsia-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
}

func (source *Interactable) React(incoming *Interactable, initiatior *Player, location *Tile) bool {
	if source.reactions == nil {
		return false
	}
	for i := range source.reactions {
		if source.reactions[i].ReactsWith != nil && source.reactions[i].ReactsWith(incoming) {
			source.reactions[i].Reaction(incoming, initiatior, location)
			return true
		}
	}
	return false
}

// Gates
func everything(*Interactable) bool {
	return true
}

func matchesCssClass(cssClass string) func(*Interactable) bool {
	return func(i *Interactable) bool {
		return i.cssClass == cssClass
	}
}

func matchesName(name string) func(*Interactable) bool {
	return func(i *Interactable) bool {
		if i == nil {
			return false
		}
		return i.name == name
	}
}

// Actions
func eat(*Interactable, *Player, *Tile) {
	// incoming interactable is discarded
}

func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) {
	return func(i *Interactable, p *Player, t *Tile) {
		p.world.leaderBoard.scoreboard.Increment(team)
		if team == p.getTeamNameSync() {
			// Otherwise you have scored a goal for a different team
			p.incrementGoalsScored()
			p.updateRecord()
		}
		fmt.Println(p.world.leaderBoard.scoreboard.GetScore(team))
	}
}

func spawnMoney(amounts []int) func(*Interactable, *Player, *Tile) {
	return func(i *Interactable, p *Player, t *Tile) {
		tiles := walkableTiles(t.stage.tiles)
		count := len(tiles)
		if count == 0 {
			return
		}
		for i := range amounts {
			randn := rand.Intn(count)
			tiles[randn].addMoneyAndNotifyAll(amounts[i])
		}
	}
}
