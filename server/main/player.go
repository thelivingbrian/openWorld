package main

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	id                       string
	username                 string
	team                     string
	icon                     string
	viewLock                 sync.Mutex
	world                    *World
	stage                    *Stage // Shouldn't exist except for tile?
	stageLock                sync.Mutex
	tile                     *Tile
	tileLock                 sync.Mutex
	updates                  chan []byte
	sessionTimeOutViolations atomic.Int32
	conn                     WebsocketConnection
	connLock                 sync.RWMutex
	tangible                 bool
	tangibilityLock          sync.Mutex
	health                   int
	healthLock               sync.Mutex
	money                    int
	moneyLock                sync.Mutex
	killCount                int
	killCountLock            sync.Mutex
	deathCount               int
	deathCountLock           sync.Mutex
	goalsScored              int
	goalsScoredLock          sync.Mutex
	killstreak               int
	streakLock               sync.Mutex
	actions                  *Actions
	menues                   map[string]Menu
	playerStages             map[string]*Stage
	pStageMutex              sync.Mutex
	hatList                  SyncHatList
}

type WebsocketConnection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	SetWriteDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
}

////////////////////////////////////////////////////////////
//  Special Movement

func (player *Player) moveNorthBoost() {
	player.moveBoost(-1, 0)
}

func (player *Player) moveSouthBoost() {
	player.moveBoost(1, 0)
}

func (player *Player) moveEastBoost() {
	player.moveBoost(0, 1)
}

func (player *Player) moveWestBoost() {
	player.moveBoost(0, -1)
}

func (player *Player) moveBoost(yOffset int, xOffset int) {
	if player.useBoost() {
		move(player, 2*yOffset, 2*xOffset)
	} else {
		move(player, yOffset, xOffset)
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	applyTeleport(player, teleport)
	sendSoundToPlayer(player, "teleport")
}

///////////////////////////////////////////////////////////////////////
// Death

func handleDeath(player *Player) {
	popAndDropMoney(player)
	removeFromTileAndStage(player) // After this should be impossible for any transfer to succeed
	player.incrementDeathCount()
	player.setHealth(150)
	player.setKillStreak(0)
	player.actions = createDefaultActions() // problematic, -> setDefaultActions(player)

	stage := player.fetchStageSync(infirmaryStagenameForPlayer(player))
	player.setStage(stage)
	player.updateRecordOnDeath(stage.tiles[2][2])
	respawnOnStage(player, stage)
}

func popAndDropMoney(player *Player) {
	tile := player.getTileSync()

	playerLostMoney := halveMoneyOf(player)
	moneyToAdd := max(playerLostMoney, 10)
	tile.addMoneyAndNotifyAllExcept(moneyToAdd, player)

	pop := soundTriggerByName("pop-death")
	tile.stage.updateAllExcept(pop, player)
}

func halveMoneyOf(player *Player) int {
	// race risk
	currentMoney := player.getMoneySync()
	newValue := currentMoney / 2
	player.setMoneyAndUpdate(newValue)
	return newValue
}

func respawnOnStage(player *Player, stage *Stage) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}

	placePlayerOnStageAt(player, stage, 2, 2)
	sendSoundToPlayer(player, soundTriggerByName("pop-death"))
	player.updatePlayerHud()
	player.updateBottomText("You have died.")
}

func removeFromTileAndStage(player *Player) {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	if player.stage == nil || player.tile == nil {
		return
	}
	player.tile.removePlayerAndNotifyOthers(player)
	player.stage.removeLockedPlayerById(player.id)
}

func infirmaryStagenameForPlayer(player *Player) string {
	team := player.getTeamNameSync()
	if team != "sky-blue" && team != "fuchsia" {
		return "clinic"
	}
	longitude := strconv.Itoa(rand.IntN(4))
	latitude := ""
	if team == "fuchsia" {
		latitude = "0"
	}
	if team == "sky-blue" {
		latitude = "3"
	}
	return fmt.Sprintf("infirmary:%s-%s", latitude, longitude)
}

////////////////////////////////////////////////////////////
//	Updates

