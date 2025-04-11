package main

import (
	"container/heap"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const SESSION_SNAPSHOT_INTERVAL_IN_MIN = 30

var CAPACITY_PER_TEAM = 128 // Is modified by test => not const

type World struct {
	db                  *DB
	config              *Configuration
	worldPlayers        map[string]*Player
	wPlayerMutex        sync.Mutex
	teamQuantities      map[string]int
	teamPlayerStatus    TeamPlayerStatus
	incomingPlayers     map[string]*LoginRequest
	incomingPlayerMutex sync.Mutex
	playersToLogout     chan *Player
	worldStages         map[string]*Stage
	wStageMutex         sync.Mutex
	leaderBoard         *LeaderBoard
	sessionStats        *WorldSessionData
}

type TeamPlayerStatus struct {
	sync.Mutex
	lastStatusCheck    time.Time
	fuchsiaPlayerCount int
	skyBluePlayerCount int
}

type LoginRequest struct {
	Token     string
	Record    PlayerRecord
	timestamp time.Time
}

type WorldSessionData struct {
	sessionStartTime       time.Time
	peakSessionPlayerCount atomic.Int64
	peakSessionKillStreak  atomic.Int64
	peakSessionKiller      string
	TotalSessionLogins     atomic.Int64
	TotalSessionLogouts    atomic.Int64
}

//////////////////////////////////////////////////////////////////
// Create World

func createGameWorld(db *DB, config *Configuration) *World {
	out := &World{
		db:                  db,
		config:              config,
		worldPlayers:        make(map[string]*Player),
		wPlayerMutex:        sync.Mutex{},
		teamQuantities:      map[string]int{},
		incomingPlayers:     make(map[string]*LoginRequest),
		incomingPlayerMutex: sync.Mutex{},
		playersToLogout:     make(chan *Player, 0),
		worldStages:         make(map[string]*Stage),
		wStageMutex:         sync.Mutex{},
		leaderBoard:         createLeaderBoard(),
		sessionStats: &WorldSessionData{
			sessionStartTime: time.Now(),
		},
	}
	if config.loadPreviousState {
		loadPreviousState(out)
	}
	go processMostDangerous(out, &out.leaderBoard.mostDangerous)
	go processLogouts(out.playersToLogout)
	return out
}

func createLeaderBoard() *LeaderBoard {
	lb := &LeaderBoard{mostDangerous: MaxStreakHeap{items: make([]PlayerStreakRecord, 0), index: make(map[string]int), incoming: make(chan PlayerStreakRecord)}}
	return lb
}

func loadPreviousState(world *World) {
	lastStatus, err := getMostRecentSessionData(context.TODO(), world.db.sessionData, world.config.serverName)
	if err != nil {
		logger.Error().Msg("Error: " + err.Error())
	}
	if lastStatus == nil {
		return
	}
	world.leaderBoard.scoreboard.Import(lastStatus.Scoreboard)
}

///////////////////////////////////////////////////////////////
// World Session State

func periodicSnapshot(world *World) {
	ticker := time.NewTicker(time.Duration(SESSION_SNAPSHOT_INTERVAL_IN_MIN) * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		saveCurrentStatus(world)
	}
}

func saveCurrentStatus(world *World) {
	status := SessionDataRecord{
		ServerName:             world.config.serverName,
		Timestamp:              time.Now(),
		SessionStartTime:       world.sessionStats.sessionStartTime,
		PeakSessionPlayerCount: int(world.sessionStats.peakSessionPlayerCount.Load()),
		PeakSessionKillSteak: SessionStreakRecord{
			Streak:     int(world.sessionStats.peakSessionKillStreak.Load()),
			PlayerName: world.sessionStats.peakSessionKiller,
		},
		TotalSessionLogins:     int(world.sessionStats.TotalSessionLogins.Load()),
		TotalSessionLogouts:    int(world.sessionStats.TotalSessionLogouts.Load()),
		CurrentTeamPlayerCount: CopyTeamQuantities(world),
		Scoreboard:             world.leaderBoard.scoreboard.Export(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := saveGameStatus(ctx, world.db.sessionData, status); err != nil {
		logger.Error().Msg(fmt.Sprintf("failed to save current game status: %v", err))
	}
}

func incrementSessionLogins(w *World) {
	w.sessionStats.TotalSessionLogins.Add(1)
}

func incrementSessionLogouts(w *World) {
	w.sessionStats.TotalSessionLogouts.Add(1)
}

func trySetPeakPlayerCount(w *World, count int) bool {
	return SetMaxAtomicIfGreater(&w.sessionStats.peakSessionPlayerCount, count)
}

func trySetPeakKillStreak(w *World, streak int) bool {
	return SetMaxAtomicIfGreater(&w.sessionStats.peakSessionKillStreak, streak)
}

func SetMaxAtomicIfGreater(atom *atomic.Int64, newValue int) bool {
	// CAS loop for atomicity
	for {
		current := atom.Load()
		if int(current) >= newValue {
			return false
		}
		if atom.CompareAndSwap(current, int64(newValue)) {
			return true
		}
	}
}

func CopyTeamQuantities(world *World) map[string]int {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	newMap := make(map[string]int, len(world.teamQuantities))
	for key, value := range world.teamQuantities {
		newMap[key] = value
	}
	return newMap
}

/////////////////////////////////////////////////////
// Add / Remove / Find Players

func (world *World) addPlayer(p *Player) int {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	// Update team count
	previousCount := world.teamQuantities[p.team]
	world.teamQuantities[p.team] = previousCount + 1
	// Update world players
	world.worldPlayers[p.id] = p
	return len(world.worldPlayers)
}

func (world *World) removePlayer(p *Player) {
	world.wPlayerMutex.Lock()
	delete(world.worldPlayers, p.id)
	previousCount := world.teamQuantities[p.team]
	world.teamQuantities[p.team] = previousCount - 1
	world.wPlayerMutex.Unlock()
}

func (world *World) getPlayerById(id string) *Player {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	player := world.worldPlayers[id]
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
		sendUnableToJoinMessage(conn, "You are already logged in.")
		logger.Warn().Msg("User attempting to log in but is logged in already: " + incoming.Record.Username)
		return nil
	}

	newPlayer := world.newPlayerFromRecord(incoming.Record, incoming.Token)
	if world.teamAtCapacity(newPlayer.getTeamNameSync()) {
		sendUnableToJoinMessage(conn, "Your team is at capacity.")
		return nil
	}

	newPlayer.updateRecordOnLogin()
	stage := getStageFromStageName(newPlayer, incoming.Record.StageName)
	// if stage == nil {
	// 	// Technically impossible because of clinic -> panic failsafe ?
	// 	logger.Warn().Msg("WARN: Player " + newPlayer.username + " on unloadable stage: " + incoming.Record.StageName)
	// 	return nil
	// }

	emptyScreen := emptyScreenForStage(stage)
	if !sendInitialScreen(conn, emptyScreen) {
		return nil
	}

	newPlayer.conn = conn
	go newPlayer.sendUpdates()

	// correct order?
	count := world.addPlayer(newPlayer)
	trySetPeakPlayerCount(world, count)

	placePlayerOnStageAt(newPlayer, stage, incoming.Record.Y, incoming.Record.X)
	return newPlayer
}

const unableToJoin = `
<div id="main_view" hx-swap-oob="true">
	<b>
		<span>Unable to join.<br />
		%s</span><br />
		<a href="#" hx-get="/worlds" hx-target="#page"> Try again</a>
	</b>
</div>`

func sendUnableToJoinMessage(conn WebsocketConnection, description string) {
	errorMessage := fmt.Sprintf(unableToJoin, description)
	conn.WriteMessage(websocket.TextMessage, []byte(errorMessage))
}

func sendInitialScreen(conn WebsocketConnection, screen []byte) bool {
	err := conn.SetWriteDeadline(time.Now().Add(4000 * time.Millisecond))
	if err != nil {
		logger.Warn().Msg("Failed to set deadline on join")
		return false
	}
	err = conn.WriteMessage(websocket.TextMessage, screen)
	if err != nil {
		logger.Warn().Msg("Failed to write message on join")
		return false
	}
	return true
}

func (world *World) teamAtCapacity(teamName string) bool {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	count := world.teamQuantities[teamName]
	return count >= CAPACITY_PER_TEAM
}

func (world *World) newPlayerFromRecord(record PlayerRecord, id string) *Player {
	// probably take this out later...
	if record.Team == "" {
		record.Team = "sky-blue"
	}
	updatesForPlayer := make(chan []byte) // raise capacity?
	newPlayer := &Player{
		id:                       id,
		username:                 record.Username,
		stage:                    nil,
		updates:                  updatesForPlayer,
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

	player.world.playersToLogout <- player

}

func completeLogout(player *Player) {
	player.updateRecordOnLogout() // Should return error

	player.world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{id: player.id, username: player.username, killstreak: 0, team: ""}
	player.world.removePlayer(player)
	incrementSessionLogouts(player.world)

	player.closeConnectionSync() // If Read deadline is missed conn may still be open
	close(player.updates)

	logger.Info().Msg("Logout complete: " + player.username)
}

///////////////////////////////////////////////////////////////
// References / Lookup

func getRelativeTile(source *Tile, yOff, xOff int, player *Player) *Tile {
	destY := source.y + yOff
	destX := source.x + xOff
	if validCoordinate(destY, destX, source.stage) {
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

			if validCoordinate(mod(destY, len(newStage.tiles)), destX, newStage) {
				return newStage.tiles[mod(destY, len(newStage.tiles))][destX]
			}
		}
		if escapesHorizontally {
			var newStage *Stage
			if xOff > 0 {
				newStage = player.fetchStageSync(source.stage.east)
			}
			if xOff < 0 {
				newStage = player.fetchStageSync(source.stage.west)
			}

			if validCoordinate(destY, mod(destX, len(newStage.tiles)), newStage) {
				return newStage.tiles[destY][mod(destX, len(newStage.tiles))]
			}
		}

		return nil
	}
}

///////////////////////////////////////////////////////////////
// LeaderBoards

type LeaderBoard struct {
	mostDangerous MaxStreakHeap
	scoreboard    Scoreboard
}

type Scoreboard struct {
	data sync.Map
}

func (s *Scoreboard) Increment(team string) int {
	val, _ := s.data.LoadOrStore(team, &atomic.Int64{})
	score := val.(*atomic.Int64)

	score.Add(1)
	return int(score.Load())
}

func (s *Scoreboard) ResetAll() {
	s.data.Range(func(key, value interface{}) bool {
		if score, ok := value.(*atomic.Int64); ok {
			score.Store(0)
		}
		return true
	})
}

func (s *Scoreboard) GetScore(team string) int {
	val, ok := s.data.Load(team)
	if !ok {
		return 0 // Team does not exist
	}

	score := val.(*atomic.Int64)
	return int(score.Load())
}

func (s *Scoreboard) Export() map[string]int {
	result := make(map[string]int)
	s.data.Range(func(key, value interface{}) bool {
		team, ok := key.(string)
		if !ok {
			return true
		}
		score, ok := value.(*atomic.Int64)
		if !ok {
			return true
		}
		result[team] = int(score.Load())
		return true
	})
	return result
}

func (s *Scoreboard) Import(data map[string]int) {
	// Existing Data is cleared.
	s.data.Range(func(key, _ interface{}) bool {
		s.data.Delete(key)
		return true
	})
	// Import new data.
	for team, score := range data {
		scoreVal := &atomic.Int64{}
		scoreVal.Store(int64(score))
		s.data.Store(team, scoreVal)
	}
}

//////////////////////////////////////////////////////////////////////////
// Most Dangerous Heap

type MaxStreakHeap struct {
	items    []PlayerStreakRecord
	index    map[string]int // Keep track of indices of record by id
	incoming chan PlayerStreakRecord
}

type PlayerStreakRecord struct {
	id         string
	killstreak int
	username   string
	team       string
}

func (heap *MaxStreakHeap) Len() int {
	return len(heap.items)
}

func (heap *MaxStreakHeap) Less(i, j int) bool {
	return heap.items[i].killstreak > heap.items[j].killstreak
}

func (heap *MaxStreakHeap) Swap(i, j int) {
	heap.items[i], heap.items[j] = heap.items[j], heap.items[i]
	heap.index[heap.items[i].id], heap.index[heap.items[j].id] = i, j
}

func (h *MaxStreakHeap) Push(x interface{}) {
	n := len(h.items)
	item := x.(PlayerStreakRecord)
	h.items = append(h.items, item)
	h.index[h.items[n].id] = n
	heap.Fix(h, n)
}

func (heap *MaxStreakHeap) Pop() interface{} {
	old := heap.items
	n := len(old)
	item := old[n-1]
	heap.items = old[0 : n-1]
	delete(heap.index, item.id)
	return item
}

func (heap *MaxStreakHeap) Peek() PlayerStreakRecord {
	if len(heap.items) == 0 {
		return PlayerStreakRecord{}
	}
	return heap.items[0]
}

//

func processMostDangerous(world *World, h *MaxStreakHeap) {
	for {
		previousMostDangerous := h.Peek()
		event, ok := <-h.incoming
		if !ok {
			logger.Warn().Msg("Stopping Processing for High-KillStreak Heap - incoming closed")
			break
		}
		position, found := h.index[event.id]
		if !found {
			if event.killstreak != 0 {
				heap.Push(h, event)
			}
		} else {
			if event.killstreak != 0 {
				h.items[position] = event
				heap.Fix(h, position)
			} else {
				heap.Remove(h, position)
			}
		}
		currentMostDangerous := h.Peek()

		if trySetPeakKillStreak(world, currentMostDangerous.killstreak) {
			world.sessionStats.peakSessionKiller = currentMostDangerous.username
		}

		func(currentMostDangerous, previousMostDangerous PlayerStreakRecord) {
			if currentMostDangerous.id != previousMostDangerous.id {
				// new routine prevents deadlock from tangible check (removed)
				crownMostDangerousById(world, currentMostDangerous)
				return
			}
			// current and previous are the same
			if currentMostDangerous.killstreak != previousMostDangerous.killstreak {
				update := fmt.Sprintf("@[%s|%s] is still the most dangerous bloop! (Streak: @[%d|red])", currentMostDangerous.username, currentMostDangerous.team, currentMostDangerous.killstreak)
				broadcastBottomText(world, update)
			}
		}(currentMostDangerous, previousMostDangerous)
	}
}

func crownMostDangerousById(world *World, streakEvent PlayerStreakRecord) {
	if streakEvent.id == "" {
		return
	}
	player := world.getPlayerById(streakEvent.id)
	if player == nil {
		return
	}
	player.addHatByName("most-dangerous")
	player.world.notifyChangeInMostDangerous(streakEvent)
}

////////////////////////////////////////////////////////////
// World Player bottom text updates

func (world *World) notifyChangeInMostDangerous(streakEvent PlayerStreakRecord) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, p := range world.worldPlayers {
		if p.id == streakEvent.id {
			p.updateBottomText("You are the most dangerous bloop!")
		} else {
			template := "@[%s|%s] has become the most dangerous bloop! (Streak: @[%d|red])"
			update := fmt.Sprintf(template, streakEvent.username, streakEvent.team, streakEvent.killstreak)
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

func broadcastUpdate(world *World, message string) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, p := range world.worldPlayers {
		updateOne(message, p)
	}
}

func awardHatByTeam(world *World, team, hat string) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, p := range world.worldPlayers {
		if p.getTeamNameSync() == team {
			p.addHatByName(hat)
		}
	}
}
