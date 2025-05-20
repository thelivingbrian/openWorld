package main

import (
	"fmt"
	"math/rand"
	"time"
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
		"black-hole": {
			{ReactsWith: interactableHasName("ball-fuchsia"), Reaction: hideByTeam("fuchsia")},
			{ReactsWith: interactableHasName("ball-sky-blue"), Reaction: hideByTeam("sky-blue")},
			{ReactsWith: everything, Reaction: eat},
		},
		"goal-sky-blue": {
			{ReactsWith: playerTeamAndBallNameMatch("sky-blue"), Reaction: scoreGoalForTeam("sky-blue")},
			{ReactsWith: PlayerAndTeamMatchButDifferentBall("sky-blue"), Reaction: pass},
			{ReactsWith: interactableIsNil, Reaction: showScoreToPlayer("sky-blue")},
		},
		"goal-fuchsia": {
			{ReactsWith: playerTeamAndBallNameMatch("fuchsia"), Reaction: scoreGoalForTeam("fuchsia")},
			{ReactsWith: PlayerAndTeamMatchButDifferentBall("fuchsia"), Reaction: pass},
			{ReactsWith: interactableIsNil, Reaction: showScoreToPlayer("fuchsia")},
		},

		////////////////////////////////////////////////////////////////
		// Tutorial :
		"tutorial-black-hole": {
			{ReactsWith: interactableIsABall, Reaction: tutorial2HideAndNotify},
			{ReactsWith: everything, Reaction: eat},
		},
		"tutorial-goal-sky-blue": {
			{ReactsWith: playerTeamAndBallNameMatch("sky-blue"), Reaction: finishTutorial},
			{ReactsWith: PlayerAndTeamMatchButDifferentBall("sky-blue"), Reaction: notifyAndPass("Try using the matching ball.")},
		},
		"tutorial-goal-fuchsia": {
			{ReactsWith: playerTeamAndBallNameMatch("fuchsia"), Reaction: finishTutorial},
			{ReactsWith: PlayerAndTeamMatchButDifferentBall("fuchsia"), Reaction: notifyAndPass("Try using the matching ball.")},
		},
		"gold-target": {
			{ReactsWith: interactableHasName("ball-gold"), Reaction: destroyInRangeSkipingSelf(5, 3, 10, 8)},
		},
		"set-team-wild-text-and-delete": {
			{ReactsWith: interactableIsNil, Reaction: setTeamWildText},
		},
		"tutorial-exchange": {
			{ReactsWith: interactableIsARing, Reaction: tutorialExchange},
		},
		"teleport-home": {
			{ReactsWith: interactableIsNil, Reaction: teleportHomeInteraction},
		},

		////////////////////////////////////////////////////////////////
		// machines :

		// catapult
		"catapult-west": {
			{ReactsWith: interactableIsNil, Reaction: catapultWest},
			{ReactsWith: everything, Reaction: pass},
		},
		"catapult-north": {
			{ReactsWith: interactableIsNil, Reaction: catapultNorth},
			{ReactsWith: everything, Reaction: pass},
		},
		"catapult-south": {
			{ReactsWith: interactableIsNil, Reaction: catapultSouth},
			{ReactsWith: everything, Reaction: pass},
		},
		"catapult-east": {
			{ReactsWith: interactableIsNil, Reaction: catapultEast},
			{ReactsWith: everything, Reaction: pass},
		},

		// environment
		"lily-pad": {
			{ReactsWith: interactableIsNil, Reaction: playSoundForAll("water-splash")},
			{ReactsWith: interactableIsARing, Reaction: makeDangerousForOtherTeam},
			{ReactsWith: everything, Reaction: pass},
		},
		"death-trap": {
			{ReactsWith: interactableIsNil, Reaction: killInstantly},
			{ReactsWith: everything, Reaction: pass},
		},
		"exchange-ring": {
			{ReactsWith: interactableIsARing, Reaction: damageAndSpawn},
		},

		////////////////////////////////////////////////////////////////
		// Puzzles:

		// Lavander Ball Puzzle
		"target-lavender": {
			{ReactsWith: interactableHasName("ball-lavender"), Reaction: checkSolveAndRemoveInteractable},
			{ReactsWith: everything, Reaction: pass},
		},
		"target-dark-lavender": {
			{ReactsWith: interactableHasName("ball-dark-lavender"), Reaction: awardPuzzleHat},
			{ReactsWith: everything, Reaction: pass},
		},

		// Airlock
		"airlock-close": {{ReactsWith: interactableIsNil, Reaction: closeAirlockDoors}},
		"airlock-arm":   {{ReactsWith: interactableIsNil, Reaction: armAirlockDoors}},
		"airlock-open":  {{ReactsWith: interactableIsNil, Reaction: openAirlockDoors}},
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

func never(*Interactable, *Player) bool {
	return false
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

func finishTutorial(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	destroyEveryotherInteractable(i, p, t)
	p.goalsScored.CompareAndSwap(0, 1)
	p.updateBottomText("You scored a goal! View stats in menu... ")
	go func() {
		time.Sleep(time.Millisecond * time.Duration(1600))
		ownLock := p.tangibilityLock.TryLock()
		if !ownLock || !p.tangible {
			return
		}
		defer p.tangibilityLock.Unlock()
		openStatsMenu(p)
	}()
	return nil, false
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
var SCORE_TO_WIN = 100

func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) (outgoing *Interactable, ok bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		if team != p.getTeamNameSync() {
			logger.Error().Msg("ERROR TEAM CHECK FAILED - " + p.getTeamNameSync() + " is not equal to: " + team)
			return nil, false
		}

		// Scoreboard
		score := p.world.leaderBoard.scoreboard.Increment(team)
		oppositeTeamName := oppositeTeamName(team)
		scoreOpposing := p.world.leaderBoard.scoreboard.GetScore(oppositeTeamName)

		// Award hat
		p.incrementGoalsScored()
		p.addHatByName("score-1-goal", true)
		p.addAccomplishmentByName(scoreAGoal)

		// Database
		p.updateRecord()
		p.world.db.saveScoreEvent(p.getTileSync(), p, fmt.Sprintf("Notes - Team: %s, Score %d", team, score))

		// World Updates
		broadcastUpdate(p.world, soundTriggerByName("huge-explosion"))
		if score < SCORE_TO_WIN {
			message := fmt.Sprintf("@[%s|%s] scored a goal!<br /> The score is: @[%s %d|%s] to @[%s %d|%s]", p.username, team, team, score, team, oppositeTeamName, scoreOpposing, oppositeTeamName)
			broadcastBottomText(p.world, message)
		} else {
			// Awards
			awardHatByTeam(p.world, team, "winning-team", true)
			p.addAccomplishmentByName(winningAGame)

			// Games won stat?

			// Reset and notify
			p.world.leaderBoard.scoreboard.ResetAll()
			message := fmt.Sprintf("@[%s|%s] won the game for @[%s|%s]!", p.username, team, team, team)
			broadcastBottomText(p.world, message)
		}

		saveCurrentStatus(p.world)

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

func showScoreToPlayer(team string) func(*Interactable, *Player, *Tile) (outgoing *Interactable, ok bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		// Get scores
		score := p.world.leaderBoard.scoreboard.GetScore(team)
		oppositeTeamName := oppositeTeamName(team)
		scoreOpposing := p.world.leaderBoard.scoreboard.GetScore(oppositeTeamName)

		// Add a sound?
		//broadcastUpdate(p.world, soundTriggerByName("info?"))

		message := fmt.Sprintf("The score is - @[%s %d|%s] to @[%s %d|%s]. ", team, score, team, oppositeTeamName, scoreOpposing, oppositeTeamName)

		playerTeam := p.getTeamNameSync()
		rand := rand.Intn(2) // Displaying both messages is too much - especially on mobile
		if team == playerTeam && rand > 0 {
			message = fmt.Sprintf("...Find the @[Ball|%s] to score a point!", playerTeam)
		}

		p.updateBottomText(message)

		return nil, false
	}
}

func placeInteractableOnStagePriorityCovered(stage *Stage, interactable *Interactable) {
	covered, uncovered := sortWalkableTiles(stage.tiles)
	if len(covered)+len(uncovered) == 0 {
		logger.Error().Msg("Trying to place interactable but no valid tiles on: " + stage.name)
		return
	}

	tiles := append([]*Tile(nil), covered...)
	for {
		if len(tiles) == 0 {
			tiles = append([]*Tile(nil), uncovered...)
			tiles = append([]*Tile(nil), covered...)
		}
		index := rand.Intn(len(tiles))
		if trySetInteractable(tiles[index], interactable) {
			return
		}
		tiles = append(tiles[:index], tiles[index+1:]...)
	}
}

func tryPlaceInteractableOnStage(stage *Stage, interactable *Interactable) {
	tiles, uncovered := sortWalkableTiles(stage.tiles)
	tiles = append(tiles, uncovered...)
	for i := 0; i < len(tiles); i++ {
		index := rand.Intn(len(tiles))
		if trySetInteractable(tiles[index], interactable) {
			return
		}
	}
}

////////////////////////////////////////////////////////////////
// machines

func moveInitiator(yOff, xOff int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		move(p, yOff, xOff)
		return i, true
	}
}

func killInstantly(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	handleDeath(p)
	return nil, false
}

func moveInitiatorPushSurrounding(yOff, xOff int) func(*Interactable, *Player, *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		for _, tile := range getVanNeumannNeighborsOfTile(t) {
			p.push(tile, nil, yOff, xOff)
		}
		move(p, yOff, xOff)
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
		{
			ReactsWith: playerHasTeam(oppositeTeamName(initiatorTeam)),
			Reaction:   damageWithinRadiusAndReset(2, dmg, p.id),
		},
		{ReactsWith: interactableIsNil, Reaction: playSoundForAll("water-splash")},
		{ReactsWith: everything, Reaction: pass},
	}
	t.interactable.cssClass = initiatorTeam + "-b thick r0"
	t.interactable.reactions = newReactions
	t.stage.updateAll(interactableBoxSpecific(t.y, t.x, t.interactable))
	addMoneyToStage(t.stage, 50)
	addMoneyToStage(t.stage, dmg)
	return nil, false
}

