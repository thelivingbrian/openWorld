package main

import "github.com/google/uuid"

type GridClickDetails struct {
	CollectionName  string
	Location        []string
	GridType        string
	ScreenID        string
	Y               int
	X               int
	Tool            string
	SelectedAssetId string
	haveASelection  bool
	selectedX       int
	selectedY       int
}

type Coordinate struct {
	Y, X int
}

func (col *Collection) gridClickAction(details *GridClickDetails, blueprint *Blueprint) {
	switch details.Tool {
	case "select":
		gridSelect(details)
	case "replace":
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		gridReplace(details, blueprint.Tiles, selectedPrototype)
	case "fill":
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		gridFill(details, blueprint.Tiles, selectedPrototype)
	case "between":
		selectedPrototype := col.getPrototypeOrCreateInvalid(details.SelectedAssetId)
		gridFillBetween(details, blueprint.Tiles, selectedPrototype)
	case "place":
		fragment := col.getFragmentFromAssetId(details.SelectedAssetId)
		gridPlaceFragment(details, blueprint.Tiles, fragment)
	case "rotate":
		gridRotate(details, blueprint.Tiles)
	case "place-blueprint":
		gridPlaceOnBlueprint(details, blueprint)
		col.applyEveryInstruction(blueprint)
	case "interactable-replace":
		interactable := col.findInteractableById(details.SelectedAssetId)
		interactableReplace(details, blueprint.Tiles, interactable)
	case "interactable-delete":
		interactableReplace(details, blueprint.Tiles, nil)
	case "toggle":
		gridToggleGroundStatus(details, blueprint.Ground)
		impactedCells := SubGrid(blueprint.Ground, details.Y-1, details.X-1, 3, 3)
		smoothCorners(impactedCells)
	case "toggle-between":
		gridToggleBetween(details, blueprint.Ground)
	case "toggle-fill":
		gridToggleFill(details, blueprint.Ground, nil, -1)
		smoothCorners(blueprint.Ground)
	}
}

func pasteTiles(y, x int, source, dest [][]TileData) {
	for i := range dest {
		if y+i >= len(source) {
			break
		}
		for j := range dest[i] {
			if x+j >= len(source[y+i]) {
				break
			}
			if dest[i][j].PrototypeId != "" {
				source[y+i][x+j].PrototypeId = dest[i][j].PrototypeId
				source[y+i][x+j].Transformation = dest[i][j].Transformation
			}
			if dest[i][j].InteractableId != "" {
				source[y+i][x+j].InteractableId = dest[i][j].InteractableId
			}
		}
	}
}

func (col *Collection) applyEveryInstruction(blueprint *Blueprint) {
	for _, instruction := range blueprint.Instructions {
		col.applyInstruction(blueprint.Tiles, instruction)
	}
}

func (col *Collection) applyInstruction(source [][]TileData, instruction Instruction) {
	gridToApply := rotateTimesN(col.getTileGridByAssetId(instruction.GridAssetId), instruction.ClockwiseRotations)
	pasteTiles(instruction.Y, instruction.X, source, gridToApply)
}

func (col *Collection) getTileGridByAssetId(assetId string) [][]TileData {
	fragment := col.getFragmentById(assetId)
	if fragment != nil {
		return fragment.Blueprint.Tiles
	}
	out := make([][]TileData, 0)
	proto := col.findPrototypeById(assetId)
	if proto != nil {
		out = append(out, append(make([]TileData, 0), TileData{PrototypeId: assetId, Transformation: Transformation{}}))
	}
	return out
}

func gridPlaceFragment(details *GridClickDetails, modifications [][]TileData, selectedFragment Fragment) {
	for i := range selectedFragment.Blueprint.Tiles {
		if details.Y+i < len(modifications) {
			for j := range selectedFragment.Blueprint.Tiles[i] {
				if details.X+j < len(modifications[details.Y+i]) {
					modifications[details.Y+i][details.X+j] = selectedFragment.Blueprint.Tiles[i][j]
				}
			}
		}
	}
}

func gridSelect(event *GridClickDetails) {
	event.haveASelection = true
	event.selectedY, event.selectedX = event.Y, event.X
}

func gridReplace(event *GridClickDetails, modifications [][]TileData, selectedProto Prototype) {
	modifications[event.Y][event.X].PrototypeId = selectedProto.ID
}

func interactableReplace(event *GridClickDetails, modifications [][]TileData, selectedInteractable *InteractableDescription) {
	modifications[event.Y][event.X].InteractableId = ""
	if selectedInteractable != nil {
		modifications[event.Y][event.X].InteractableId = selectedInteractable.ID
	}
}

func gridFill(event *GridClickDetails, grid [][]TileData, selectedPrototype Prototype) {
	targetId := grid[event.Y][event.X].PrototypeId
	seen := make([][]bool, len(grid))
	for row := range seen {
		seen[row] = make([]bool, len(grid[row]))
	}
	fill(event, grid, selectedPrototype, seen, targetId)
}

