package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type PlayerSocketEvent struct {
	Token    string `json:"token"`
	Name     string `json:"eventname"`
	MenuName string `json:"menuName"`
	Arg0     string `json:"arg0"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

const MAX_IDLE_IN_SECONDS = 600 * time.Second

func (world *World) NewSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Error:")
		return
	}
	defer conn.Close()

	token, success := getTokenFromFirstMessage(conn)
	if !success {
		logger.Info().Msg("Invalid Connection")
		return
	}

	incoming := world.retreiveIncoming(token)
	if incoming == nil {
		logger.Info().Msg("player not found with token: " + token)
		return
	}

	player := world.join(incoming, conn)
	if player == nil {
		logger.Info().Msg("Failed to join player with token: " + token)
		return
	}

	incrementSessionLogins(world)
	handleNewPlayer(player)
}

func handleNewPlayer(player *Player) {
	defer initiateLogout(player)
	logger.Info().Msg("New Connection from: " + player.username)
	lastRead := time.Unix(0, 0)
	for {
		player.conn.SetReadDeadline(time.Now().Add(MAX_IDLE_IN_SECONDS))
		_, msg, err := player.conn.ReadMessage()
		if err != nil {
			// break will initiate logout:
			sendUpdate(player, divLogOutResume("Inactive. Logging out", player.world.config.domainName))
			break
		}
		currentRead := time.Now()

		event, success := getKeyPress(msg) // If your press was read it must now complete before logout e.g. player is tangible
		if !success {
			logger.Info().Msg("Invalid input")
			continue
		}

		if player.handlePressActive(event) {
			lastRead = currentRead
			time.Sleep(20 * time.Millisecond)
			continue
		}

		elapsedTime := time.Since(lastRead)
		if elapsedTime <= 40*time.Millisecond {
			continue
		}
		lastRead = currentRead

		player.handlePress(event)
	}
}

func getTokenFromFirstMessage(conn *websocket.Conn) (token string, success bool) {
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		logger.Error().Err(err).Msg("Error reading message from Connection: ")
		return "", false
	}

	var msg struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing JSON:")
		return "", false
	}

	return msg.Token, true
}

func getKeyPress(input []byte) (event *PlayerSocketEvent, success bool) {
	event = &PlayerSocketEvent{}
	err := json.Unmarshal(input, event)
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing JSON:")
		return nil, false
	}
	return event, true
}

func (player *Player) handlePressActive(event *PlayerSocketEvent) bool {
	if event.Name == "Space-On" {
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
		return true
	}
	return false
}

func (player *Player) handlePress(event *PlayerSocketEvent) {
	switch event.Name {
	case "w":
		moveNorth(player)
	case "a":
		moveWest(player)
	case "s":
		moveSouth(player)
	case "d":
		moveEast(player)
	case "W":
		player.moveNorthBoost()
	case "A":
		player.moveWestBoost()
	case "S":
		player.moveSouthBoost()
	case "D":
		player.moveEastBoost()
	case "f":
		updateEntireExistingScreen(player)
	case "g":
		makeHallucinate(player)
	case "h":
		player.cycleHats()
	case "e":
		rotate(player, true)
		//spawnNewNPCWithRandomMovement(player, 300)
	case "Shift-On":
		updateOne(divInputShift(), player)
	case "Shift-Off":
		updateOne(divInput(), player)
	case "menuOn":
		openPauseMenu(player)
	case "menuOff":
		turnMenuOff(player)
	case "menuDown":
		menuDown(player, *event)
	case "menuUp":
		menuUp(player, *event)
	case "menuClick":
		if menu, ok := player.menues[event.MenuName]; ok {
			menu.attemptClick(player, *event)
		}
	default:
		// Unrecognized input
	}
}

/////////////////////////////////////////////////////////////////////////
// NPCs

type MockConn struct{}

func (m *MockConn) WriteMessage(messageType int, data []byte) error {
	return nil
}

func (m *MockConn) ReadMessage() (messageType int, p []byte, err error) {
	return 1, []byte("mock data"), nil
}

func (m *MockConn) Close() error {
	// Adjust so that subsequent reads have error?
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func spawnNewPlayerWithRandomMovement(ref *Player, interval int) (*Player, context.CancelFunc) {
	username := "user-" + uuid.New().String()
	refTile := ref.getTileSync()
	record := PlayerRecord{Username: username, Health: 50, StageName: ref.stage.name, X: refTile.x, Y: refTile.y, Team: "test-team-2", Trim: "red-b thick"}
	loginRequest := createLoginRequest(record)
	ref.world.addIncoming(loginRequest)
	newPlayer := ref.world.join(loginRequest, &MockConn{})
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				<-newPlayer.updates
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(interval) * time.Millisecond)
				randn := rand.Intn(5000)

				if randn%4 == 0 {
					moveNorth(newPlayer)
				}
				if randn%4 == 1 {
					moveSouth(newPlayer)
				}
				if randn%4 == 2 {
					moveEast(newPlayer)
				}
				if randn%4 == 3 {
					moveWest(newPlayer)
				}
			}
		}
	}(ctx)
	return newPlayer, cancel
}
