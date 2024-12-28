package main

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (world *World) NewSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	token, success := getTokenFromFirstMessage(conn) // Pattern for input?
	if !success {
		fmt.Println("Invalid Connection")
		return
	}

	incoming := world.retreiveIncoming(token)
	if incoming == nil {
		fmt.Println("player not found with token: " + token)
		return
	}
	existingPlayer := world.join(*incoming)
	if existingPlayer == nil {
		fmt.Println("Failed to join player with token: " + token)
		return
	}
	existingPlayer.conn = conn

	handleNewPlayer(existingPlayer)
}

func handleNewPlayer(existingPlayer *Player) {
	defer logOut(existingPlayer)
	go existingPlayer.sendUpdates()
	stage := getStageFromStageName(existingPlayer.world, existingPlayer.stageName)
	placePlayerOnStageAt(existingPlayer, stage, existingPlayer.y, existingPlayer.x)
	fmt.Println("New Connection")
	for {
		_, msg, err := existingPlayer.conn.ReadMessage()
		if err != nil {
			// This allows for rage quit by pressing X, should add timeout to encourage finding safety
			fmt.Println("Ending", err)
			break
		}

		event, success := getKeyPress(msg)
		if !success {
			fmt.Println("Invalid input")
			continue
		}
		if event.Token != existingPlayer.id {
			// check mildly irrelevant?
			fmt.Println("Cheating")
			break
		}
		// Throttle input here?

		existingPlayer.handlePress(event)
		if existingPlayer.conn == nil {
			return
		}
	}
}

func logOut(player *Player) {
	time.Sleep(500 * time.Millisecond)
	player.removeFromTileAndStage() // check if successful?

	player.updateRecord() // Should return error
	player.world.wPlayerMutex.Lock()
	delete(player.world.worldPlayers, player.id)
	index, exists := player.world.leaderBoard.mostDangerous.index[player]
	if exists {
		// this is unsafe index position can change
		heap.Remove(&player.world.leaderBoard.mostDangerous, index)
		//  If index was 0 before, need to update new most dangerous
		if index == 0 {
			fmt.Println("New Most Dangerous!")
			mostDangerous := player.world.leaderBoard.mostDangerous.Peek()
			if mostDangerous != nil {
				notifyChangeInMostDangerous(mostDangerous)
			}
		}
	}
	player.world.wPlayerMutex.Unlock()

	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn != nil {
		player.conn.Close()
	}
	player.conn = nil

	// hmm.
	close(player.updates)
	//safeClose(player.updates)

	fmt.Println("Logging Out " + player.username)
}

/*
func safeClose(ch chan []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	close(ch)
}
*/

func getTokenFromFirstMessage(conn *websocket.Conn) (token string, success bool) {
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	var msg struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", false
	}

	return msg.Token, true
}

type PlayerSocketEvent struct {
	Token    string `json:"token"`
	Name     string `json:"eventname"`
	MenuName string `json:"menuName"`
	Arg0     string `json:"arg0"`
}

func getKeyPress(input []byte) (event *PlayerSocketEvent, success bool) {
	event = &PlayerSocketEvent{}
	err := json.Unmarshal(input, event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, false
	}
	return event, true
}