func fill(event *GridClickDetails, modifications [][]TileData, selectedPrototype Prototype, seen [][]bool, targetId string) {
	seen[event.Y][event.X] = true
	modifications[event.Y][event.X].PrototypeId = selectedPrototype.ID
	deltas := []int{-1, 1}
	for _, i := range deltas {
		if event.Y+i >= 0 && event.Y+i < len(modifications) {
			shouldfill := !seen[event.Y+i][event.X] && modifications[event.Y+i][event.X].PrototypeId == targetId
			if shouldfill {
				newEvent := *event
				newEvent.Y += i
				fill(&newEvent, modifications, selectedPrototype, seen, targetId)
			}
		}
		if event.X+i >= 0 && event.X+i < len(modifications[event.Y]) {
			shouldfill := !seen[event.Y][event.X+i] && modifications[event.Y][event.X+i].PrototypeId == targetId
			if shouldfill {
				newEvent := *event
				newEvent.X += i
				fill(&newEvent, modifications, selectedPrototype, seen, targetId)
			}
		}
	}
}

func gridToggleFill(event *GridClickDetails, modifications [][]Cell, seen [][]bool, selectedStatus int) {
	if seen == nil {
		selectedStatus = modifications[event.Y][event.X].Status
		seen = make([][]bool, len(modifications))
		for row := range seen {
			seen[row] = make([]bool, len(modifications[row]))
		}
	}

	seen[event.Y][event.X] = true
	toggleCellStatus(&modifications[event.Y][event.X])

	deltas := []int{-1, 1}
	for _, i := range deltas {
		if event.Y+i >= 0 && event.Y+i < len(modifications) {
			shouldfill := !seen[event.Y+i][event.X] && modifications[event.Y+i][event.X].Status == selectedStatus
			if shouldfill {
				newEvent := *event
				newEvent.Y += i
				gridToggleFill(&newEvent, modifications, seen, selectedStatus)
			}
		}
		if event.X+i >= 0 && event.X+i < len(modifications[event.Y]) {
			shouldfill := !seen[event.Y][event.X+i] && modifications[event.Y][event.X+i].Status == selectedStatus
			if shouldfill {
				newEvent := *event
				newEvent.X += i
				gridToggleFill(&newEvent, modifications, seen, selectedStatus)
			}
		}
	}
}

func gridFillBetween(event *GridClickDetails, modifications [][]TileData, selectedPrototype Prototype) {
	if !event.haveASelection {
		gridSelect(event)
	}

	var lowx, lowy, highx, highy int
	if event.Y <= event.selectedY {
		lowy = event.Y
		highy = event.selectedY
	} else {
		lowy = event.selectedY
		highy = event.Y
	}
	if event.X <= event.selectedX {
		lowx = event.X
		highx = event.selectedX
	} else {
		lowx = event.selectedX
		highx = event.X
	}

	for i := lowy; i <= highy; i++ {
		for j := lowx; j <= highx; j++ {
			newEvent := *event
			newEvent.Y = i
			newEvent.X = j
			gridReplace(&newEvent, modifications, selectedPrototype)
		}
	}
	gridSelect(event)
}

func gridToggleBetween(event *GridClickDetails, modifications [][]Cell) {
	if !event.haveASelection {
		gridSelect(event)
	}

	var lowx, lowy, highx, highy int
	if event.Y <= event.selectedY {
		lowy = event.Y
		highy = event.selectedY
	} else {
		lowy = event.selectedY
		highy = event.Y
	}
	if event.X <= event.selectedX {
		lowx = event.X
		highx = event.selectedX
	} else {
		lowx = event.selectedX
		highx = event.X
	}

	for i := lowy; i <= highy; i++ {
		for j := lowx; j <= highx; j++ {
			newEvent := *event
			newEvent.Y = i
			newEvent.X = j
			gridToggleGroundStatus(&newEvent, modifications)
		}
	}
	impactedCells := SubGrid(modifications, lowy-1, lowx-1, highy-lowy+3, highx-lowx+3)
	smoothCorners(impactedCells)
	gridSelect(event)
}

func gridRotate(event *GridClickDetails, modifications [][]TileData) {
	transformation := &modifications[event.Y][event.X].Transformation
	transformation.ClockwiseRotations = mod(transformation.ClockwiseRotations+1, 4)
}

func gridPlaceOnBlueprint(event *GridClickDetails, blueprint *Blueprint) {
	if event.SelectedAssetId != "" {
		blueprint.Instructions = append(blueprint.Instructions, Instruction{
			ID:                 uuid.New().String(),
			X:                  event.X,
			Y:                  event.Y,
			GridAssetId:        event.SelectedAssetId,
			ClockwiseRotations: 0,
		})
	}
}

func gridToggleGroundStatus(event *GridClickDetails, modifications [][]Cell) {
	currentStatus := modifications[event.Y][event.X].Status
	modifications[event.Y][event.X].Status = (currentStatus + 1) % 2
}

func toggleCellStatus(cell *Cell) {
	currentStatus := cell.Status
	cell.Status = (currentStatus + 1) % 2
}
