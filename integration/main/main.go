package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
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
	host := os.Getenv("BLOOP_HOST")
	tokenEndpoint := host + "/insert"
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	payload := []byte("secret=" + secret + "&username=uname&stagename=team-blue:3-3&team=fuchsia&count=120")

	// Make POST to retrieve tokens
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch tokens: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}
	var tokens []string
	if err := json.Unmarshal(body, &tokens); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse token response: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Retrieved tokens:", tokens)

	//sockets := make([]*TestingSocket, 0, len(tokens))
	for _, token := range tokens {
		fmt.Println(token)
		testingSocket := createTestingSocket(host + "/screen")
		if testingSocket == nil {
			fmt.Println("failed to create testing socket")
			return
		}
		defer testingSocket.ws.Close()
		//sockets = append(sockets, testingSocket)
		testingSocket.writeOrFatal(createInitialTokenMessage(token))
		go func() {
			for testingSocket.readOrFatal() != nil {

			}
		}()

		go func(token string) {
			for {
				randn := rand.Intn(5000)
				if randn%4 == 0 {
					if testingSocket.writeOrFatal(createSocketEventMessage(token, "a")) != nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 1 {
					if testingSocket.writeOrFatal(createSocketEventMessage(token, "w")) != nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 2 {
					if testingSocket.writeOrFatal(createSocketEventMessage(token, "d")) != nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				if randn%4 == 3 {
					if testingSocket.writeOrFatal(createSocketEventMessage(token, "s")) != nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(token)
	}

	time.Sleep(120000 * time.Millisecond)
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
		return nil
		//panic(fmt.Sprintf("could not dial: %v", err))
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

func (ts *TestingSocket) writeOrFatal(msg []byte) error {
	err := ts.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		fmt.Printf("could not send message: %s, Error: %v\n", string(msg), err)
		return err
		//panic(fmt.Sprintf("could not send message: %s, Error: %v", string(msg), err))
	}
	return nil
}

func (ts *TestingSocket) readOrFatal() []byte {
	_, msg, err := ts.ws.ReadMessage()
	if err != nil {
		fmt.Printf("could not read message - Error: %v\n", err)
		return nil
		//panic(fmt.Sprintf("could not read message - Error: %v", err))
	}
	return msg
}
