package main

type ReactionRule struct {
	ReactsWith     string   `json:"reactsWith"`
	ReactsWithArgs []string `json:"reactsWithArgs,omitempty"`
	Reaction       string   `json:"reaction"`
	ReactionArgs   []string `json:"reactionArgs,omitempty"`
}

type InteractableDescription struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	SetName       string         `json:"setName"`
	CssClass      string         `json:"cssClass"`
	Pushable      bool           `json:"pushable"`
	Walkable      bool           `json:"walkable"`
	Fragile       bool           `json:"fragile"`
	Reactions     string         `json:"reactions"`
	ReactionRules []ReactionRule `json:"reactionRules,omitempty"`
}
