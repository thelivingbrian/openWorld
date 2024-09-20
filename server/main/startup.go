package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
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

////////////////////////////////////////////////////
// Configuration

type Configuration struct {
	envName     string
	port        string
	usesTLS     bool
	tlsCertPath string
	tlsKeyPath  string
	mongoHost   string
	mongoPort   string
	mongoUser   string
	mongoPass   string
}

func getConfiguration() *Configuration {
	environmentName := os.Getenv("BLOOP_ENV")

	if environmentName == "prod" {
		log.Fatal("No Prod Environment")
	} else if environmentName == "test" {
		return &Configuration{
			envName:     environmentName,
			port:        ":443",
			usesTLS:     true,
			tlsCertPath: os.Getenv("BLOOP_TLS_CERT_PATH"),
			tlsKeyPath:  os.Getenv("BLOOP_TLS_KEY_PATH"),
			mongoHost:   "localhost",
			mongoPort:   ":27017",
			mongoUser:   "",
			mongoPass:   "",
		}
	} else if environmentName == "dev" {
		return &Configuration{
			envName:     environmentName,
			port:        ":9090",
			usesTLS:     false,
			tlsCertPath: "./certificate/localhost.pem",
			tlsKeyPath:  "./certificate/localhost-key.pem",
			mongoHost:   "localhost",
			mongoPort:   ":27017",
			mongoUser:   "",
			mongoPass:   "",
		}
	}
	log.Fatal("No Configuration, exiting")
	return nil
}

func (config *Configuration) getMongoCredentialString() string {
	if config.mongoUser != "" && config.mongoPass != "" {
		return config.mongoUser + ":" + config.mongoPass + "@"
	}
	return ""
}

func (config *Configuration) getMongoURI() string {
	return "mongodb://" + config.getMongoCredentialString() + config.mongoHost + config.mongoPort
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
	SourceY   int    `json:"sourceY"`
	SourceX   int    `json:"sourceX"`
	DestY     int    `json:"destY"`
	DestX     int    `json:"destX"`
	DestStage string `json:"destStage"`
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
}

type InteractableDescription struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SetName     string `json:"setName"`
	CssClass    string `json:"cssClass"`
	Pushable    bool   `json:"pushable"`
	Destroyable bool   `json:"transformation,omitempty"`
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

	fmt.Printf("Loaded %d materials.", len(materials))
	fmt.Printf("Loaded %d areas.", len(areas))
}

func areaFromName(s string) (area Area, success bool) {
	for _, area := range areas {
		if area.Name == s {
			return area, true
		}
	}
	return Area{}, false
}
