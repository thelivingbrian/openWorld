package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	http.HandleFunc("/mass", IntegrationClientBed)

	err = http.ListenAndServe(":4440", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

func IntegrationClientBed(w http.ResponseWriter, r *http.Request) {
	// curl -X POST "http://localhost:4440/mass?stagename=team-fuchsia:3-3&read=true&count=5200&ttl=55&action=lr&team=fuchsia"
	// curl -X GET "http://localhost:4440/stats"
	stagename := r.URL.Query().Get("stagename")

	var read bool
	readParam := r.URL.Query().Get("read")
	if readParam != "" {
		var err error
		read, err = strconv.ParseBool(readParam)
		if err != nil {
			http.Error(w, "Invalid 'read' parameter", http.StatusBadRequest)
			return
		}
	}

	count := r.URL.Query().Get("count")
	if stagename == "" || count == "" {
		http.Error(w, "Missing required param stagename", http.StatusBadRequest)
		return
	}

	ttlString := r.URL.Query().Get("ttl")
	ttl, err := strconv.Atoi(ttlString)
	if err != nil {
		fmt.Println("Invalid TTL")
		ttl = 30
	}

	team := r.URL.Query().Get("team")

	action := r.URL.Query().Get("action")
	socketAction := moveInCircles
	if action == "random" {
		socketAction = moveRandomly
	}
	if action == "lr" {
		socketAction = leftRight
	}

	tokens := requestTokens(stagename, count, team)
	go func() {
		createSocketsAndSendActions(tokens, read, ttl, socketAction)
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Request successful!")
}

func createSocketsAndSendActions(tokens []string, read bool, ttl int, action func(*TestingSocket, string)) {
	for _, token := range tokens {
		testingSocket := createTestingSocket(os.Getenv("BLOOP_HOST") + "/screen")
		if testingSocket == nil {
			fmt.Println("failed to create testing socket")
			return
		}

		term := make(chan bool)
		go testingSocket.closeOnRec(term) //defer testingSocket.ws.Close() occurs in closeOnRec
		go func(chan bool) {
			time.Sleep(time.Duration(ttl) * time.Second)
			term <- true
		}(term)

		testingSocket.tryWrite(createInitialTokenMessage(token))

		if read {
			go testingSocket.readUntilNil()
		}

		go action(testingSocket, token)
	}

}

func requestTokens(stagename, count, team string) []string {
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	username := createRandomString()
	payload := []byte("secret=" + secret + "&username=" + username + "&stagename=" + stagename + "&team=" + team + "&count=" + count)

	tokenEndpoint := os.Getenv("BLOOP_HOST") + "/insert"
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Failed to fetch tokens: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v", err)
		return nil
	}
	var tokens []string
	if err := json.Unmarshal(body, &tokens); err != nil {
		fmt.Printf("Failed to parse token response: %v", err)
		return nil
	}

	fmt.Println("Retrieved tokens:", tokens)
	return tokens
}

func createRandomString() string {
	bytes := make([]byte, 8)
	_, err := crand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
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
		fmt.Printf("could not dial: %v\n", err)
		return nil
	}
	return &TestingSocket{ws: ws}
}

func createInitialTokenMessage(token string) []byte {
	var msg = struct {
		Token string
	}{
		Token: token,
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

func (ts *TestingSocket) tryWrite(msg []byte) error {
	err := ts.ws.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		fmt.Printf("could not send message: %s, Error: %v\n", string(msg), err)
		return err
	}
	return nil
}

func (ts *TestingSocket) closeOnRec(term chan bool) {
	<-term
	ts.ws.Close()
}

func (ts *TestingSocket) tryRead() []byte {
	_, msg, err := ts.ws.ReadMessage()
	if err != nil {
		fmt.Printf("could not read message - Error: %v\n", err)
		return nil
	}
	return msg
}

func (ts *TestingSocket) readUntilNil() {
	for ts.tryRead() != nil {

	}
}

func moveRandomly(ts *TestingSocket, token string) {
	for {
		randn := rand.Intn(5000)
		if randn%4 == 0 {
			if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if randn%4 == 1 {
			if ts.tryWrite(createSocketEventMessage(token, "w")) != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if randn%4 == 2 {
			if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if randn%4 == 3 {
			if ts.tryWrite(createSocketEventMessage(token, "s")) != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if randn%250 == 0 {
			if ts.tryWrite(createSocketEventMessage(token, "Space-On")) != nil {
				break
			}
		}
	}
}

func moveInCircles(ts *TestingSocket, token string) {
	for {
		if ts.tryWrite(createSocketEventMessage(token, "w")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "s")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func leftRight(ts *TestingSocket, token string) {
	for {
		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "d")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)

		if ts.tryWrite(createSocketEventMessage(token, "a")) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