func damageAndSpawn(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	y := rand.Intn(len(t.stage.tiles))
	x := rand.Intn(len(t.stage.tiles[y]))
	epicenter := t.stage.tiles[y][x]
	dmg := 50
	powerToSpawn := 3
	if i.name == "ring-big" {
		dmg = 100
		powerToSpawn = 5
	}
	go damageWithinRadius(epicenter, p.world, 4, dmg, p.id)
	t.stage.updateAll(soundTriggerByName("explosion"))
	addMoneyToStage(t.stage, dmg/5)
	addMoneyToStage(t.stage, dmg/5)
	addMoneyToStage(t.stage, dmg/2)
	for i := 0; i < powerToSpawn; i++ {
		spawnPowerup(t.stage)
	}
	return nil, false
}

func tutorialExchange(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	y := rand.Intn(len(t.stage.tiles))
	x := rand.Intn(len(t.stage.tiles[y]))
	epicenter := t.stage.tiles[y][x]
	dmg := 50
	powerToSpawn := 2
	go damageWithinRadius(epicenter, p.world, 4, dmg, p.id)
	t.stage.updateAll(soundTriggerByName("explosion"))
	addMoneyToStage(t.stage, 1)
	addMoneyToStage(t.stage, 1)
	addMoneyToStage(t.stage, 1)
	for i := 0; i < powerToSpawn; i++ {
		spawnPowerupShort(t.stage)
	}
	p.updateBottomText("*[red] berries grant you destructive powers")
	return nil, false
}

