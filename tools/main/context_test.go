package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestCompileSnap(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["snaps"]
	space := col.Spaces["toroid"]

	t.Run("Take baseline snapshots", func(t *testing.T) {
		desc0 := getAreaByName(space.Areas, "toroid:0-0")  // Ground + Protos
		desc1 := getAreaByName(space.Areas, "toroid:0-1")  // Interacables (2)
		desc2 := getAreaByName(space.Areas, "toroid:1-0")  // Uses BP for Fragment + Proto
		desc3 := getAreaByName(space.Areas, "toroid:1-1")  // Transports (1)
		desc4 := getAreaByName(space.Areas, "random-room") // 8x8 - empty

		snaps.MatchSnapshot(t, serializeForSnapshot(desc0, col))
		snaps.MatchSnapshot(t, serializeForSnapshot(desc1, col))
		snaps.MatchSnapshot(t, serializeForSnapshot(desc2, col))
		snaps.MatchSnapshot(t, serializeForSnapshot(desc3, col))
		snaps.MatchSnapshot(t, serializeForSnapshot(desc4, col))
	})

	t.Run("Snapshot after every grid action", func(t *testing.T) {
		area := getAreaByName(space.Areas, "random-room") // 8x8 - empty

		fillWithSand := makeClick(0, 0, "fill", "4")
		walledSection := makeClick(6, 1, "between", "6").withSelected(1, 7)
		sandWallGrass := makeClick(2, 2, "between", "2").withSelected(5, 7)

		innerwall0 := makeClick(4, 6, "between", "3").withSelected(1, 6)
		innerwall1 := makeClick(4, 6, "between", "3").withSelected(4, 3)
		innerwall2 := makeClick(2, 4, "between", "3").withSelected(4, 4)

		empty := makeClick(2, 7, "fill", "")

		// Toggle Ground pattern
		toggleFill := makeClick(6, 6, "toggle-fill", "")
		toggleBetween := makeClick(2, 3, "toggle-between", "").withSelected(1, 1)
		toggle := makeClick(5, 6, "toggle", "")

		// place on blueprint
		placeFrag := makeClick(6, 0, "place", "24106447-8d37-4b9d-bdf3-2df0104b4bc4")
		placeFrag2 := makeClick(6, 2, "place-blueprint", "24106447-8d37-4b9d-bdf3-2df0104b4bc4")
		placeProto := makeClick(5, 2, "place-blueprint", "4b-07")

		// rotate (not through bp) delete interactable
		rotateProto := makeClick(5, 2, "rotate", "")
		deleteInteractable := makeClick(6, 3, "interactable-delete", "")

		col.gridClickAction(fillWithSand, area.Blueprint)
		col.gridClickAction(walledSection, area.Blueprint)
		col.gridClickAction(sandWallGrass, area.Blueprint)
		col.gridClickAction(innerwall0, area.Blueprint)
		col.gridClickAction(innerwall1, area.Blueprint)
		col.gridClickAction(innerwall2, area.Blueprint)
		col.gridClickAction(empty, area.Blueprint)
		col.gridClickAction(toggleFill, area.Blueprint)
		col.gridClickAction(toggleBetween, area.Blueprint)
		col.gridClickAction(toggle, area.Blueprint)
		col.gridClickAction(placeFrag, area.Blueprint)
		col.gridClickAction(placeFrag2, area.Blueprint)
		col.gridClickAction(placeProto, area.Blueprint)
		col.gridClickAction(rotateProto, area.Blueprint)
		col.gridClickAction(deleteInteractable, area.Blueprint)

		snaps.MatchSnapshot(t, serializeForSnapshot(area, col))

	})

}

// -----------------------------------------------------------------------------
// Serialize for Snapshots
// -----------------------------------------------------------------------------

func serializeForSnapshot(area *AreaDescription, collection *Collection) string {
	if area == nil {
		return "NIL AREA DESCRIPTION"
	}
	return FormatAreaOutput(collection.areaOutputFromDescription(*area, "test-id"))
}

const areaOutputTemplate = `
Name: %s
--------------
Safe: %t	North: %s	South: %s	East: %s	West: %s
MapId: %s	LoadStrategy: %s	Spawnstrategy: %s	BroadcastGroup: %s	Weather: %s
--------------
# of Transports: %d
--------------`

