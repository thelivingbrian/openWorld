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
	tilesToHighlight := make([]*Tile, 0, len(player.actions.spaceHighlights))

	playerHighlights := highlightMapToSlice(player)
	for _, tile := range playerHighlights {
		tile.damageAll(50, player)
		destroyFragileInteractable(tile, player)
		tile.eventsInFlight.Add(1)
		tilesToHighlight = append(tilesToHighlight, tile)

		go tile.tryToNotifyAfter(100)
	}
	damageBoxes := sliceOfTileToWeatherBoxes(tilesToHighlight, randomFieryColor())
	player.stage.updateAll(damageBoxes)
	updateOne(sliceOfTileToHighlightBoxes(tilesToHighlight, ""), player)

	player.actions.spaceHighlights = map[*Tile]bool{}
	if player.actions.spaceStack.hasPower() {
		player.nextPower()
	}
}

func (player *Player) nextPower() {
	player.actions.spaceStack.pop() // Throw old power away
	player.setSpaceHighlights()
	updateOne(sliceOfTileToHighlightBoxes(highlightMapToSlice(player), spaceHighlighter()), player)
}

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

func (stack *StackOfPowerUp) pop() *PowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	if len(stack.powers) > 0 {
		out := stack.powers[len(stack.powers)-1]
		stack.powers = stack.powers[:len(stack.powers)-1]
		return out
	}
	return nil // Should be impossible but return default power instead?
}

func (stack *StackOfPowerUp) peek() *PowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	if len(stack.powers) > 0 {
		return stack.powers[len(stack.powers)-1]
	}
	return nil
}

func (stack *StackOfPowerUp) push(power *PowerUp) *StackOfPowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	stack.powers = append(stack.powers, power)
	return stack
}

//////////////////////////////////////////////////////
// Boosts

func (player *Player) addBoosts(n int) {
	// not thread-safe
	player.actions.boostCounter += n
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) useBoost() bool {
	player.actions.boostMutex.Lock()
	defer player.actions.boostMutex.Unlock()
	if player.actions.boostCounter > 0 {
		player.actions.boostCounter--
	}
	updateOne(divPlayerInformation(player), player)
	return player.actions.boostCounter > 0
}

func (stack *StackOfPowerUp) hasPower() bool {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers) > 0
}
