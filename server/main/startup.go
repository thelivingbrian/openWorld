package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

///////////////////////////////////////////////////
//Database

type DB struct {
	users         *mongo.Collection
	playerRecords *mongo.Collection
	events        *mongo.Collection
	sessionData   *mongo.Collection
}

func createDbConnection(config *Configuration) *DB {
	mongodb := mongoClient(config).Database("bloopdb")
	return &DB{mongodb.Collection("users"), mongodb.Collection("players"), mongodb.Collection("events"), mongodb.Collection("sessionData")}
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
	logLevel           string
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
	rootDomain         string // root domain is used for cookie and CORS
	isHub              bool
	domains            []string
	serverName         string
	domainName         string
	loadPreviousState  bool
}

func getConfiguration() *Configuration {
	err := godotenv.Load()
	if err != nil {
		logger.Error().Err(err).Msg("Error loading .env file")
	}

	environmentName := os.Getenv("BLOOP_ENV")
	hashKey, blockKey := retrieveKeys()

	config := Configuration{
		envName:            environmentName,
		logLevel:           os.Getenv("LOG_LEVEL"),
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
		rootDomain:         os.Getenv("ROOT_DOMAIN"),
		domains:            strings.Split(os.Getenv("DOMAINS"), ","),
		serverName:         os.Getenv("SERVER_NAME"),
		domainName:         os.Getenv("DOMAIN_NAME"),
		loadPreviousState:  strings.ToUpper(os.Getenv("LOAD_PEVIOUS_STATE")) == "TRUE",
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
	return "mongodb" + config.mongoPrefix + "://" + config.getMongoCredentialString() + config.mongoHost + config.mongoPort
}

func (config *Configuration) createCookieStore() *sessions.CookieStore {
	if len(config.hashKey) != 32 || len(config.blockKey) != 32 {
		panic("Invalid key lengths")
	}
	store := sessions.NewCookieStore(config.hashKey, config.blockKey)
	store.Options = &sessions.Options{
		Domain:   "." + config.rootDomain, // Leading dot allows subdomains
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	return store
}

func (config *Configuration) isServer() bool {
	return config.serverName != "" && config.domainName != ""
}

func (config *Configuration) originForCORS() string {
	prefix := "http://"
	if config.usesTLS {
		prefix = "https://"
	}
	return prefix + config.rootDomain
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
	CommonName  string `json:"commonName"`
	Walkable    bool   `json:"walkable"`
	Ground1Css  string `json:"ground1css"`
	Ground2Css  string `json:"ground2css"`
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
	Name           string                       `json:"name"`
	Safe           bool                         `json:"safe"`
	Tiles          [][]Material                 `json:"tiles"`
	Transports     []Transport                  `json:"transports"`
	Interactables  [][]*InteractableDescription `json:"interactables"`
	North          string                       `json:"north"`
	South          string                       `json:"south"`
	East           string                       `json:"east"`
	West           string                       `json:"west"`
	MapId          string                       `json:"mapId"`
	LoadStrategy   string                       `json:"loadStrategy,omitempty"`
	SpawnStrategy  string                       `json:"spawnStrategy,omitempty"`
	BroadcastGroup string                       `json:"broadcastGroup,omitempty"`
	Weather        string                       `json:"weather,omitempty"`
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
	areas []Area
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
	populateStructUsingFileName(&areas, "areas")
}

func areaFromName(s string) (area Area, success bool) {
	for _, area := range areas {
		if area.Name == s {
			return area, true
		}
	}
	return Area{}, false
}

///////////////////////////////////////////////////////////////
// Set global log level

func setGlobalLogLevel(logLevel string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		// in Absence of a default, NoLevel is choosen.
		//    "" is a valid logLevel, it also produces NoLevel
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}
