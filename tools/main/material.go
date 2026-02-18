package main

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

type Color struct {
	CssClassName string `json:"cssClassName"`
	R            int    `json:"R"`
	G            int    `json:"G"`
	B            int    `json:"B"`
	A            string `json:"A"`
}