func teleportHomeInteraction(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	teleport := makeHomeTeleport(p)
	if teleport == nil {
		return nil, false
	}
	p.setMenu("teleport", exitTutorial(teleport))
	turnMenuOnByName(p, "teleport")
	return nil, false
}

func makeHomeTeleport(p *Player) *Teleport {
	team := p.getTeamNameSync()
	stage := ""
	switch team {
	case "fuchsia":
		stage = "team-fuchsia:4-3"
	case "sky-blue":
		stage = "team-blue:3-4"
	default:
		return nil
	}
	return &Teleport{
		destStage:          stage,
		destY:              7,
		destX:              7,
		confirmation:       true,
		rejectInteractable: true,
	}
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
			fragile:  true,
		}
		return &bigring
	}
	ring := Interactable{
		name:     "ring-small",
		cssClass: "gold-b med r1",
		pushable: true,
		fragile:  true,
	}
	return &ring
}

// Does not reset
func damageWithinRadiusAndReset(radius, dmg int, ownerId string) func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	return func(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
		go damageWithinRadius(t, p.world, radius, dmg, ownerId) // damage can take interactable lock that is held by reacting tile
		tryPlaceInteractableOnStage(t.stage, createRing())
		t.interactable.cssClass = "white trsp20 r0"
		t.interactable.reactions = interactableReactions["lily-pad"]
		t.stage.updateAll(interactableBoxSpecific(t.y, t.x, t.interactable) + soundTriggerByName("explosion"))
		return nil, false
	}
}

