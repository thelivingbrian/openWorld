package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Initializing...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/mass", IntegrationA)

	err = http.ListenAndServe(":4440", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

func IntegrationA(w http.ResponseWriter, r *http.Request) {
	tokens := []string{
		"ee45a63020fd11c23c5a598bfc199a3b",
		"de1165dd9700d364c049362a9606d090",
		"5dcd3fc6b60477a15c144b18b4074df2",
		"7160f313d63f596f8decef131aba4263",
		"52ee9ccb5954ce08c25256089837b140",
		"624ee350df0f51e075cf8518ea93470b",
		"e45f1c9496f62fb40d23360d50c07d16",
		"16b7b579ce88aa5d39160d7553a41a75",
		"52526db832db170aed6fb8387fe68d3a",
		"97c4347dfff766541540542db90f70c3",
		"d447ae9cdc99311008b00897bcd0eea2",
		"43db6f941753ab4ec29f0ed76625fc7b",
		"4f83d962f415f28b686297b2459d69c6",
		"0c336def98e1fd02c72bfa10783f4ec5",
		"b91dd5cbd8a293db80566e08f3ee3940",
		"6d3f4b02b931ce74b1315f5e7f72219a",
		"5b01bd12217a78b8494a0a473d99fb8c",
		"08d1f28af7ca3d797d0071ab13f6ecd9",
		"7b8f9250d3231e9e184c43748e747962",
		"8e42defd84ac8f67d208bcb0668efb38",
		"1f810c2fe166fa70826d0c801bdbea63",
		"fb8753f5c8d81013736a51ea5c44f0a2",
		"7536e11f6594ad3eddfe3d6c91ecc8af",
		"c16ed86b709fa62e5beed8ddcb2d3151",
		"4185b8bcab7b865ec9aba1685c890a77",
		"c142f21e9f429ab8c0fcf99a340447ab",
		"be8d9b0f75614a9242df1e49916ad4c9",
		"c72fd3875a1573a2827115d6c24a6873",
		"9638711578618db3b2da3a365375ac51",
		"24de2d4d20dd3fe0f47ddce1625829ea",
	}
	//sockets := make([]*TestingSocket, 0, len(tokens))
	for _, token := range tokens {
		fmt.Println(token)
		testingSocket := createTestingSocket("http://localhost:9090/screen")
		defer testingSocket.ws.Close()
		//sockets = append(sockets, testingSocket)
		testingSocket.writeOrFatal(createInitialTokenMessage(token))
		// go func() {
		// 	for testingSocket.readOrFatal() != nil {

		// 	}
		// }()

		go func(token string) {
			for {

				randn := 0 //rand.Intn(5000)
				if randn%4 == 0 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "a"))
					//fmt.Println(testingSocket.readOrFatal())
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 0 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "w"))
					//fmt.Println(testingSocket.readOrFatal())
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 0 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "d"))
					//fmt.Println(testingSocket.readOrFatal())
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 0 {
					testingSocket.writeOrFatal(createSocketEventMessage(token, "s"))
					//fmt.Println(testingSocket.readOrFatal())
					time.Sleep(100 * time.Millisecond)
				}

				// testingSocket.writeOrFatal(createSocketEventMessage(token, "w"))
				// _ = testingSocket.readOrFatal()
				// time.Sleep(100 * time.Millisecond)
			}
		}(token)
	}

	time.Sleep(90000 * time.Millisecond)
}

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
		Token: token, //"",
	}
	initialTokenMessage, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("could not marshal: %v", err))
	}
	return initialTokenMessage
}

// import ?
type PlayerSocketEvent struct {
	Token    string `json:"token"`
	Name     string `json:"eventname"`
	MenuName string `json:"menuName"`
	Arg0     string `json:"arg0"`
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
		fmt.Println(fmt.Sprintf("could not read message - Error: %v", err))
		return nil
		//panic(fmt.Sprintf("could not read message - Error: %v", err))
	}
	return msg
}
