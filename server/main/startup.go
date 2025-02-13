package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

///////////////////////////////////////////////////
//Database

type DB struct {
	users         *mongo.Collection
	playerRecords *mongo.Collection
	events        *mongo.Collection
}

func createDbConnection(config *Configuration) *DB {
	mongodb := mongoClient(config).Database("bloopdb")
	return &DB{mongodb.Collection("users"), mongodb.Collection("players"), mongodb.Collection("events")}
}

func mongoClient(config *Configuration) *mongo.Client {
	clientOptions := options.Client().ApplyURI(config.getMongoURI())

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

////////////////////////////////////////////////////
// Configuration

type Configuration struct {
	envName            string
	port               string
	usesTLS            bool
	tlsCertPath        string
	tlsKeyPath         string
	mongoHost          string
	mongoPort          string
	mongoPrefix        string
	mongoUser          string
	mongoPass          string
	hashKey            []byte
	blockKey           []byte
	googleClientId     string
	googleClientSecret string
	googleCallbackUrl  string
	isHub              bool
	serverName         string
	domainName         string
}

func getConfiguration() *Configuration {
	environmentName := os.Getenv("BLOOP_ENV")
	hashKey, blockKey := retrieveKeys()

	config := Configuration{
		envName:            environmentName,
		port:               os.Getenv("BLOOP_PORT"),
		usesTLS:            true,
		tlsCertPath:        os.Getenv("BLOOP_TLS_CERT_PATH"),
		tlsKeyPath:         os.Getenv("BLOOP_TLS_KEY_PATH"),
		mongoHost:          os.Getenv("MONGO_HOST"),
		mongoPort:          os.Getenv("MONGO_PORT"),
		mongoPrefix:        os.Getenv("MONGO_PREFIX"),
		mongoUser:          os.Getenv("MONGO_USER"),
		mongoPass:          os.Getenv("MONGO_PASS"),
		hashKey:            hashKey,
		blockKey:           blockKey,
		googleClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		googleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		googleCallbackUrl:  os.Getenv("GOOGLE_CALLBACK_URL"),
		isHub:              strings.ToUpper(os.Getenv("IS_HUB")) == "TRUE",
		serverName:         os.Getenv("SERVER_NAME"),
		domainName:         os.Getenv("DOMAIN_NAME"),
	}

	if environmentName == "prod" {
		log.Fatal("No Prod Environment")
	} else if environmentName == "test" {
		// Nothing to do
	} else if environmentName == "dev" {
		config.usesTLS = false
	} else {
		log.Fatal("No Configuration, exiting")
	}

	return &config
}

func (config *Configuration) getMongoCredentialString() string {
	if config.mongoUser != "" && config.mongoPass != "" {
		return config.mongoUser + ":" + config.mongoPass + "@"
	}
	return ""
}

func (config *Configuration) getMongoURI() string {
	// add config .mongoPrefix
	return "mongodb" + config.mongoPrefix + "://" + config.getMongoCredentialString() + config.mongoHost + config.mongoPort
}

func (config *Configuration) createCookieStore() *sessions.CookieStore {
	if len(config.hashKey) != 32 || len(config.blockKey) != 32 {
		panic("Invalid key lengths")
	}
	return sessions.NewCookieStore(config.hashKey, config.blockKey)
}

func (config *Configuration) isServer() bool {
	return config.serverName != "" && config.domainName != ""
}

func retrieveKeys() (hashKey, blockKey []byte) {
	hashKeyBase64 := os.Getenv("COOKIE_HASH_KEY")
	hashKey, err := base64.StdEncoding.DecodeString(hashKeyBase64)
	if err != nil {
		log.Fatalf("Error decoding Base64 key: %v", err)
	}

	blockKeyBase64 := os.Getenv("COOKIE_BLOCK_KEY")
	blockKey, err = base64.StdEncoding.DecodeString(blockKeyBase64)
	if err != nil {
		log.Fatalf("Error decoding Base64 key: %v", err)
	}
	if len(hashKey) != 32 || len(blockKey) != 32 {
		panic(fmt.Sprintf("Invalid key length for hashkey[%d] or blockKey[%d] expecting 32 bytes.", len(hashKey), len(blockKey)))
	}
	return hashKey, blockKey
}

////////////////////////////////////////////////////
// Load Resources from JSON

type Material struct {
	ID          int    `json:"id"`
	CommonName  string `json:"commonName"`
	CssColor    string `json:"cssColor"`
	Walkable    bool   `json:"walkable"`
	Floor1Css   string `json:"layer1css"`
	Floor2Css   string `json:"layer2css"`
	Ceiling1Css string `json:"ceiling1css"`
	Ceiling2Css string `json:"ceiling2css"`
	DisplayText string `json:"displayText"`
}

// add color
type Transport struct {
	SourceY            int    `json:"sourceY"`
	SourceX            int    `json:"sourceX"`
	DestY              int    `json:"destY"`
	DestX              int    `json:"destX"`
	DestStage          string `json:"destStage"`
	Confirmation       bool   `json:"confirmation"`
	RejectInteractable bool   `json:"rejectInteractable"`
}

type Area struct {
	Name             string                       `json:"name"`
	Safe             bool                         `json:"safe"`
	Tiles            [][]int                      `json:"tiles"`
	Transports       []Transport                  `json:"transports"`
	Interactables    [][]*InteractableDescription `json:"interactables"`
	DefaultTileColor string                       `json:"defaultTileColor"`
	North            string                       `json:"north"`
	South            string                       `json:"south"`
	East             string                       `json:"east"`
	West             string                       `json:"west"`
	MapId            string                       `json:"mapId"`
	LoadStrategy     string                       `json:"loadStrategy,omitempty"`
	SpawnStrategy    string                       `json:"spawnStrategy,omitempty"`
}

type InteractableDescription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SetName   string `json:"setName"`
	CssClass  string `json:"cssClass"`
	Pushable  bool   `json:"pushable"`
	Walkable  bool   `json:"walkable"`
	Fragile   bool   `json:"fragile"`
	Reactions string `json:"reactions"`
}

var (
	materials []Material
	areas     []Area
)

func populateStructUsingFileName[T any](ptr *T, filename string) {
	jsonData, err := os.ReadFile(fmt.Sprintf("./data/%s.json", filename))
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, ptr); err != nil {
		panic(err)
	}
}

// This should return values instead of populating globals
func loadFromJson() {
	populateStructUsingFileName(&materials, "materials")
	populateStructUsingFileName(&areas, "areas")

	//fmt.Printf("Loaded %d materials.", len(materials))
	//fmt.Printf("Loaded %d areas.", len(areas))
}

func areaFromName(s string) (area Area, success bool) {
	for _, area := range areas {
		if area.Name == s {
			return area, true
		}
	}
	return Area{}, false
}
