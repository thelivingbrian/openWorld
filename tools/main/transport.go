package main

type Transport struct {
	SourceY            int    `json:"sourceY"`
	SourceX            int    `json:"sourceX"`
	DestY              int    `json:"destY"`
	DestX              int    `json:"destX"`
	DestStage          string `json:"destStage"`
	Confirmation       bool   `json:"confirmation"`
	RejectInteractable bool   `json:"rejectInteractable"`
}