// not a reaction; will lock tile
func damageWithinRadius(tile *Tile, world *World, radius, dmg int, ownerId string) {
	tiles := getTilesInRadius(tile, radius)
	trapSetter := world.getPlayerById(ownerId)
	if trapSetter != nil {
		damageAndIndicate(tiles, trapSetter, tile.stage, dmg)
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
	// Awards
	p.addHatByName("puzzle-solve", false) // worth having hat for puzzles?
	p.addAccomplishmentByName(puzzle0)

	// add boost 13,5
	p.getTileSync().stage.tiles[13][5].addBoostsAndNotifyAll()

	// spawn escape lily 15,5
	lilySpot := p.getTileSync().stage.tiles[15][5]
	lilySpot.interactableMutex.Lock()
	defer lilySpot.interactableMutex.Unlock()
	lily := Interactable{name: "lily-pad", cssClass: "black trsp20 lime-b thin r0", walkable: true, fragile: true}
	setLockedInteractableAndUpdate(lilySpot, &lily)

	// TODO: reward sound

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

// airlock
func openAirlockDoors(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	topLeft := findTopLeftOpenSwitch(t)
	if topLeft == nil {
		logger.Warn().Msgf("unexpected region for airlock press at %d,%d", t.y, t.x)
		return nil, false
	}

	tiles := getOrderedRegion(t.stage, topLeft.y, topLeft.x+1, 6, 2)
	for _, tile := range tiles {
		tile.interactableMutex.Lock()
		defer tile.interactableMutex.Unlock()

		if tile.interactable != nil && tile.interactable.name == "airlock-door" {
			tile.interactable.walkable = true
			tile.interactable.cssClass = ""
			tile.stage.updateAll(interactableBoxSpecific(tile.y, tile.x, tile.interactable))
		}
		if tile.interactable != nil && tile.interactable.name == "airlock-close-switch" {
			tile.interactable.reactions[0].ReactsWith = never
		}
	}
	return nil, true
}

func findTopLeftOpenSwitch(t *Tile) *Tile {
	s := t.stage

	hasBelow := isAirlockOpenSwitch(s, t.y+5, t.x)
	hasRight := isAirlockOpenSwitch(s, t.y, t.x+3)

	switch {
	case hasBelow && hasRight:
		return t
	case hasBelow && !hasRight:
		return s.tiles[t.y][t.x-3]
	case !hasBelow && hasRight:
		return s.tiles[t.y-5][t.x]
	case !hasBelow && !hasRight:
		return s.tiles[t.y-5][t.x-3]
	}
	return nil
}

func isAirlockOpenSwitch(s *Stage, y, x int) bool {
	if !validCoordinate(y, x, s) {
		return false
	}
	t := s.tiles[y][x]
	t.interactableMutex.Lock()
	defer t.interactableMutex.Unlock()

	if t.interactable != nil && t.interactable.name == "airlock-open-switch" {
		return true
	}
	return false
}

func closeAirlockDoors(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	topLeft := findTopLeftCloseSwitch(t)
	if topLeft == nil {
		logger.Warn().Msgf("unexpected region for airlock press at %d,%d", t.y, t.x)
		return nil, false
	}

	// Risk of hitting starting/player tile
	tiles := getOrderedRegion(t.stage, topLeft.y-2, topLeft.x, 2, 2)
	tiles = append(tiles, getOrderedRegion(t.stage, topLeft.y+2, topLeft.x, 2, 2)...)
	for _, tile := range tiles {
		tile.interactableMutex.Lock()
		defer tile.interactableMutex.Unlock()

		if tile.interactable != nil && tile.interactable.name == "airlock-door" {
			tile.interactable.walkable = false
			tile.interactable.cssClass = "s-hoz chocolate-b thick no-lr"
			tile.stage.updateAll(interactableBoxSpecific(tile.y, tile.x, tile.interactable))
		}
	}

	return nil, false
}

func findTopLeftCloseSwitch(t *Tile) *Tile {
	s := t.stage

	hasBelow := isAirlockCloseSwitch(s, t.y+1, t.x)
	hasRight := isAirlockCloseSwitch(s, t.y, t.x+1)

	switch {
	case hasBelow && hasRight:
		return t
	case hasBelow && !hasRight:
		return s.tiles[t.y][t.x-1]
	case !hasBelow && hasRight:
		return s.tiles[t.y-1][t.x]
	case !hasBelow && !hasRight:
		return s.tiles[t.y-1][t.x-1]
	}
	return nil
}

func isAirlockCloseSwitch(s *Stage, y, x int) bool {
	if !validCoordinate(y, x, s) {
		return false
	}
	t := s.tiles[y][x]
	t.interactableMutex.Lock()
	defer t.interactableMutex.Unlock()

	return t.interactable != nil &&
		t.interactable.name == "airlock-close-switch"
}

func armAirlockDoors(i *Interactable, p *Player, t *Tile) (*Interactable, bool) {
	// Need arm step to prevent close after being saved
	topLeft := findTopLeftDoor(t)
	if topLeft == nil {
		logger.Warn().Msgf("unexpected region for airlock press at %d,%d", t.y, t.x)
		return nil, false
	}

	// Risk of hitting starting/player tile
	tiles := getOrderedRegion(t.stage, topLeft.y+2, topLeft.x, 2, 2)
	for _, tile := range tiles {
		tile.interactableMutex.Lock()
		defer tile.interactableMutex.Unlock()

		if tile.interactable != nil && tile.interactable.name == "airlock-close-switch" {
			tile.interactable.reactions[0].ReactsWith = interactableIsNil
		}
	}

	return nil, false
}

func findTopLeftDoor(t *Tile) *Tile {
	s := t.stage

	hasBelow := isAirlockDoor(s, t.y+1, t.x)
	hasRight := isAirlockDoor(s, t.y, t.x+1)
	topGroup := isAirlockDoor(s, t.y+4, t.x) || isAirlockArmOnly(s, t.y+4, t.x)

	yAdjustment := 0
	if !topGroup {
		yAdjustment = -4
	}

	switch {
	case hasBelow && hasRight:
		return s.tiles[t.y+yAdjustment][t.x]
	case hasBelow && !hasRight:
		return s.tiles[t.y+yAdjustment][t.x-1]
	case !hasBelow && hasRight:
		return s.tiles[t.y-1+yAdjustment][t.x]
	case !hasBelow && !hasRight:
		return s.tiles[t.y-1+yAdjustment][t.x-1]
	}
	return nil
}

func isAirlockDoor(s *Stage, y, x int) bool {
	if !validCoordinate(y, x, s) {
		return false
	}
	t := s.tiles[y][x]
	t.interactableMutex.Lock()
	defer t.interactableMutex.Unlock()

	return t.interactable != nil &&
		t.interactable.name == "airlock-door"
}

func isAirlockArmOnly(s *Stage, y, x int) bool {
	if !validCoordinate(y, x, s) {
		return false
	}
	t := s.tiles[y][x]
	t.interactableMutex.Lock()
	defer t.interactableMutex.Unlock()

	return t.interactable != nil &&
		t.interactable.name == "airlock-arm-only"
}
