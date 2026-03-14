package main

type AreaDescription struct {
	Name           string      `json:"name"`
	Safe           bool        `json:"safe"`
	Blueprint      *Blueprint  `json:"blueprint"`
	Transports     []Transport `json:"transports"`
	North          string      `json:"north,omitempty"`
	South          string      `json:"south,omitempty"`
	East           string      `json:"east,omitempty"`
	West           string      `json:"west,omitempty"`
	MapId          string      `json:"mapId"`
	LoadStrategy   string      `json:"loadStrategy"`
	SpawnStrategy  string      `json:"spawnStrategy"`
	BroadcastGroup string      `json:"broadcastGroup,omitempty"`
	Weather        string      `json:"weather,omitempty"`
}

/////////////////////////////////
// Output (Compiled from description - must match engine)

type AreaOutput struct {
	Name           string                       `json:"name"`
	Safe           bool                         `json:"safe"`
	Tiles          [][]Material                 `json:"tiles"`
	Interactables  [][]*InteractableDescription `json:"interactables"`
	Transports     []Transport                  `json:"transports"`
	North          string                       `json:"north,omitempty"`
	South          string                       `json:"south,omitempty"`
	East           string                       `json:"east,omitempty"`
	West           string                       `json:"west,omitempty"`
	MapId          string                       `json:"mapId,omitempty"`
	LoadStrategy   string                       `json:"loadStrategy,omitempty"`
	SpawnStrategy  string                       `json:"spawnStrategy"`
	BroadcastGroup string                       `json:"broadcastGroup,omitempty"`
	Weather        string                       `json:"weather,omitempty"`
}

type Material struct {
	Walkable    bool   `json:"walkable,omitempty"`
	Ground1Css  string `json:"ground1css,omitempty"`
	Ground2Css  string `json:"ground2css,omitempty"`
	Floor1Css   string `json:"layer1css,omitempty"`
	Floor2Css   string `json:"layer2css,omitempty"`
	Ceiling1Css string `json:"ceiling1css,omitempty"`
	Ceiling2Css string `json:"ceiling2css,omitempty"`
	DisplayText string `json:"displayText,omitempty"`
}

///////////////////////////////////////////
// Blueprint

type Blueprint struct {
	Tiles             [][]TileData `json:"tiles"`
	Instructions      []Instruction
	Ground            [][]Cell
	DefaultTileColor  string
	DefaultTileColor1 string
}

type TileData struct {
	PrototypeId       string         `json:"prototypeId,omitempty"`
	Transformation    Transformation `json:"transformation,omitempty"`
	InteractableId    string         `json:"interactableId,omitempty"`
	InteractableState string         `json:"interactableState,omitempty"`
}

type Transformation struct {
	ClockwiseRotations int `json:"clockwiseRotations,omitempty"`
}

type Instruction struct {
	ID                 string
	X                  int
	Y                  int
	GridAssetId        string
	ClockwiseRotations int
}

type Cell struct {
	Status                                     int
	BottomRight, BottomLeft, TopRight, TopLeft bool
}

/////////////////////////////////////
//  Transports

type Transport struct {
	SourceY            int    `json:"sourceY"`
	SourceX            int    `json:"sourceX"`
	DestY              int    `json:"destY"`
	DestX              int    `json:"destX"`
	DestStage          string `json:"destStage"`
	Confirmation       bool   `json:"confirmation"`
	RejectInteractable bool   `json:"rejectInteractable"`
}