func (player *Player) sendUpdates() {
	var buffer bytes.Buffer
	const maxBufferSize = 10 * 256 * 1024

	shouldSendUpdates := true
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case update, ok := <-player.updates:
			if !ok {
				logger.Info().Msg("Player:" + player.username + "- update channel closed")
				return
			}
			if !shouldSendUpdates {
				continue
			}

			if buffer.Len()+len(update) < maxBufferSize {
				// Accumulate the update in the buffer.
				buffer.Write(update)
			} else {
				logger.Warn().Msg(fmt.Sprintf("Player: %s - buffer exceeded %d bytes, wiping buffer\n", player.username, maxBufferSize))
				buffer.Reset()
			}
		case <-ticker.C:
			if !shouldSendUpdates || buffer.Len() == 0 {
				continue
			}
			// Every 25ms, if there's anything in the buffer, send it.
			err := sendUpdate(player, buffer.Bytes())
			if err != nil {
				//logger.Warn().Err(err).Msg("Error - Stopping furture sends: ")
				shouldSendUpdates = false
				player.closeConnectionSync()
			}

			buffer.Reset()
		}
	}
}

func sendUpdate(player *Player, update []byte) error {
	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn == nil {
		return errors.New("Connection is expired for: " + player.username)
	}

	err := player.conn.SetWriteDeadline(time.Now().Add(2000 * time.Millisecond))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to set write deadline:")
		return err
	}
	err = player.conn.WriteMessage(websocket.TextMessage, update)
	if err != nil {
		// Technically can be any write error
		logger.Debug().Msg("Incrementing websocket session timeout violations for: " + player.username)
		if player.sessionTimeOutViolations.Add(1) >= 2 {
			return err
		}
	}

	return nil
}

// Updates - Enqueue
func updateOthersAfterMovement(player *Player, current, previous *Tile) {
	previous.stage.updateAllExcept(characterBox(previous), player)
	current.stage.updateAllExcept(characterBox(current), player)
}

func updateAllAfterMovement(current, previous *Tile) {
	previous.stage.updateAll(characterBox(previous))
	current.stage.updateAll(characterBox(current))
}

func updatePlayerAfterMovement(player *Player, current, previous *Tile) {
	impactedHighlights := player.updateSpaceHighlights()

	playerIcon := playerBoxSpecifc(current.y, current.x, player.getIconSync())

	previousBoxes := ""
	if previous != nil && previous.stage == player.getStageSync() {
		previousBoxes += characterBox(previous)
	}

	player.updates <- []byte(highlightBoxesForPlayer(player, impactedHighlights) + previousBoxes + playerIcon)
}

func updatePlayerAfterStageChange(p *Player) {
	p.setSpaceHighlights()
	updateEntireExistingScreen(p)
}

func updateEntireExistingScreen(player *Player) {
	player.updates <- entireScreenAsSwaps(player)
}

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(`
			<div id="bottom_text">
				&nbsp;&nbsp;> %s
			</div>`, processStringForColors(message))
	updateOne(msg, player)
}

func (player *Player) updatePlayerHud() {
	player.updatePlayerBox()
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) updatePlayerBox() {
	icon := player.setIcon()
	tile := player.getTileSync()
	updateOne(playerBoxSpecifc(tile.y, tile.x, icon), player)
}

func updateIconForAll(player *Player) {
	player.setIcon()
	tile := player.getTileSync()
	tile.stage.updateAll(characterBox(tile))
}

func updateIconForAllIfTangible(player *Player) {
	player.setIcon()
	ownLock := player.tangibilityLock.TryLock()
	if !ownLock || !player.tangible {
		return
	}
	defer player.tangibilityLock.Unlock()
	tile := player.getTileSync()
	tile.stage.updateAll(characterBox(tile))
}

func sendSoundToPlayer(player *Player, soundName string) {
	updateOne(soundTriggerByName(soundName), player)
}

func soundTriggerByName(soundName string) string {
	return fmt.Sprintf(`<div id="sound">%s</div>`, soundName)

}

// chan Update

func updateOne(update string, player *Player) {
	player.updates <- []byte(update)
}

// Database update

func (player *Player) updateRecord() {
	currentTile := player.getTileSync()
	go player.world.db.updateRecordForPlayer(player, currentTile)
}

func (player *Player) updateRecordOnDeath(respawnTile *Tile) {
	// vs just incrementing death count?
	go player.world.db.updateRecordForPlayer(player, respawnTile)
}

func (player *Player) updateRecordOnLogin() {
	go player.world.db.updateLoginForPlayer(player)
}
func (player *Player) updateRecordOnLogout() {
	currentTile := player.getTileSync()
	go player.world.db.updatePlayerRecordOnLogout(player, currentTile)
}

/////////////////////////////////////////////////////////////
// Stages

