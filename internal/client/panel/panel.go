package panel

import (
	"PMud/internal/client/termwidth"
	"strings"
)

func BoxLines(lines []string, contentWidth int) []string {
	innerWidth := contentWidth
	for _, line := range lines {
		if width := termwidth.Width(line); width > innerWidth {
			innerWidth = width
		}
	}

	border := "+" + strings.Repeat("-", innerWidth+2) + "+"
	boxed := make([]string, 0, len(lines)+2)
	boxed = append(boxed, border)
	for _, line := range lines {
		boxed = append(boxed, "| "+termwidth.RightPad(line, innerWidth)+" |")
	}
	boxed = append(boxed, border)
	return boxed
}
