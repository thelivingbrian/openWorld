package main

import "sync"

type Actions struct {
	spaceHighlights     map[*Tile]bool
	spaceHighlightMutex sync.Mutex
	spaceStack          *StackOfPowerUp
	boostCounter        int
	boostMutex          sync.Mutex
}

type PowerUp struct {
	areaOfInfluence [][2]int
	//damageAtRadius  [4]int // unused
}

type StackOfPowerUp struct {
	powers     []*PowerUp
	powerMutex sync.Mutex
}

func createDefaultActions() *Actions {
	return &Actions{
		spaceHighlights:     map[*Tile]bool{},
		spaceHighlightMutex: sync.Mutex{},
		spaceStack:          &StackOfPowerUp{},
		boostCounter:        0,
		boostMutex:          sync.Mutex{},
	}
}

/////////////////////////////////////////////////////////
// Space Highlights

func (player *Player) setSpaceHighlights() {
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	player.actions.spaceHighlights = map[*Tile]bool{}
	currentTile := player.getTileSync()
	absCoordinatePairs := findOffsetsGivenPowerUp(currentTile.y, currentTile.x, player.actions.spaceStack.peek())
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
		}
	}
}

func (player *Player) updateSpaceHighlights() []*Tile { // Returns removed highlights
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	currentTile := player.getTileSync()
	absCoordinatePairs := findOffsetsGivenPowerUp(currentTile.y, currentTile.x, player.actions.spaceStack.peek())
	var impactedTiles []*Tile
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
			if _, contains := previous[tile]; contains {
				delete(previous, tile)
			} else {
				impactedTiles = append(impactedTiles, tile)
			}
		}
	}
	return append(impactedTiles, mapOfTileToArray(previous)...)
}

func (player *Player) activatePower() {
	playerHighlights := highlightMapToSlice(player)
	damageAndIndicate(playerHighlights, player, 50)
	sendSoundToPlayer(player, "explosion")
	updateOne(sliceOfTileToHighlightBoxes(playerHighlights, ""), player)

	//player.actions.spaceHighlights = map[*Tile]bool{}
	_, powerCount := player.actions.spaceStack.pop()
	if powerCount > 0 {
		player.setSpaceHighlights()
		updateOne(sliceOfTileToHighlightBoxes(highlightMapToSlice(player), spaceHighlighter()), player)
	}
	updateOne(spanPower(powerCount), player)
}

/*
func (player *Player) nextPower() {
	count := player.actions.spaceStack.pop() // Throw old power away
	player.setSpaceHighlights()
	updateOne(sliceOfTileToHighlightBoxes(highlightMapToSlice(player), spaceHighlighter())+spanPower(count), player)
}
*/

func highlightMapToSlice(player *Player) []*Tile {
	out := make([]*Tile, 0)
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	for tile := range player.actions.spaceHighlights {
		out = append(out, tile)
	}
	return out
}

// Power up stack
func (stack *StackOfPowerUp) pop() (*PowerUp, int) {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	var power *PowerUp
	if len(stack.powers) > 0 {
		power = stack.powers[len(stack.powers)-1]
		stack.powers = stack.powers[:len(stack.powers)-1]
	}
	return power, len(stack.powers)
}

func (stack *StackOfPowerUp) peek() *PowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	if len(stack.powers) > 0 {
		return stack.powers[len(stack.powers)-1]
	}
	return nil
}

func (stack *StackOfPowerUp) push(power *PowerUp) int {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	stack.powers = append(stack.powers, power)
	return len(stack.powers)
}

func addPowerToStack(player *Player, power *PowerUp) {
	if power == nil {
		return
	}
	powerCount := player.actions.spaceStack.push(power)
	if powerCount == 1 {
		sendSoundToPlayer(player, "power-up-space")
	}
	updateOne(spanPower(powerCount), player)
}

func (stack *StackOfPowerUp) count() int {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers)
}

// toctoa?
func (stack *StackOfPowerUp) hasPower() bool {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers) > 0
}

//////////////////////////////////////////////////////
// Boosts
func (player *Player) addBoosts(n int) int {
	player.actions.boostMutex.Lock()
	defer player.actions.boostMutex.Unlock()
	player.actions.boostCounter += n
	return player.actions.boostCounter
}

func (player *Player) addBoostsAndUpdate(n int) {
	boostCount := player.addBoosts(n)
	updateOne(spanBoosts(boostCount), player)
}

func decrementBoost(player *Player) int {
	player.actions.boostMutex.Lock()
	defer player.actions.boostMutex.Unlock()
	if player.actions.boostCounter > 0 {
		player.actions.boostCounter--
	}
	return player.actions.boostCounter
}

func (player *Player) getBoostCountSync() int {
	player.actions.boostMutex.Lock()
	defer player.actions.boostMutex.Unlock()
	return player.actions.boostCounter
}

func (player *Player) useBoost() bool {
	boostCount := decrementBoost(player)
	updateOne(spanBoosts(boostCount), player)
	return boostCount > 0
}
