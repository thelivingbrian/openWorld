package main

import (
	"fmt"
	"math/rand"
)

type Interactable struct {
	name     string
	pushable bool
	//walkable       bool // problematic, because push onto walkable is undefined?
	cssClass       string
	fragile        bool
	reactions      []InteractableReaction // Lowest index match wins
	rejectTeleport bool
}

type InteractableReaction struct {
	ReactsWith func(incoming *Interactable, initiatior *Player) bool
	Reaction   func(incoming *Interactable, initiatior *Player, location *Tile) (outgoing *Interactable, push bool) // rotate ?
}

var interactableReactions map[string][]InteractableReaction

func init() {
	interactableReactions = map[string][]InteractableReaction{
		// Capture the flag :
		"black-hole": []InteractableReaction{
			InteractableReaction{ReactsWith: everything, Reaction: eat}},
		"goal-sky-blue": []InteractableReaction{
			InteractableReaction{ReactsWith: playerTeamAndBallNameMatch("sky-blue"), Reaction: scoreGoalForTeam("sky-blue")},
			InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("sky-blue"), Reaction: pass},
		},
		"goal-fuchsia": []InteractableReaction{
			InteractableReaction{ReactsWith: playerTeamAndBallNameMatch("fuchsia"), Reaction: scoreGoalForTeam("fuchsia")},
			InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("fuchsia"), Reaction: pass},
		},
		// Tutorial :
		"tutorial-black-hole": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsABall, Reaction: tutorial2HideAndNotify},
			InteractableReaction{ReactsWith: everything, Reaction: eat},
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
			InteractableReaction{ReactsWith: interactableHasName("ball-gold"), Reaction: destroyInRangeSkipingSelf(5, 3, 10, 8)},
		},
		"set-team-wild-text-and-delete": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: setTeamWildText},
		},
	}

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

func interactableIsNil(i *Interactable, p *Player) bool {
	return i == nil
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

func interactableIsABall(i *Interactable, _ *Player) bool {
	if i == nil {
		return false
	}
	if len(i.name) < 5 {
		return false
	}
	return i.name[0:5] == "ball-"
}

func PlayerAndTeamMatchButDifferentBall(team string) func(*Interactable, *Player) bool {
	return func(i *Interactable, p *Player) bool {
		// if i == nil {
		// 	return false
		// }
		// if len(i.name) < 5 {
		// 	return false
		// }
		return interactableIsABall(i, nil) && team == p.getTeamNameSync()
	}
}

// Actions
func eat(*Interactable, *Player, *Tile) (*Interactable, bool) {
	// incoming interactable is discarded
	return nil, false
}

func pass(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	return i, true
}

func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) (outgoing *Interactable, ok bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		if team != p.getTeamNameSync() {
			fmt.Println("ERROR TEAM CHECK FAILED? ", p.getTeamNameSync(), team)
			return nil, false
		}
		p.world.leaderBoard.scoreboard.Increment(team)
		score := p.world.leaderBoard.scoreboard.GetScore(team)
		fmt.Println(score)

		p.incrementGoalsScored()
		p.updateRecord()
		message := fmt.Sprintf("@[%s|%s] has scored a goal! @[Team %s|%s] now has @[%d|%s] points!", p.username, team, team, team, score, team)
		broadcastBottomText(p.world, message)

		return nil, false
	}
}

/*
func scoreGoalForPlayer(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	team := p.getTeamNameSync()
	p.world.leaderBoard.scoreboard.Increment(team)
	p.incrementGoalsScored()
	p.updateRecord()
	// need to  give a trim
	fmt.Println(p.world.leaderBoard.scoreboard.GetScore(team))
	return nil, false
}
*/

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

func setTeamWildText(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	t.bottomText = teamColorWildRegex.ReplaceAllString(t.material.DisplayText, `@[$1|`+p.getTeamNameSync()+`]`)
	//fmt.Println(t.material.DisplayText)
	swapInteractableAndUpdate(t, nil)
	//t.interactable = nil
	return nil, true
}

func destroyEveryotherInteractable(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	tiles := everyOtherTileOnStage(t)
	for i := range tiles {
		go destroyInteractable(tiles[i], p)
	}
	return nil, false
}

func destroyInRange(yMin, xMin, yMax, xMax int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		tiles := t.stage.tiles
		if yMin >= len(tiles) || yMax >= len(tiles) {
			return nil, false
		}
		if xMin >= len(tiles[yMin]) || xMax >= len(tiles[yMin]) {
			return nil, false
		}
		for i := yMin; i <= yMax; i++ {
			for j := xMin; j <= xMax; j++ {
				go destroyInteractable(tiles[i][j], p)
			}
		}
		return nil, false
	}
}

func destroyInRangeSkipingSelf(yMin, xMin, yMax, xMax int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		tiles := t.stage.tiles
		if yMin >= len(tiles) || yMax >= len(tiles) {
			return nil, false
		}
		if xMin >= len(tiles[yMin]) || xMax >= len(tiles[yMin]) {
			return nil, false
		}
		for i := yMin; i <= yMax; i++ {
			for j := xMin; j <= xMax; j++ {
				if tiles[i][j] == t {
					continue
				}
				go destroyInteractable(tiles[i][j], p)
			}
		}
		return nil, false
	}
}

func tutorial2HideAndNotify(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	stagenames := []string{"tutorial2:0-1", "tutorial2:0-2", "tutorial2:1-2", "tutorial2:2-0", "tutorial2:2-1"}
	index := rand.Intn(5)
	stagename := stagenames[index]
	stage := p.fetchStageSync(stagename)
	tiles := walkableTiles(stage.tiles)
	placed := false
	for !placed {
		index = rand.Intn(len(tiles))
		placed = trySetInteractable(tiles[index], i)
	}
	p.updateBottomText("@[black holes|black] will absorb balls and spit them out elsewhere")
	return nil, false
}

func trySetInteractable(tile *Tile, i *Interactable) bool {
	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false
	}
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		return false
	}
	tile.interactable = i
	return true
}

// Create an always false auth that notifies to prevent consuming incoming
func notify(notification string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		p.updateBottomText(notification)
		return i, true
	}
}
