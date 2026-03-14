package main

type ReactionRule struct {
	ReactsWith     string   `json:"reactsWith"`
	ReactsWithArgs []string `json:"reactsWithArgs,omitempty"`
	Reaction       string   `json:"reaction"`
	ReactionArgs   []string `json:"reactionArgs,omitempty"`
}

type InteractableStateDescription struct {
	CssClass       string         `json:"cssClass,omitempty"`
	Pushable       bool           `json:"pushable,omitempty"`
	Walkable       bool           `json:"walkable,omitempty"`
	Fragile        bool           `json:"fragile,omitempty"`
	RejectTeleport bool           `json:"rejectTeleport,omitempty"`
	Reactions      string         `json:"reactions,omitempty"`
	ReactionRules  []ReactionRule `json:"reactionRules,omitempty"`
}

type InteractableDescription struct {
	ID             string                                  `json:"id"`
	Name           string                                  `json:"name"`
	SetName        string                                  `json:"setName"`
	State          string                                  `json:"state,omitempty"`
	DefaultState   string                                  `json:"defaultState,omitempty"`
	States         map[string]InteractableStateDescription `json:"states,omitempty"`
	CssClass       string                                  `json:"cssClass"`
	Pushable       bool                                    `json:"pushable"`
	Walkable       bool                                    `json:"walkable"`
	Fragile        bool                                    `json:"fragile"`
	RejectTeleport bool                                    `json:"rejectTeleport,omitempty"`
	Reactions      string                                  `json:"reactions"`
	ReactionRules  []ReactionRule                          `json:"reactionRules,omitempty"`
}