func (player *Player) handlePress(event *PlayerSocketEvent) {
	if event.Name == "w" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "twilight"</script> </div>`
		updateOne(class, player)*/
		player.moveNorth()
	}
	if event.Name == "a" {
		player.moveWest()
	}
	if event.Name == "s" {
		player.moveSouth()
	}
	if event.Name == "d" {
		player.moveEast()
	}
	if event.Name == "W" {
		player.moveNorthBoost()
	}
	if event.Name == "A" {
		player.moveWestBoost()
	}
	if event.Name == "S" {
		player.moveSouthBoost()
	}
	if event.Name == "D" {
		player.moveEastBoost()
	}
	if event.Name == "f" {
		updateScreenFromScratch(player)
	}
	if event.Name == "g" {

		//Full swap takes priority in either order, otherwise both may apply

		//exTile := `<div id="t1-0" class="box top green"></div>
		//			<div id="t1-1" class="box top green"></div>`

		go func() {
			for i := 0; i <= 80; i++ {
				time.Sleep(20 * time.Millisecond)
				updateOne(generateDivs(i), player)
			}

		}()

		//updateOne(generateWeatherSolid("blue trsp20"), player)
		//player.updates <- generateWeatherSolidBytes("night trsp20")

		/*
			trsp := []string{"trsp20", "trsp40", "trsp60", "trsp80"}
			for _, t := range trsp {
				d := time.Duration(75)
				player.updates <- generateWeatherDynamic(twoColorParity("red", "pink", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("pink", "red", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("green", "grass", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("grass", "green", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("blue", "ice", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("ice", "blue", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("black", "half-gray", t))
				time.Sleep(d * time.Millisecond)
				player.updates <- generateWeatherDynamic(twoColorParity("half-gray", "black", t))
				time.Sleep(d * time.Millisecond)

			}
		*/

		/*
			//player.updateBottomText("Heyo ;) ")
			spawnNewPlayerWithRandomMovement(player, 250)
			npcs++
			println("npcs: ", npcs)
		*/
	}
	if event.Name == "Space-On" {
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
	}
	if event.Name == "Shift-On" {
		updateOne(divInputShift(), player)
	}
	if event.Name == "Shift-Off" {
		updateOne(divInput(), player)
	}
	if event.Name == "menuOn" {
		openPauseMenu(player)
	}
	if event.Name == "menuOff" {
		turnMenuOff(player)
	}
	if event.Name == "menuDown" {
		menuDown(player, *event)
	}
	if event.Name == "menuUp" {
		menuUp(player, *event)
	}
	if event.Name == "menuClick" {
		menu, ok := player.menues[event.MenuName]
		if ok {
			menu.attemptClick(player, *event)
		}
	}
}

func generateDivs(frame int) string {
	var sb strings.Builder

	// Define the center of the grid
	center := 7.5

	// Determine which color set to use based on the frame
	var col1, col2 string
	if ((frame / 20) % 2) == 0 {
		// Use the first color set
		col1, col2 = "red trsp40", "gold trsp40"
	} else {
		// Use the second color set
		col1, col2 = "blue trsp40", "green trsp40"
	}

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			dx := float64(i) - center
			dy := float64(j) - center

			// Radius from the center
			r := math.Sqrt(dx*dx + dy*dy)

			// Base angle in radians, range (-π, π]
			angle := math.Atan2(dy, dx)

			// Add rotation based on the frame
			angle += float64(frame) * 0.1

			// Determine pattern: If this value is even, use col1; if odd, use col2
			// Multiplying angle by r gives a spiral-like indexing pattern.
			colorIndex := int((angle * r))

			var color string
			if colorIndex%2 == 0 {
				color = col1
			} else {
				color = col2
			}

			sb.WriteString(fmt.Sprintf(`<div id="w%d-%d" class="box zw %s"></div>`+"\n", i, j, color))
		}
	}

	return sb.String()
}

func generateWeatherSolid(color string) string {
	var sb strings.Builder

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			sb.WriteString(fmt.Sprintf(`<div id="w%d-%d" class="box zw %s"></div>`+"\n", i, j, color))
		}
	}

	return sb.String()
}

func generateWeatherDumb(color string) string {
	out := ""

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			out += fmt.Sprintf(`<div id="w%d-%d" class="box zw %s"></div>`+"\n", i, j, color)
		}
	}

	return out
}

