package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// make endpoint
func TestIntegrationA(t *testing.T) {
	start := time.Now()
	fmt.Println("0")
	fmt.Println("Elapsed time:", time.Since(start))
	//world := createGameWorld(testdb())

	//req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	//world.addIncoming(req)
	//p := world.join(req)
	//initialCoordiate := 2

	testingSocket := createTestingSocket("https://bloop.world/screen")
	defer testingSocket.ws.Close()

	tokens := []string{
		"a0b04e4b4fcf61e0c25a38d4169644b7",
		"fb88adf9b531a6acf9afb8a0f659515d",
		"370e35a67e80e73836f56c065d6cf0d5",
		"6208f5cdbb9ed7c2cf4dfeee280b61fc",
		"1a85fe5d6c4456beed30ded32fdc1f8b",
		"5eb1c74f433c6ad9f42098d6d090b29f",
		"458406fc7a4caa051ee3277db0a76289",
		"2077d7dbb3afd2537b7597d80a792e93",
		"ae23f91998f038782b024fa3d7644fd2",
		"0ad47ed927d8c0a4b27c34e7b62bc52a",
		"536d3166a727f83d8cdf772a306b613b",
		"b018499aa7da8b738b2db1c86aa4dda4",
		"99aa10125ea751b8eca2df947d62a843",
		"c93707d557fb1932a0bfff534f9b1267",
		"51a456360236c4fcea929926050f0db7",
		"714bc969e4ed7d4277291a0e41306764",
		"db8c4cf969b7bf6499c37f5d83d523ab",
		"f968526b306042a57010927dfee869bd",
		"2849be7b6890ab9772f4711e30533c05",
		"fb4c2f57d19ba3af282ac25bea735623",
	}

	sockets := make([]*TestingSocket, 0, len(tokens))
	for _, token := range tokens {
		fmt.Println(token)
		testingSocket := createTestingSocket("https://bloop.world/screen")
		defer testingSocket.ws.Close()
		sockets = append(sockets, testingSocket)
		testingSocket.writeOrFatal(createInitialTokenMessage(token))
		_ = testingSocket.readOrFatal()
		go func() {
			for {

				randn := rand.Intn(5000)
				if randn%4 == 0 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "w"))
					_ = testingSocket.readOrFatal()
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 1 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "s"))
					_ = testingSocket.readOrFatal()
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 2 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "d"))
					_ = testingSocket.readOrFatal()
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 3 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "a"))
					_ = testingSocket.readOrFatal()
					time.Sleep(100 * time.Millisecond)
				}

				// testingSocket.writeOrFatal(createSocketEventMessage(token, "w"))
				// _ = testingSocket.readOrFatal()
				// time.Sleep(100 * time.Millisecond)
			}
		}()
	}
	time.Sleep(30000 * time.Millisecond)
}

func TestSocketJoinAndMove(t *testing.T) {
	world := createGameWorld(testdb())

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: "test-large"})
	world.addIncoming(req)
	//p := world.join(req)
	//initialCoordiate := 2

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		world.NewSocketConnection(w, r)
	}))
	defer server.Close()

	//testingSocket := createTestingSocket("https://bloop.world/screen")
	testingSocket := createTestingSocket(server.URL)
	defer testingSocket.ws.Close()

	//id := "fc7ee25f-376b-422e-866f-7bde0fd742f8"

	testingSocket.writeOrFatal(createInitialTokenMessage(req.Token))
	_ = testingSocket.readOrFatal()

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))
	_ = testingSocket.readOrFatal()

	testingSocket.writeOrFatal(createSocketEventMessage(req.Token, "d"))
	_ = testingSocket.readOrFatal()

	// Assert
	if len(world.worldPlayers) != 1 {
		t.Error("Incorrect number of players")
	}

	if world.leaderBoard.mostDangerous.Peek().username == req.Record.Username {
		t.Error("New Player should not be most dangerous") // Minimum killstreak player prevents newly joining from being most dangerous
	}

	//fmt.Println(p.x)
	/*
		Cannot be tested via the socket
			if initialCoordiate == p.x {
				t.Error("Player has not moved") // This can fail due to race
			}
	*/
}

func TestMostDangerous(t *testing.T) {
	world := createGameWorld(testdb())
	stage := getStageFromStageName(world, "test-large")

	req := createLoginRequest(PlayerRecord{Username: "test1", Y: 2, X: 2, StageName: stage.name})
	world.addIncoming(req)
	p1 := world.join(req)
	go p1.sendUpdates()
	p1.placeOnStage(stage)

	req2 := createLoginRequest(PlayerRecord{Username: "test2", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req2)
	p2 := world.join(req2)
	go p2.sendUpdates()
	p2.placeOnStage(stage)

	req3 := createLoginRequest(PlayerRecord{Username: "test3", Y: 3, X: 3, StageName: stage.name})
	world.addIncoming(req3)
	p3 := world.join(req3)
	go p3.sendUpdates()
	p3.placeOnStage(stage)

	// Assert
	p2.incrementKillStreak()
	if world.leaderBoard.mostDangerous.Peek() != p2 {
		t.Error("Invalid leader should be p2")
	}

	p3.incrementKillStreak()
	p3.incrementKillStreak()
	p3.incrementKillStreak()
	p1.incrementKillStreak()
	p1.incrementKillStreak()
	if world.leaderBoard.mostDangerous.Peek() != p3 {
		t.Error("Invalid leader should be p3")
	}

	logOut(p3)
	if world.leaderBoard.mostDangerous.Peek() != p1 {
		t.Error("Invalid leader should be p1")
	}

}

//////////////////////////////////////////////////////////////////////////////////////
// Helpers

type TestingSocket struct {
	ws *websocket.Conn
}

func createTestingSocket(url string) *TestingSocket {
	if url[0:4] == "http" {
		url = "ws" + url[len("http"):]

	}
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(fmt.Sprintf("could not dial: %v", err))
	}
	return &TestingSocket{ws: ws}
}

func createInitialTokenMessage(token string) []byte {
	var msg = struct {
		Token string
	}{
		Token: token, //"TestToke",
	}
	initialTokenMessage, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("could not marshal: %v", err))
	}
	return initialTokenMessage
}

func createSocketEventMessage(token, name string) []byte {
	var msg = PlayerSocketEvent{
		Token: token,
		Name:  name,
	}
	socketMsg, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("could not marshal: %v", err))
	}
	return socketMsg
}

func (ts *TestingSocket) writeOrFatal(msg []byte) {
	err := ts.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		panic(fmt.Sprintf("could not send message: %s, Error: %v", string(msg), err))
	}
}

func (ts *TestingSocket) readOrFatal() []byte {
	_, msg, err := ts.ws.ReadMessage()
	if err != nil {
		panic(fmt.Sprintf("could not read message - Error: %v", err))
	}
	return msg
}
