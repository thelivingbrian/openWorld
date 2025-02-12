package main

import (
	"container/heap"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type World struct {
	db                  *DB
	worldPlayers        map[string]*Player
	wPlayerMutex        sync.Mutex
	teamQuantities      map[string]int
	incomingPlayers     map[string]*LoginRequest
	incomingPlayerMutex sync.Mutex
	worldStages         map[string]*Stage
	wStageMutex         sync.Mutex
	leaderBoard         *LeaderBoard
}

type LoginRequest struct {
	Token     string
	Record    PlayerRecord
	timestamp time.Time
}

func createGameWorld(db *DB) *World {
	minimumKillstreak := Player{id: "HS-only", killstreak: 0} // Do somewhere else?
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: []*Player{&minimumKillstreak}, index: make(map[*Player]int)}}
	return &World{
		db:                  db,
		worldPlayers:        make(map[string]*Player),
		wPlayerMutex:        sync.Mutex{},
		teamQuantities:      map[string]int{},
		incomingPlayers:     make(map[string]*LoginRequest),
		incomingPlayerMutex: sync.Mutex{},
		worldStages:         make(map[string]*Stage),
		wStageMutex:         sync.Mutex{},
		leaderBoard:         lb,
	}
}

func (world *World) addPlayer(p *Player) {
	world.wPlayerMutex.Lock()
	world.worldPlayers[p.id] = p
	previousCount := world.teamQuantities[p.team]
	world.teamQuantities[p.team] = previousCount + 1
	world.wPlayerMutex.Unlock()
}

func (world *World) removePlayer(p *Player) {
	world.wPlayerMutex.Lock()
	delete(world.worldPlayers, p.id)
	previousCount := world.teamQuantities[p.team]
	world.teamQuantities[p.team] = previousCount - 1
	world.wPlayerMutex.Unlock()
}

func (world *World) getPlayerByName(id string) *Player {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	player, _ := world.worldPlayers[id]
	return player
}

//////////////////////////////////////////////////
//  Log in

func createLoginRequest(record PlayerRecord) *LoginRequest {
	return &LoginRequest{
		Token:     createRandomToken(),
		Record:    record,
		timestamp: time.Now(),
	}
}

func (world *World) addIncoming(loginRequest *LoginRequest) {
	world.incomingPlayerMutex.Lock()
	defer world.incomingPlayerMutex.Unlock()
	world.incomingPlayers[loginRequest.Token] = loginRequest
}

func (world *World) retreiveIncoming(token string) *LoginRequest {
	world.incomingPlayerMutex.Lock()
	defer world.incomingPlayerMutex.Unlock()
	// Can leak spammed unattempted tokens
	// Use list and iterate fully instead?
	request, ok := world.incomingPlayers[token]
	if ok {
		delete(world.incomingPlayers, token)
		if isLessThan15SecondsAgo(request.timestamp) {
			return request
		}
	}
	return nil
}

func isLessThan15SecondsAgo(t time.Time) bool {
	if time.Since(t) < 0 {
		// t is in the future
		return false
	}
	return time.Since(t) < 450*time.Second
}

func createRandomToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(token)
}

func (world *World) join(incoming *LoginRequest, conn WebsocketConnection) *Player {
	// need log levels
	//fmt.Println("New Player: " + record.Username)
	//fmt.Println("Token: " + token)

	if world.isLoggedInAlready(incoming.Record.Username) {
		fmt.Println("User attempting to log in but is logged in already: " + incoming.Record.Username)
		return nil
	}

	newPlayer := world.newPlayerFromRecord(incoming.Record, incoming.Token)
	world.addPlayer(newPlayer)

	world.leaderBoard.mostDangerous.Lock()
	world.leaderBoard.mostDangerous.Push(newPlayer)
	world.leaderBoard.mostDangerous.Unlock()

	newPlayer.conn = conn
	go newPlayer.sendUpdates()
	stage := getStageFromStageName(newPlayer, incoming.Record.StageName)
	placePlayerOnStageAt(newPlayer, stage, incoming.Record.Y, incoming.Record.X)

	return newPlayer
}

