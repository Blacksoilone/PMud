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

func RenderLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func EqualWidths(lines []string) bool {
	if len(lines) == 0 {
		return true
	}
	wantWidth := termwidth.Width(lines[0])
	for _, line := range lines[1:] {
		if termwidth.Width(line) != wantWidth {
			return false
		}
	}
	return true
}
