package main

type Blueprint struct {
	Tiles             [][]TileData `json:"tiles"`
	Instructions      []Instruction
	Ground            [][]Cell
	DefaultTileColor  string
	DefaultTileColor1 string
}

type TileData struct {
	PrototypeId    string         `json:"prototypeId,omitempty"`
	Transformation Transformation `json:"transformation,omitempty"`
	InteractableId string         `json:"interactableId,omitempty"`
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

func addGroundToMaterial(material Material, cell *Cell, color0, color1 string) Material {
	if cell == nil {
		material.Ground1Css = color0
		return material
	}
	if material.Ground2Css != "" {
		return material
	}
	primary, secondary := color0, color1
	if cell.Status != 0 {
		primary, secondary = color1, color0
	}
	material.Ground2Css = primary
	if cell.TopLeft || cell.TopRight || cell.BottomLeft || cell.BottomRight {
		material.Ground1Css = secondary
	}
	if cell.TopLeft {
		material.Ground2Css += " r0-tl"
	}
	if cell.TopRight {
		material.Ground2Css += " r0-tr"
	}
	if cell.BottomLeft {
		material.Ground2Css += " r0-bl"
	}
	if cell.BottomRight {
		material.Ground2Css += " r0-br"
	}
	return material
}
