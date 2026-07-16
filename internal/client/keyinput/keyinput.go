package keyinput

import (
	"unicode"
	"unicode/utf8"

	"PMud/internal/client/tui"
)

type Action struct {
	Input tui.Input
	Quit  bool
}

func Decode(data []byte) []Action {
	actions := make([]Action, 0, len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			data = data[size:]
			continue
		}
		if action, ok := decodeRune(r); ok {
			actions = append(actions, action)
		}
		data = data[size:]
	}
	return actions
}

func decodeRune(r rune) (Action, bool) {
	switch r {
	case '\n', '\r':
		return Action{Input: tui.Input{Kind: tui.InputSubmit}}, true
	case '\b', '\x7f':
		return Action{Input: tui.Input{Kind: tui.InputBackspace}}, true
	case '\x15':
		return Action{Input: tui.Input{Kind: tui.InputClear}}, true
	case '\x12':
		return Action{Input: tui.Input{Kind: tui.InputForceRedraw}}, true
	case '\x03':
		return Action{Quit: true}, true
	}
	if !unicode.IsPrint(r) {
		return Action{}, false
	}
	return Action{Input: tui.Input{Kind: tui.InputText, Text: string(r)}}, true
}
