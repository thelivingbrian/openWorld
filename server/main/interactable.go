package main

import (
	"fmt"
	"math/rand"
	"strings"
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
			InteractableReaction{ReactsWith: everything, Reaction: eat}, // reduces number of rings
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
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: playSoundForAll("water-splash")},
			InteractableReaction{ReactsWith: interactableIsARing, Reaction: makeDangerousForOtherTeam},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"death-trap": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsNil, Reaction: killInstantly},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"exchange-ring": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableIsARing, Reaction: damageAndSpawn},
		},
		// Puzzles:
		"target-lavender": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableHasName("ball-lavender"), Reaction: checkSolveAndRemoveInteractable},
			InteractableReaction{ReactsWith: everything, Reaction: pass},
		},
		"target-dark-lavender": []InteractableReaction{
			InteractableReaction{ReactsWith: interactableHasName("ball-dark-lavender"), Reaction: awardPuzzleHat},
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

func playerHasTeam(team string) func(*Interactable, *Player) bool {
	return func(_ *Interactable, p *Player) bool {
		if p == nil {
			return false
		}
		return p.getTeamNameSync() == team
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

func interactableIsARing(i *Interactable, _ *Player) bool {
	if i == nil {
		return false
	}
	if len(i.name) < 5 {
		return false
	}
	return i.name[0:5] == "ring-"
}

func oppositeTeamName(team string) string {
	if team == "sky-blue" {
		return "fuchsia"
	}
	if team == "fuchsia" {
		return "sky-blue"
	}
	return ""
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

func playSoundForInitiator(soundName string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		sendSoundToPlayer(p, soundName)
		return nil, false
	}
}

func playSoundForAll(soundName string) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		t.stage.updateAll(soundTriggerByName(soundName))
		return nil, false
	}
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
			logger.Error().Msg("ERROR TEAM CHECK FAILED - " + p.getTeamNameSync() + " is not equal to: " + team)
			return nil, false
		}
		p.world.leaderBoard.scoreboard.Increment(team) // should return result ?
		score := p.world.leaderBoard.scoreboard.GetScore(team)
		oppositeTeamName := oppositeTeamName(team)
		scoreOpposing := p.world.leaderBoard.scoreboard.GetScore(oppositeTeamName)

		totalGoals := p.incrementGoalsScored()
		if totalGoals == 1 {
			p.addHatByName("score-1-goal")
		}
		if totalGoals == 1000 {
			p.addHatByName("score-1000-goal")
		}
		p.updateRecord()
		message := fmt.Sprintf("@[%s|%s] scored a goal!<br /> The score is: @[%s %d|%s] to @[%s %d|%s]", p.username, team, team, score, team, oppositeTeamName, scoreOpposing, oppositeTeamName)
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
		logger.Info().Msg("Ball is hidden on: " + stagename)
		stage := p.fetchStageSync(stagename)
		placeInteractableOnStagePriorityCovered(stage, i)

		return nil, false
	}
}