func getStageByNameOrGetDefault(player *Player, stagename string) *Stage {
	// Never returns nil - Used on login as protection against old records with nonexistant locations
	stage := player.fetchStageSync(stagename)
	if stage == nil {
		logger.Warn().Msg("WARNING: Fetching default stage instead of: " + stagename)
		// clinic must exist - Add test ?
		stage = player.fetchStageSync("clinic")
		if stage == nil {
			panic("Default stage not found")
		}
	}

	return stage
}

/////////////////////////////////////////////////////////////
//  Hats

func (player *Player) addHatByName(hatName string) {
	hat := player.hatList.addByName(hatName)
	if hat == nil {
		return
	}
	logger.Debug().Msg("Adding Hat: " + hat.Name)
	player.world.db.addHatToPlayer(player.username, *hat)
	updateIconForAllIfTangible(player) // May not originate from click hence check tangible
}

func (player *Player) cycleHats() {
	player.hatList.nextValid()
	updateIconForAll(player)
}

/////////////////////////////////////////////////////////////
// Observers

// Does not handle death
func (player *Player) setHealth(n int) {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	player.health = n
}

func (player *Player) getHealthSync() int {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	return player.health
}

// Icon Observer, note that health can not be locked
func (player *Player) setIcon() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	if player.health <= 50 {
		player.icon = "dim-" + player.team + " " + player.hatList.currentTrim() + " r0"
		return player.icon
	} else {
		player.icon = player.team + " " + player.hatList.currentTrim() + " r0"
		return player.icon
	}
}

func (player *Player) getTeamNameSync() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	return player.team
}

// Remove ?
func (player *Player) setStage(stage *Stage) {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.stage = stage
}

// Use Tile ?
func (player *Player) getStageNameSync() string {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stage.name
}

func (player *Player) getStageSync() *Stage {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stage
}

func (player *Player) setMoney(n int) {
	player.moneyLock.Lock()
	defer player.moneyLock.Unlock()
	player.money = n
}

func (player *Player) addMoney(n int) int {
	player.moneyLock.Lock()
	defer player.moneyLock.Unlock()
	player.money += n
	return player.money
}

func (player *Player) setMoneyAndUpdate(n int) {
	player.setMoney(n)
	updateOne(spanMoney(n), player)
}

func (player *Player) addMoneyAndUpdate(n int) {
	totalMoney := player.addMoney(n)
	if totalMoney > 2*1000 {
		player.addHatByName("made-of-money")
	}
	if totalMoney > 100*1000 {
		player.addHatByName("made-of-money-2")
	}
	updateOne(spanMoney(totalMoney), player)
}

func (player *Player) getMoneySync() int {
	player.moneyLock.Lock()
	defer player.moneyLock.Unlock()
	return player.money
}

func (player *Player) getKillStreakSync() int {
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	return player.killstreak
}

func (player *Player) setKillStreak(n int) int {
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	player.killstreak = n
	player.world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{id: player.id, username: player.username, killstreak: n, team: player.getTeamNameSync()}
	return player.killstreak
}

func (player *Player) incrementKillStreak() int {
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	player.killstreak++
	player.world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{id: player.id, username: player.username, killstreak: player.killstreak, team: player.getTeamNameSync()}
	return player.killstreak
}

func (player *Player) getKillCountSync() int {
	player.killCountLock.Lock()
	defer player.killCountLock.Unlock()
	return player.killCount
}

// killCount Observer - no direct set
func (player *Player) incrementKillCount() {
	player.killCountLock.Lock()
	defer player.killCountLock.Unlock()
	player.killCount++
}

func (player *Player) getDeathCountSync() int {
	player.deathCountLock.Lock()
	defer player.deathCountLock.Unlock()
	return player.deathCount
}

// deathCount Observer - no direct set
func (player *Player) incrementDeathCount() {
	player.deathCountLock.Lock()
	defer player.deathCountLock.Unlock()
	player.deathCount++
}

// goals observer no direct set
func (player *Player) incrementGoalsScored() int {
	player.goalsScoredLock.Lock()
	defer player.goalsScoredLock.Unlock()
	player.goalsScored++
	return player.goalsScored
}

func (player *Player) getGoalsScored() int {
	player.goalsScoredLock.Lock()
	defer player.goalsScoredLock.Unlock()
	return player.goalsScored
}

// generally will trigger a logout
func (player *Player) closeConnectionSync() error {
	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn == nil {
		return errors.New("player connection nil before attempted close")
	}
	return player.conn.Close()
}
