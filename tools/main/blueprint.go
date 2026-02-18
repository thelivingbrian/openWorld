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

func clearTiles(y, x, height, width int, source [][]TileData) {
	for i := 0; i < height; i++ {
		if y+i >= len(source) {
			break
		}
		for j := 0; j < width; j++ {
			if x+j >= len(source[y+i]) {
				break
			}
			source[y+i][x+j].PrototypeId = ""
			source[y+i][x+j].InteractableId = ""
		}
	}
}

func rotateTimesN(input [][]TileData, n int) [][]TileData {
	rotations := mod(n, 4)
	out := input
	for i := 0; i < rotations; i++ {
		out = rotateClockwise(out)
		for y := range out {
			for x := range out[y] {
				out[y][x].Transformation.ClockwiseRotations++
			}
		}
	}
	return out
}

func rotateClockwise[T any](input [][]T) [][]T {
	outheight := len(input[0])
	out := make([][]T, outheight)
	for i := 0; i < outheight; i++ {
		out[i] = make([]T, len(input))
		for j := 0; j < len(input); j++ {
			out[i][j] = input[len(input)-j-1][i]
		}
	}
	return out
}

func generateMaterialsForGround(bp *Blueprint) [][]Material {
	if bp.Ground == nil {
		bp.Ground = make([][]Cell, len(bp.Tiles))
		for n := range bp.Ground {
			bp.Ground[n] = make([]Cell, len(bp.Tiles[n]))
		}
	}
	out := make([][]Material, len(bp.Ground))
	for i := range bp.Ground {
		out[i] = make([]Material, len(bp.Ground[i]))
		for j := range bp.Ground[i] {
			out[i][j] = createMaterialForGround(bp.Ground[i][j], bp.DefaultTileColor, bp.DefaultTileColor1)
		}
	}
	return out
}

func createMaterialForGround(cell Cell, color0, color1 string) Material {
	material := Material{Ground1Css: "", Ground2Css: ""}
	return addGroundToMaterial(material, &cell, color0, color1)
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

func SubGrid(grid [][]Cell, y, x, height, width int) [][]Cell {
	rowStart := max(0, y)
	rowEnd := min(y+height, len(grid))
	colStart := max(0, x)
	colEnd := min(x+width, len(grid[0]))

	sub := make([][]Cell, 0, rowEnd-rowStart)
	for r := rowStart; r < rowEnd; r++ {
		rowSlice := grid[r][colStart:colEnd]
		sub = append(sub, rowSlice)
	}
	return sub
}