func (world *World) newPlayerFromRecord(record PlayerRecord, id string) *Player {
	// probably take this out later...
	if record.Team == "" {
		record.Team = "sky-blue"
	}
	updatesForPlayer := make(chan []byte, 0) // raise capacity?
	newPlayer := &Player{
		id:                       id,
		username:                 record.Username,
		stage:                    nil,
		updates:                  updatesForPlayer,
		clearUpdateBuffer:        make(chan struct{}, 0),
		sessionTimeOutViolations: atomic.Int32{},
		tangible:                 true,
		tangibilityLock:          sync.Mutex{},
		actions:                  createDefaultActions(),
		world:                    world,
		menues:                   map[string]Menu{"pause": pauseMenu, "map": mapMenu, "stats": statsMenu, "respawn": respawnMenu}, // terrifying
		playerStages:             make(map[string]*Stage),
		team:                     record.Team,
		//trim:                     record.Trim,
		health:      record.Health,
		money:       record.Money,
		killCount:   record.KillCount,
		deathCount:  record.DeathCount,
		goalsScored: record.GoalsScored,
		hatList:     SyncHatList{HatList: record.HatList},
	}

	newPlayer.setIcon()
	return newPlayer
}

func (world *World) isLoggedInAlready(username string) bool {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, player := range world.worldPlayers {
		if player.username == username {
			return true
		}
	}
	return false
}

// ////////////////////////////////////////////////////
//
//	Logging Out

var playersToLogout = make(chan *Player, 500)

func processLogouts(players chan *Player) {
	for {
		player, ok := <-players
		if !ok {
			return
		}

		completeLogout(player)
	}
}
func initiatelogout(player *Player) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	player.tangible = false

	fmt.Println("initate logout: " + player.username)
	removeFromTileAndStage(player)

	playersToLogout <- player

}

func completeLogout(player *Player) {
	player.updateRecord() // Should return error

	// new method
	player.setKillStreakAndUpdate(0) // Don't update
	player.world.leaderBoard.mostDangerous.Lock()
	index, exists := player.world.leaderBoard.mostDangerous.index[player]
	if exists {
		heap.Remove(&player.world.leaderBoard.mostDangerous, index)
	}
	player.world.leaderBoard.mostDangerous.Unlock()

	player.world.removePlayer(player)

	player.closeConnectionSync() // uneeded but harmless?
	player.connLock.Lock()
	player.conn = nil
	player.connLock.Unlock()

	close(player.updates)
	// close(player.clearUpdateBuffer)

	fmt.Println("Logout complete: " + player.username)

}

func fullyRemovePlayer_do(player *Player) {
	removeFromTileAndStage(player)
}

func fullyRemovePlayer(player *Player) bool {
	found := false
	for i := 0; i < 5; i++ {
		if player.tile.removePlayerAndNotifyOthers(player) {
			found = true
			break
		}
	}

	if !found {
		fmt.Println("Never removed player from tile successfully")
	}

	player.stage.playerMutex.Lock()
	_, ok := player.stage.playerMap[player.id]
	delete(player.stage.playerMap, player.id)
	player.stage.playerMutex.Unlock()

	return found && ok
}

///////////////////////////////////////////////////////////////
// References / Lookup

func getRelativeTile(source *Tile, yOff, xOff int, player *Player) *Tile {
	destY := source.y + yOff
	destX := source.x + xOff
	if validCoordinate(destY, destX, source.stage.tiles) {
		return source.stage.tiles[destY][destX]
	} else {
		escapesVertically, escapesHorizontally := validityByAxis(destY, destX, source.stage.tiles)
		if escapesVertically && escapesHorizontally {
			// in bloop world cardinal direction travel may be non-communative
			// therefore north-east etc neighbor is not uniquely defined
			// order can probably be uniquely determined when tile.y != tile.x
			return nil
		}
		if escapesVertically {
			var newStage *Stage
			if yOff > 0 {
				newStage = player.fetchStageSync(source.stage.south)
			}
			if yOff < 0 {
				newStage = player.fetchStageSync(source.stage.north)
			}

			if newStage != nil {
				if validCoordinate(mod(destY, len(newStage.tiles)), destX, newStage.tiles) {
					return newStage.tiles[mod(destY, len(newStage.tiles))][destX]
				}
			}
			return nil
		}
		if escapesHorizontally {
			var newStage *Stage
			if xOff > 0 {
				newStage = player.fetchStageSync(source.stage.east)
			}
			if xOff < 0 {
				newStage = player.fetchStageSync(source.stage.west)
			}

			if newStage != nil {
				if validCoordinate(destY, mod(destX, len(newStage.tiles)), newStage.tiles) {
					return newStage.tiles[destY][mod(destX, len(newStage.tiles))]
				}
			}
			return nil
		}

		return nil
	}
}

