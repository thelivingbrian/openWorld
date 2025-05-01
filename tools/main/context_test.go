package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestCompileSnap(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["snaps"] // change to snaps

	t.Run("Test make empty grid for TestGridActions", func(t *testing.T) {
		space := col.Spaces["toroid"]
		desc := getAreaByName(space.Areas, "toroid:0-1")

		fmt.Println(FormatAreaOutput(col.areaOutputFromDescription(*desc, "fakeid")))
	})

}

// -----------------------------------------------------------------------------
// Serialize for Snapshots
// -----------------------------------------------------------------------------

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
