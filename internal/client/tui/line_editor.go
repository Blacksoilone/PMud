package tui

import (
	"strings"
	"unicode/utf8"
)

const defaultHistoryLimit = 100

type LineEditor struct {
	text         []rune
	cursor       int
	anchor       string
	history      []string
	historyIndex int
	historyLimit int
}

func NewLineEditor(historyLimit int) LineEditor {
	if historyLimit <= 0 {
		historyLimit = defaultHistoryLimit
	}
	return LineEditor{historyLimit: historyLimit, historyIndex: -1}
}

func (e LineEditor) String() string { return string(e.text) }

func (e LineEditor) Cursor() int { return e.cursor }

func (e LineEditor) RenderWithCursor() string {
	const inverse = "\x1b[7m"
	const reset = "\x1b[0m"
	var builder strings.Builder
	for index, char := range e.text {
		if index == e.cursor {
			builder.WriteString(inverse)
			builder.WriteRune(char)
			builder.WriteString(reset)
			continue
		}
		builder.WriteRune(char)
	}
	if e.cursor == len(e.text) {
		builder.WriteString(inverse)
		builder.WriteByte(' ')
		builder.WriteString(reset)
	}
	return builder.String()
}

func (e LineEditor) Insert(value string) LineEditor {
	value = validUTF8(value)
	if value == "" {
		return e
	}
	text := append([]rune(nil), e.text...)
	insert := []rune(value)
	text = append(text, make([]rune, len(insert))...)
	copy(text[e.cursor+len(insert):], text[e.cursor:])
	copy(text[e.cursor:], insert)
	e.text = text
	e.cursor += len(insert)
	e = e.resetHistoryCursor()
	return e
}

func (e LineEditor) Backspace() LineEditor {
	if e.cursor == 0 {
		return e
	}
	e.text = append(append([]rune(nil), e.text[:e.cursor-1]...), e.text[e.cursor:]...)
	e.cursor--
	e = e.resetHistoryCursor()
	return e
}

func (e LineEditor) Delete() LineEditor {
	if e.cursor >= len(e.text) {
		return e
	}
	e.text = append(append([]rune(nil), e.text[:e.cursor]...), e.text[e.cursor+1:]...)
	e = e.resetHistoryCursor()
	return e
}

func (e LineEditor) MoveLeft() LineEditor {
	if e.cursor > 0 {
		e.cursor--
	}
	return e
}

func (e LineEditor) MoveRight() LineEditor {
	if e.cursor < len(e.text) {
		e.cursor++
	}
	return e
}

func (e LineEditor) Home() LineEditor { e.cursor = 0; return e }

func (e LineEditor) End() LineEditor { e.cursor = len(e.text); return e }

func (e LineEditor) Clear() LineEditor {
	e.text = nil
	e.cursor = 0
	e = e.resetHistoryCursor()
	return e
}

func (e LineEditor) Replace(value string) LineEditor {
	e.text = []rune(validUTF8(value))
	e.cursor = len(e.text)
	e = e.resetHistoryCursor()
	return e
}

func (e LineEditor) Submit(line string) LineEditor {
	if line != "" {
		history := append([]string(nil), e.history...)
		history = append(history, line)
		if len(history) > e.historyLimit {
			history = history[len(history)-e.historyLimit:]
		}
		e.history = history
	}
	e.text = nil
	e.cursor = 0
	e.anchor = ""
	e.historyIndex = -1
	return e
}

func (e LineEditor) HistoryPrevious() LineEditor {
	if len(e.history) == 0 {
		return e
	}
	if e.historyIndex == -1 {
		e.anchor = e.String()
		e.historyIndex = len(e.history) - 1
	} else if e.historyIndex > 0 {
		e.historyIndex--
	}
	e.text = []rune(e.history[e.historyIndex])
	e.cursor = len(e.text)
	return e
}

func (e LineEditor) HistoryNext() LineEditor {
	if e.historyIndex == -1 {
		return e
	}
	if e.historyIndex < len(e.history)-1 {
		e.historyIndex++
		e.text = []rune(e.history[e.historyIndex])
	} else {
		e.historyIndex = -1
		e.text = []rune(e.anchor)
		e.anchor = ""
	}
	e.cursor = len(e.text)
	return e
}

func (e LineEditor) resetHistoryCursor() LineEditor {
	e.historyIndex = -1
	e.anchor = ""
	return e
}

func validUTF8(value string) string {
	if utf8.ValidString(value) {
		return value
	}
	return string([]rune(value))
}
