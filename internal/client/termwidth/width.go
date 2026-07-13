package termwidth

import "strings"

func Width(text string) int {
	width := 0
	for _, char := range text {
		width += runeWidth(char)
	}
	return width
}

func LineWidth(text string) int {
	maxWidth := 0
	for line := range strings.SplitSeq(text, "\n") {
		width := Width(line)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

func RightPad(text string, targetWidth int) string {
	width := Width(text)
	if width >= targetWidth {
		return text
	}
	return text + strings.Repeat(" ", targetWidth-width)
}

func runeWidth(char rune) int {
	if isWideRune(char) {
		return 2
	}
	return 1
}

func isWideRune(char rune) bool {
	return (char >= 0x1100 && char <= 0x115F) ||
		(char >= 0x2E80 && char <= 0xA4CF) ||
		(char >= 0xAC00 && char <= 0xD7A3) ||
		(char >= 0xF900 && char <= 0xFAFF) ||
		(char >= 0xFE10 && char <= 0xFE19) ||
		(char >= 0xFE30 && char <= 0xFE6F) ||
		(char >= 0xFF00 && char <= 0xFF60) ||
		(char >= 0xFFE0 && char <= 0xFFE6)
}