func FormatAreaOutput(area AreaOutput) string {
	header := fmt.Sprintf(areaOutputTemplate,
		area.Name, area.Safe, area.North, area.South, area.East, area.West,
		area.MapId, area.LoadStrategy, area.SpawnStrategy, area.BroadcastGroup, area.Weather,
		len(area.Transports))

	return header + "\n" +
		FormatMaterialGrid(area.Tiles) + "\n" +
		FormatInteractableGrid(area.Interactables)
}

//////////////////////////////////////////////////////
// Material Grid

func FormatMaterialGrid(grid [][]Material) string {
	var b strings.Builder

	for r, row := range grid {
		var out [4][]string

		for _, m := range row {
			a, b1, b2, c := FormatMaterial(m)
			out[0] = append(out[0], a)
			out[1] = append(out[1], b1)
			out[2] = append(out[2], b2)
			out[3] = append(out[3], c)
		}

		for i := 0; i < len(out); i++ {
			b.WriteString(strings.Join(out[i], " | "))
			b.WriteByte('\n')
		}

		// Separator line (skip after last row)
		if r < len(grid)-1 {
			width := len(strings.Join(out[0], " | "))
			b.WriteString(strings.Repeat("-", width))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func FormatMaterial(m Material) (string, string, string, string) {
	const slotWidth = 15
	s1 := center(strconv.FormatBool(m.Walkable), slotWidth)

	css := chooseCss(m)
	if css == "" {
		css = "-" // visual placeholder
	}
	s2 := center(css, slotWidth)
	s3 := center("", slotWidth)
	if len(css) > slotWidth {
		s3 = center(css[slotWidth:], slotWidth)
	}

	s4 := center(strconv.FormatBool(m.DisplayText != ""), slotWidth)

	return s1, s2, s3, s4
}

func chooseCss(m Material) string {
	switch {
	case m.Ceiling2Css != "":
		return m.Ceiling2Css
	case m.Ceiling1Css != "":
		return m.Ceiling1Css
	case m.Floor2Css != "":
		return m.Floor2Css
	case m.Floor1Css != "":
		return m.Floor1Css
	case m.Ground2Css != "":
		return m.Ground2Css
	default:
		return m.Ground1Css
	}
}

//////////////////////////////////////////////////////
// Interactable Grid

func FormatInteractableGrid(grid [][]*InteractableDescription) string {
	var b strings.Builder

	for r, row := range grid {
		var out [3][]string // 3 rows per cell

		for _, i := range row {
			a, b1, c := FormatInteractable(i)
			out[0] = append(out[0], a)
			out[1] = append(out[1], b1)
			out[2] = append(out[2], c)
		}

		for i := 0; i < len(out); i++ {
			b.WriteString(strings.Join(out[i], " | "))
			b.WriteByte('\n')
		}

		// Separator line (skip after last row)
		if r < len(grid)-1 {
			width := len(strings.Join(out[0], " | "))
			b.WriteString(strings.Repeat("-", width))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func FormatInteractable(i *InteractableDescription) (string, string, string) {
	const slotWidth = 11
	if i == nil {
		return center("", slotWidth), center("-", slotWidth), center("", slotWidth)
	}
	// Row 0 – Name
	s1 := center(i.Name, slotWidth)

	// Row 1 – CssClass
	s2 := center(i.CssClass, slotWidth)

	// Row 2 – flags: Pushable Walkable Fragile Reactions?
	flags := fmt.Sprintf("%c %c %c %c",
		boolTF(i.Pushable),
		boolTF(i.Walkable),
		boolTF(i.Fragile),
		boolTF(i.Reactions != ""))

	s3 := center(flags, slotWidth)

	return s1, s2, s3
}

//////////////////////////////////////////////////////
// Helpers

func center(s string, slotWidth int) string {
	s = trim(s, slotWidth)
	left := (slotWidth - len(s)) / 2
	right := slotWidth - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func trim(s string, slotWidth int) string {
	if len(s) > slotWidth {
		return s[:slotWidth]
	}
	return s
}

func boolTF(b bool) rune {
	if b {
		return 'T'
	}
	return 'F'
}