///////////////////////////////////////////////////////////////
// LeaderBoards

type LeaderBoard struct {
	//richest *Player
	mostDangerous MaxStreakHeap
	//oldest        *Player
	scoreboard Scoreboard
}

// Team Scoreboards
type TeamScore struct {
	sync.Mutex
	score int
}

type Scoreboard struct {
	data sync.Map
}

// Increment updates the score for a team atomically
func (s *Scoreboard) Increment(team string) {
	val, _ := s.data.LoadOrStore(team, &TeamScore{})
	teamScore := val.(*TeamScore)

	teamScore.Lock()
	teamScore.score += 1
	teamScore.Unlock()
}

func (s *Scoreboard) GetScore(team string) int {
	val, ok := s.data.Load(team)
	if !ok {
		return 0 // Team does not exist
	}

	teamScore := val.(*TeamScore)

	teamScore.Lock()
	defer teamScore.Unlock()
	return teamScore.score
}

//

type MaxStreakHeap struct {
	items []*Player
	index map[*Player]int // Keep track of item indices
	sync.Mutex
}

// must hold lock before calling
// have concerns here...
func (h *MaxStreakHeap) Len() int {
	//h.Lock()
	//defer h.Unlock()
	return len(h.items)
}

func (h *MaxStreakHeap) Less(i, j int) bool {
	//h.Lock()
	//defer h.Unlock()
	return h.items[i].getKillStreakSync() > h.items[j].getKillStreakSync()
}

func (h *MaxStreakHeap) Swap(i, j int) {
	//h.Lock()
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i]], h.index[h.items[j]] = i, j
	//h.Unlock()
}

func (h *MaxStreakHeap) Push(x interface{}) {
	//h.Lock()
	n := len(h.items)
	item := x.(*Player)
	h.items = append(h.items, item)
	h.index[h.items[n]] = n // would need fix if not at bottom. (e.g. richest)
	//h.Unlock()
}

func (h *MaxStreakHeap) Pop() interface{} {
	// h.Lock()
	// defer h.Unlock()
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	delete(h.index, item)
	return item
}

func (h *MaxStreakHeap) Peek() *Player {
	// h.Lock()
	// defer h.Unlock()
	if len(h.items) == 0 {
		return nil
		//panic("Heap Underflow")
	}
	return h.items[0]
}

// Update fixes the heap after player has a change in killstreak, notiying any change in most dangerous
func (h *MaxStreakHeap) Update(player *Player) {
	h.Lock()
	defer h.Unlock()
	previousMostDangerous := h.Peek()

	index := h.index[player]
	heap.Fix(h, index)

	currentMostDangerous := h.Peek()
	if currentMostDangerous != previousMostDangerous {
		currentMostDangerous.addHatByName("most-dangerous")
		notifyChangeInMostDangerous(currentMostDangerous)
	}
}

func notifyChangeInMostDangerous(currentMostDangerous *Player) {
	if currentMostDangerous.id == "HS-only" {
		return
	}
	currentMostDangerous.world.wPlayerMutex.Lock()
	defer currentMostDangerous.world.wPlayerMutex.Unlock()
	for _, p := range currentMostDangerous.world.worldPlayers {
		if p == currentMostDangerous {
			p.updateBottomText("You are the most dangerous bloop!")
		} else {
			p.updateBottomText(currentMostDangerous.username + " has become the most dangerous bloop!")
		}
	}
}

func broadcastBottomText(world *World, message string) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, p := range world.worldPlayers {
		p.updateBottomText(message)
	}
}
