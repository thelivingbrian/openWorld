package main

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////
// Visual Fx

func flashBackgroundColorIfTangible(player *Player, color string) {
	ownLock := player.tangibilityLock.TryLock() // Try because what if this player is currently intangible due to event damaging the initiator of this flash
	if !ownLock {
		return
	}
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}
	flashBackgroundColor(player, color)
}

func flashBackgroundColor(player *Player, color string) {
	script := fmt.Sprintf(`<div id="script"><script>flashBg("%s")</script></div>`, color)
	updateOne(script, player)
}

func changePageBackgroundColor(player *Player, bgColor string) {
	script := fmt.Sprintf(`<div id="script"><script>document.body.className="%s"</script></div>`, bgColor)
	updateOne(script, player)
}

func makeHallucinate(player *Player) {
	go func() {
		ownLock := player.tangibilityLock.TryLock()
		if !ownLock {
			return
		}
		defer player.tangibilityLock.Unlock()
		if !player.tangible {
			return
		}
		for i := 0; i <= 80; i++ {
			time.Sleep(20 * time.Millisecond)
			updateOne(generateDivs(i), player)
		}
	}()
}

func generateDivs(frame int) string {
	var sb strings.Builder

	// Define the center of the grid
	center := 7.5

	// Determine which color set to use based on the frame
	var col1, col2 string
	if ((frame / 20) % 2) == 0 {
		// Use the first color set
		col1, col2 = "red trsp40", "gold trsp40"
	} else {
		// Use the second color set
		col1, col2 = "blue trsp40", "green trsp40"
	}

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			dx := float64(i) - center
			dy := float64(j) - center

			// Radius from the center
			r := math.Sqrt(dx*dx + dy*dy)

			// Base angle in radians, range (-π, π]
			angle := math.Atan2(dy, dx)

			// Add rotation based on the frame
			angle += float64(frame) * 0.1

			// Determine pattern: If this value is even, use col1; if odd, use col2
			// Multiplying angle by r gives a spiral-like indexing pattern.
			colorIndex := int((angle * r))

			var color string
			if colorIndex%2 == 0 {
				color = col1
			} else {
				color = col2
			}

			sb.WriteString(fmt.Sprintf(`[~ id="Lw1" y="%d" x="%d" class="box zw %s"]`+"\n", i, j, color))
		}
	}

	return sb.String()
}

func oldFx(player *Player) {
	go func() {
		ownLock := player.tangibilityLock.TryLock()
		if !ownLock {
			return
		}
		defer player.tangibilityLock.Unlock()
		if !player.tangible {
			return
		}
		updateOne(generateDivs2("blue", 0, 2), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("blue", 1, 2), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("blue", 0, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("blue", 1, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("blue", 2, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 0, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 1, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 2, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 3, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 0, 5), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 1, 5), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("green", 2, 5), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 0, 2), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 1, 2), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 0, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 1, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 2, 3), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 0, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 1, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 2, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 3, 4), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 0, 5), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 1, 5), player)
		time.Sleep(20 * time.Millisecond)
		updateOne(generateDivs2("red", 2, 5), player)
	}()
}

func generateDivs2(color string, check int, check2 int) string {
	var sb strings.Builder

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			setColor := color
			if (i+j)%check2 == check {
				setColor = "ice"
			}
			sb.WriteString(fmt.Sprintf(`[~ id="Lw1" y="%d" x="%d" class="box zw %s"]`+"\n", i, j, setColor))
		}
	}

	return sb.String()
}

/////////////////////////////////////////////////////////////////////////
// Everything below this point moderately invalid?

func generateWeatherSolid(color string) string {
	var sb strings.Builder

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			sb.WriteString(fmt.Sprintf(`[~ id="w" y="%d" x="%d" class="box zw %s"]`+"\n", i, j, color))
		}
	}

	return sb.String()
}

func generateWeatherSolidByteBuffer(color string) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 12000))

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			fmt.Fprintf(buf, `[~ id="w" y="%d" x="%d" class="box zw %s"]`, i, j, color)
		}
	}

	return buf.Bytes()
}

func generateWeatherDumb(color string) string {
	out := ""

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			out += fmt.Sprintf(`[~ id="w" y="%d" x="%d" class="box zw %s"]`, i, j, color)
		}
	}

	return out
}

func generateWeatherSolidBytes(color string) []byte {
	const rows = 16
	const cols = 16

	// Estimate capacity to avoid growth:
	// Each element has a pattern:
	// <div id="w{i}-{j}" class="box zw {color}"></div>\n
	//
	// Breakdown of constant parts:
	// `[~ id="Lw1-`      = 11 bytes (including the quote)
	// `-`                = 1 byte
	// `" class="box zw ` = 16 bytes (including the leading quote)
	// `"]`       = 2 bytes
	// Total constant overhead per line = 30 bytes
	//
	// Plus length of i and j = worst case of 34 bytes? + color
	//
	// We have 256 lines (16x16):
	// capacity ~ 256 * (34 + len(color))
	estCap := 256 * (40 + len(color))
	b := make([]byte, 0, estCap)

	prefix := []byte(`[~ id="Lw1-`)
	sep := []byte(`-`)
	cls := []byte(`" class="box zw `)
	suffix := []byte(`"]`)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			b = append(b, prefix...)
			b = strconv.AppendInt(b, int64(i), 10)
			b = append(b, sep...)
			b = strconv.AppendInt(b, int64(j), 10)
			b = append(b, cls...)
			b = append(b, color...)
			b = append(b, suffix...)
		}
	}

	return b
}
