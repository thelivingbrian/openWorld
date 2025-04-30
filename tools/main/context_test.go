package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestCompileSnap(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["bloop"] // change to snaps

	t.Run("Test make empty grid for TestGridActions", func(t *testing.T) {
		space := col.Spaces["rooms"]
		desc := getAreaByName(space.Areas, "sandy")
		mat, err := col.compileMaterialsFromBlueprint(desc.Blueprint)
		if err != nil {
			panic("error with compile for test")
		}
		fmt.Println(FormatGrid(mat))
	})

}

const slotWidth = 15 // fixed width for every cell

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

func FormatGrid(grid [][]Material) string {
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
	s1 := center(strconv.FormatBool(m.Walkable))

	css := chooseCss(m)
	if css == "" {
		css = "-" // visual placeholder
	}
	s2 := center(css)
	s3 := center("")
	if len(css) > slotWidth {
		s3 = center(css[slotWidth:])
	}

	s4 := center(strconv.FormatBool(m.DisplayText != ""))

	return s1, s2, s3, s4
}

func chooseCss(m Material) string {
	switch {
	case m.Ceiling2Css != "":
		return m.Ceiling2Css
	case m.Ceiling1Css != "":
		return m.Ceiling1Css
	case m.Ground2Css != "":
		return m.Ground2Css
	case m.Ground1Css != "":
		return m.Ground1Css
	case m.Floor2Css != "":
		return m.Floor2Css
	default:
		return m.Floor1Css
	}
}

func center(s string) string { // pads/trim to 15 runes
	if len(s) >= slotWidth {
		return s[:slotWidth]
	}
	left := (slotWidth - len(s)) / 2
	right := slotWidth - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