func generateWeatherSolidBytes(color string) []byte {
	const rows = 16
	const cols = 16

	// Estimate capacity to avoid growth:
	// Each element has a pattern:
	// <div id="w{i}-{j}" class="box zw {color}"></div>\n
	//
	// Breakdown of constant parts:
	// "<div id=\"w"      = 10 bytes (including the quote)
	// "-"                = 1 byte
	// "\" class=\"box zw " = 15 bytes (including the leading quote)
	// "\"></div>"       = 8 bytes (including quotes and newline)
	// Total constant overhead per line = 10 + 1 + 15 + 9 = 34 bytes
	//
	// Now add the length for i and j (up to "15") and "w":
	// "w" + i + "-" + j: "w" (1 byte), max i=2 digits, "-" (1 byte), max j=2 digits
	// Max i and j length = 2 digits each = 4 bytes + "w" + "-" = 6 bytes max
	// So worst: 34 (constant) + 6 (id part) = 40 bytes + len(color) per line
	//
	// We have 256 lines (16x16):
	// capacity ~ 256 * (40 + len(color))
	estCap := 256 * (40 + len(color))
	b := make([]byte, 0, estCap)

	prefix := []byte(`<div id="w`)
	sep := []byte(`-`)
	cls := []byte(`" class="box zw `)
	suffix := []byte(`"></div>`)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			b = append(b, prefix...)
			b = strconv.AppendInt(b, int64(i), 10)
			b = append(b, sep...)
			b = strconv.AppendInt(b, int64(j), 10)
			b = append(b, cls...)
			b = append(b, color...)
			b = append(b, suffix...)
		}
	}

	return b
}

func generateWeatherDynamic(getColor func(i, j int) string) []byte {
	estCap := 256 * 60
	b := make([]byte, 0, estCap)

	prefix := []byte(`<div id="w`)
	sep := []byte(`-`)
	cls := []byte(`" class="box zw `)
	suffix := []byte(`"></div>`)

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			b = append(b, prefix...)
			b = strconv.AppendInt(b, int64(i), 10)
			b = append(b, sep...)
			b = strconv.AppendInt(b, int64(j), 10)
			b = append(b, cls...)
			b = append(b, getColor(i, j)...)
			b = append(b, suffix...)
		}
	}
	return b
}

func twoColorParity(c1, c2, t string) func(i, j int) string {
	return func(i, j int) string {
		if (i+j)%2 == 0 {
			return c1 + "-b thick " + c2 + " " + t
		} else {
			return c2 + "-b thick " + c1 + " " + t

		}
	}
}

func generateDivs3(frame int) string {
	var sb strings.Builder

	// Define a set of colors to cycle through
	colors := []string{"red", "blue", "green", "gold", "white", "black", "half-gray"}

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			// Compute color based on i, j, and the current frame.
			// This will cause the color pattern to "shift" each frame.
			color := colors[(i+j+frame)%len(colors)]

			sb.WriteString(fmt.Sprintf(`<div id="t%d-%d" class="box top %s"></div>`+"\n", i, j, color))
		}
	}

	return sb.String()
}

var npcs = 0

func spawnNewPlayerWithRandomMovement(ref *Player, interval int) (*Player, context.CancelFunc) {
	username := "user-" + uuid.New().String()
	record := PlayerRecord{Username: username, Health: 50, Y: ref.y, X: ref.x, StageName: ref.stage.name, Team: "fuchsia", Trim: "white-b thick"}
	loginRequest := createLoginRequest(record)
	ref.world.addIncoming(loginRequest)
	newPlayer := ref.world.join(loginRequest)
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
				<-newPlayer.clearUpdateBuffer
			}
		}
	}(ctx)
	s := getStageFromStageName(newPlayer.world, newPlayer.stageName)
	placePlayerOnStageAt(newPlayer, s, newPlayer.y, newPlayer.x)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(interval) * time.Millisecond)
				randn := rand.Intn(5000)

				if randn%4 == 0 {
					newPlayer.moveNorth()
				}
				if randn%4 == 1 {
					newPlayer.moveSouth()
				}
				if randn%4 == 2 {
					newPlayer.moveEast()
				}
				if randn%4 == 3 {
					newPlayer.moveWest()
				}
			}
		}
	}(ctx)
	return newPlayer, cancel
}
