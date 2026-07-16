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

	top := "┌" + strings.Repeat("─", innerWidth+2) + "┐"
	bottom := "└" + strings.Repeat("─", innerWidth+2) + "┘"
	boxed := make([]string, 0, len(lines)+2)
	boxed = append(boxed, top)
	for _, line := range lines {
		boxed = append(boxed, "│ "+termwidth.RightPad(line, innerWidth)+" │")
	}
	boxed = append(boxed, bottom)
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

func JoinHorizontal(left []string, right []string, gap int) []string {
	leftWidth := blockWidth(left)
	rightWidth := blockWidth(right)
	height := max(len(left), len(right))

	separator := strings.Repeat(" ", gap)
	joined := make([]string, 0, height)
	for index := range height {
		leftLine := paddedLine(left, index, leftWidth)
		rightLine := paddedLine(right, index, rightWidth)
		joined = append(joined, leftLine+separator+rightLine)
	}
	return joined
}

func blockWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if lineWidth := termwidth.Width(line); lineWidth > width {
			width = lineWidth
		}
	}
	return width
}

func paddedLine(lines []string, index int, width int) string {
	if index >= len(lines) {
		return strings.Repeat(" ", width)
	}
	return termwidth.RightPad(lines[index], width)
}
