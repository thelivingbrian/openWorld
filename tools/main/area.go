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
