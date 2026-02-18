package main

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
