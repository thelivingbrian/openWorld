package main

import (
	"container/heap"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var CAPACITY_PER_TEAM = 128

type World struct {
	db                  *DB
	config              *Configuration
	worldPlayers        map[string]*Player
	wPlayerMutex        sync.Mutex
	teamQuantities      map[string]int
	status              WorldStatus
	incomingPlayers     map[string]*LoginRequest
	incomingPlayerMutex sync.Mutex
	worldStages         map[string]*Stage
	wStageMutex         sync.Mutex
	leaderBoard         *LeaderBoard
}

type WorldStatus struct {
	sync.Mutex
	lastStatusCheck    time.Time
	fuchsiaPlayerCount int
	skyBluePlayerCount int
}

type WorldStatusDiv struct {
	ServerName   string
	DomainName   string
	FuchsiaCount int
	SkyBlueCount int
	Vacancy      bool
}

type LoginRequest struct {
	Token     string
	Record    PlayerRecord
	timestamp time.Time
}

func createGameWorld(db *DB, config *Configuration) *World {
	minimumKillstreak := Player{id: "HS-only", killstreak: 0} // Do somewhere else?
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: []*Player{&minimumKillstreak}, index: make(map[*Player]int)}}
	return &World{
		db:                  db,
		config:              config,
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

const ACCEPTABLE_LOG_IN_DELAY_SECONDS = 15

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

	for key, request := range world.incomingPlayers {
		if isOverNSecondsAgo(request.timestamp, ACCEPTABLE_LOG_IN_DELAY_SECONDS) {
			delete(world.incomingPlayers, key)
			continue
		}
		if key == token {
			delete(world.incomingPlayers, key)
			return request
		}

	}
	return nil
}

func isOverNSecondsAgo(t time.Time, n int) bool {
	return time.Since(t) > time.Duration(n)*time.Second
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
	if world.isLoggedInAlready(incoming.Record.Username) {
		errorMessage := fmt.Sprintf(unableToJoin, "You are already logged in.")
		conn.WriteMessage(websocket.TextMessage, []byte(errorMessage))
		logger.Warn().Msg("User attempting to log in but is logged in already: " + incoming.Record.Username)
		return nil
	}

	newPlayer := world.newPlayerFromRecord(incoming.Record, incoming.Token)
	if world.teamAtCapacity(newPlayer.getTeamNameSync()) {
		errorMessage := fmt.Sprintf(unableToJoin, "Your team is at capacity.")
		conn.WriteMessage(websocket.TextMessage, []byte(errorMessage))
		return nil
	}
	world.addPlayer(newPlayer)

	world.leaderBoard.mostDangerous.LockThenPush(newPlayer)
	// world.leaderBoard.mostDangerous.Lock()
	// world.leaderBoard.mostDangerous.Push(newPlayer)
	// world.leaderBoard.mostDangerous.Unlock()

	newPlayer.conn = conn
	go newPlayer.sendUpdates()
	stage := getStageFromStageName(newPlayer, incoming.Record.StageName)
	placePlayerOnStageAt(newPlayer, stage, incoming.Record.Y, incoming.Record.X)

	return newPlayer
}

var unableToJoin = `
<div id="main_view" hx-swap-oob="true">
	<b>
		<span>Unable to join.<br />
		%s</span><br />
		<a href="#" hx-get="/worlds" hx-target="#page"> Try again</a>
	</b>
</div>`

func (world *World) teamAtCapacity(teamName string) bool {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	count, _ := world.teamQuantities[teamName]
	if count >= CAPACITY_PER_TEAM {
		return true
	}
	return false
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
		health:                   record.Health,
		money:                    record.Money,
		killCount:                record.KillCount,
		deathCount:               record.DeathCount,
		goalsScored:              record.GoalsScored,
		hatList:                  SyncHatList{HatList: record.HatList},
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
func initiateLogout(player *Player) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	player.tangible = false

	logger.Info().Msg("initate logout: " + player.username)
	//   Add time delay to prevent rage quit ?
	removeFromTileAndStage(player)

	playersToLogout <- player

}

func completeLogout(player *Player) {
	player.updateRecord() // Should return error

	// new method
	// player.setKillStreakAndUpdate(0) // Don't update
	player.setKillStreak(0)
	player.world.leaderBoard.mostDangerous.Update(player)
	player.world.leaderBoard.mostDangerous.Lock()
	index, exists := player.world.leaderBoard.mostDangerous.index[player]
	if exists {
		heap.Remove(&player.world.leaderBoard.mostDangerous, index)
	}
	player.world.leaderBoard.mostDangerous.Unlock()

	player.world.removePlayer(player)

	player.closeConnectionSync() // If Read deadline is missed conn may still be open
	// player.connLock.Lock()
	// player.conn = nil
	// player.connLock.Unlock()

	close(player.updates)
	// close(player.clearUpdateBuffer)

	logger.Info().Msg("Logout complete: " + player.username)

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

// Must hold locks before calling :

func (h *MaxStreakHeap) Len() int {
	return len(h.items)
}

func (h *MaxStreakHeap) Less(i, j int) bool {
	return h.items[i].getKillStreakSync() > h.items[j].getKillStreakSync()
}

func (h *MaxStreakHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i]], h.index[h.items[j]] = i, j
}

func (h *MaxStreakHeap) Push(x interface{}) {
	n := len(h.items)
	item := x.(*Player)
	h.items = append(h.items, item)
	h.index[h.items[n]] = n // would need fix if not at bottom. (e.g. richest)
}

func (h *MaxStreakHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	delete(h.index, item)
	return item
}

func (h *MaxStreakHeap) Peek() *Player {
	if len(h.items) == 0 {
		return nil
	}
	return h.items[0]
}

// Higher level interfaces:

func (h *MaxStreakHeap) LockThenPush(player *Player) {
	h.Lock()
	defer h.Unlock()
	h.Push(player)
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
		crownMostDangerous(currentMostDangerous)
	}
}

func crownMostDangerous(player *Player) {
	player.addHatByName("most-dangerous")
	notifyChangeInMostDangerous(player)
}

func (h *MaxStreakHeap) RemoveAndNotifyChange(player *Player) {
	h.Lock()
	defer h.Unlock()
	index, exists := player.world.leaderBoard.mostDangerous.index[player]
	if !exists {
		logger.Warn().Msg("Trying to remove player from mostDangerous but player does not exist!")
		return
	}
	heap.Remove(&player.world.leaderBoard.mostDangerous, index)
	if index == 0 {
		crownMostDangerous(h.Peek())
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
			update := fmt.Sprintf("@[%s|%s] has become the most dangerous bloop! (Streak: @[%d|red])", currentMostDangerous.username, currentMostDangerous.getTeamNameSync(), currentMostDangerous.getKillStreakSync())
			p.updateBottomText(update)
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