func placeInteractableOnStagePriorityCovered(stage *Stage, interactable *Interactable) {
	tiles, uncovered := sortWalkableTiles(stage.tiles)
	if len(tiles) == 0 {
		tiles = uncovered
	}
	placed := false
	for !placed {
		index := rand.Intn(len(tiles))
		placed = trySetInteractable(tiles[index], interactable)
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

func killInstantly(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	handleDeath(p)
	return nil, false
}

var superBoostEast = moveInitiator(0, 11)
var superBoostWest = moveInitiator(0, -11)
var superBoostNorth = moveInitiator(-11, 0)
var superBoostSouth = moveInitiator(11, 0)

func moveInitiatorPushSurrounding(yOff, xOff int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		for _, tile := range getVanNeumannNeighborsOfTile(t) {
			p.push(tile, nil, yOff, xOff)
		}
		p.move(yOff, xOff)
		sendSoundToPlayer(p, "wind-swoosh")
		t.stage.updateAllExcept(soundTriggerByName("woody-swoosh"), p)
		return nil, false
	}
}

var catapultEast = moveInitiatorPushSurrounding(0, 11)
var catapultWest = moveInitiatorPushSurrounding(0, -11)
var catapultNorth = moveInitiatorPushSurrounding(-11, 0)
var catapultSouth = moveInitiatorPushSurrounding(11, 0)

func makeDangerousForOtherTeam(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	initiatorTeam := p.getTeamNameSync()
	dmg := 50
	if i.name == "ring-big" {
		dmg = 100
	}
	newReactions := []InteractableReaction{
		InteractableReaction{
			ReactsWith: playerHasTeam(oppositeTeamName(initiatorTeam)),
			Reaction:   damageWithinRadiusAndReset(2, dmg, p.id),
		},
		InteractableReaction{ReactsWith: interactableIsNil, Reaction: playSoundForAll("water-splash")},
		InteractableReaction{ReactsWith: everything, Reaction: pass},
	}
	t.interactable.cssClass = initiatorTeam + "-b thick r0"
	t.interactable.reactions = newReactions
	t.stage.updateAll(lockedInteractableBox(t))
	addMoneyToStage(t.stage, 10) // should be more money
	return nil, false
}

func damageAndSpawn(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	y := rand.Intn(len(t.stage.tiles))
	x := rand.Intn(len(t.stage.tiles[y]))
	epicenter := t.stage.tiles[y][x]
	dmg := 50
	if i.name == "ring-big" {
		dmg = 100
	}
	go damageWithinRadius(epicenter, p.world, 4, dmg, p.id)
	t.stage.updateAll(soundTriggerByName("explosion"))
	addMoneyToStage(t.stage, dmg/5)
	if strings.Contains(t.stage.name, ":") {
		spacename := strings.Split(t.stage.name, ":")[0]
		lat := rand.Intn(8)
		long := rand.Intn(8)
		stagename := fmt.Sprintf("%s:%d-%d", spacename, lat, long)
		logger.Info().Msg(stagename)
		stage := p.fetchStageSync(stagename)
		placeInteractableOnStagePriorityCovered(stage, i)
	}
	return nil, false
}

func addMoneyToStage(stage *Stage, amount int) {
	walkableTiles := walkableTiles(stage.tiles)
	n := rand.Intn(len(walkableTiles))
	walkableTiles[n].addMoneyAndNotifyAll(amount)
}

func createRing() *Interactable {
	n := rand.Intn(10)
	if n == 0 {
		bigring := Interactable{
			name:     "ring-big",
			cssClass: "gold-b thick r1",
			pushable: true,
		}
		return &bigring
	}
	ring := Interactable{
		name:     "ring-small",
		cssClass: "gold-b med r1",
		pushable: true,
	}
	return &ring
}

func damageWithinRadiusAndReset(radius, dmg int, ownerId string) func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		go damageWithinRadius(t, p.world, radius, dmg, ownerId) // damage can take interactable lock that is held by reacting tile
		placeInteractableOnStagePriorityCovered(t.stage, createRing())
		t.interactable.cssClass = "white trsp20 r0"
		t.interactable.reactions = interactableReactions["lily-pad"]
		t.stage.updateAll(lockedInteractableBox(t) + soundTriggerByName("explosion"))
		return nil, false
	}
}

// not a reaction; will lock tile
func damageWithinRadius(tile *Tile, world *World, radius, dmg int, ownerId string) {
	tiles := getTilesInRadius(tile, radius)
	trapSetter := world.getPlayerById(ownerId)
	if trapSetter != nil {
		trapSetter.tangibilityLock.Lock()
		defer trapSetter.tangibilityLock.Unlock()
		if trapSetter.tangible {
			damageAndIndicate(tiles, trapSetter, tile.stage, dmg)
		}
	}
}

///////////////////////////////////////////////////////////////
// Puzzles

func checkSolveAndRemoveInteractable(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	setLockedInteractableAndUpdate(t, nil)
	stage := t.stage
	tileA := stage.tiles[3][3]
	tileB := stage.tiles[12][12]
	interactableA := tryGetInteractable(tileA)
	interactableB := tryGetInteractable(tileB)
	if interactableA != nil && interactableA.name == "fragile-push" && interactableB != nil && interactableB.name == "basic-push" {
		reward := Interactable{name: "ball-dark-lavender", cssClass: "dark-lavender r1 lavender-b thick", pushable: true, fragile: true}
		p.push(stage.tiles[7][7], &reward, 0, 1)
	}
	return nil, false
}

func awardPuzzleHat(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	p.addHatByName("puzzle-solve-1")
	// spawn escape lily
	return nil, false
}

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
	p.updateBottomText("black holes will absorb balls and spit them out elsewhere")
	return nil, false
}
