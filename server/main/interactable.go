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

var interactableReactions map[string][]InteractableReaction

func init() {
	interactableReactions = map[string][]InteractableReaction{
		// Capture the flag :
		"black-hole": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableHasName("ball-fuchsia"), Reaction: hideByTeam("fuchsia")},
			InteractableReaction{ReactsWith: interactableHasName("ball-sky-blue"), Reaction: hideByTeam("sky-blue")},
			InteractableReaction{ReactsWith: everything, Reaction: eat},
		},
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
			InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("sky-blue"), Reaction: notifyAndPass("Try using the matching ball.")},
		},
		"tutorial-goal-fuchsia": []InteractableReaction{
			InteractableReaction{ReactsWith: playerTeamAndBallNameMatch("fuchsia"), Reaction: destroyEveryotherInteractable},
			InteractableReaction{ReactsWith: PlayerAndTeamMatchButDifferentBall("fuchsia"), Reaction: notifyAndPass("Try using the matching ball.")},
		},
		"gold-target": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableHasName("ball-gold"), Reaction: destroyInRangeSkipingSelf(5, 3, 10, 8)},
		},
		"set-team-wild-text-and-delete": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: setTeamWildText},
		},
		// machines :
		"catapult-west": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: catapultWest},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"catapult-north": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: catapultNorth},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"catapult-south": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: catapultSouth},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"catapult-east": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: catapultEast},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"lily-pad": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: eat},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
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

////////////////////////////////////////////////////////////////////////
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
		return interactableIsABall(i, nil) && team == p.getTeamNameSync()
	}
}

////////////////////////////////////////////////////////////////////////////
// Rections

// Basic
func eat(*Interactable, *Player, *Tile) (*Interactable, bool) {
	// incoming interactable is discarded
	return nil, false
}

func pass(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	return i, true
}

// Spawn and destroy
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

// Capture the flag
func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) (outgoing *Interactable, ok bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		if team != p.getTeamNameSync() {
			fmt.Println("ERROR TEAM CHECK FAILED? ", p.getTeamNameSync(), team)
			return nil, false
		}
		p.world.leaderBoard.scoreboard.Increment(team)
		scoreSkyBlue := p.world.leaderBoard.scoreboard.GetScore("sky-blue")
		scoreFuchsia := p.world.leaderBoard.scoreboard.GetScore("fuchsia")
		//fmt.Println(scoreSkyBlue)

		totalGoals := p.incrementGoalsScored()
		if totalGoals == 1 {
			p.addHatByName("score-1-goal")
		}
		p.updateRecord()
		message := fmt.Sprintf("@[%s|%s] scored a goal!<br /> The score is: @[Sky-blue %d|sky-blue] to @[Fuchsia %d|fuchsia]", p.username, team, scoreSkyBlue, scoreFuchsia)
		broadcastBottomText(p.world, message)

		return hideByTeam(team)(i, p, t)
	}
}

func hideByTeam(team string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		lat := rand.Intn(8)
		long := rand.Intn(8)
		stagename := "arcade:0-2"
		if team == "sky-blue" {
			stagename = fmt.Sprintf("team-fuchsia:%d-%d", lat, long)
		}
		if team == "fuchsia" {
			stagename = fmt.Sprintf("team-blue:%d-%d", lat, long)
		}
		fmt.Println(stagename)
		stage := p.fetchStageSync(stagename)
		tiles, uncovered := sortWalkableTiles(stage.tiles)
		if len(tiles) == 0 {
			tiles = uncovered
		}
		placed := false
		for !placed {
			index := rand.Intn(len(tiles))
			placed = trySetInteractable(tiles[index], i)
		}

		return nil, false
	}
}

////////////////////////////////////////////////////////////////
// machines

func moveInitiator(yOff, xOff int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		p.move(yOff, xOff)
		return i, true
	}
}

var superBoostEast = moveInitiator(0, 11)
var superBoostWest = moveInitiator(0, -11)
var superBoostNorth = moveInitiator(-11, 0)
var superBoostSouth = moveInitiator(11, 0)

func moveInitiatorPushSurrounding(yOff, xOff int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		fmt.Println("base", t.y, t.x)
		for _, tile := range getVanNeumannNeighborsOfTile(t) {
			fmt.Println("pushing", tile.y, tile.x)
			p.push(tile, nil, yOff, xOff)
		}
		p.move(yOff, xOff)
		return nil, false
	}
}

var catapultEast = moveInitiatorPushSurrounding(0, 11)
var catapultWest = moveInitiatorPushSurrounding(0, -11)
var catapultNorth = moveInitiatorPushSurrounding(-11, 0)
var catapultSouth = moveInitiatorPushSurrounding(11, 0)

// Tutorial
func setTeamWildText(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	t.bottomText = teamColorWildRegex.ReplaceAllString(t.material.DisplayText, `@[$1|`+p.getTeamNameSync()+`]`)
	setLockedInteractableAndUpdate(t, nil)
	return nil, true
}

// Create an always false auth that notifies to prevent consuming incoming
func notifyAndPass(notification string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		p.updateBottomText(notification)
		return i, true
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
