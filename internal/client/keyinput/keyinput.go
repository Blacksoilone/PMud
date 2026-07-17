package keyinput

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"PMud/internal/client/tui"
)

type Action struct {
	Input tui.Input
	Quit  bool
}

type Decoder struct {
	pending []byte
}

func (d *Decoder) HasStandaloneEscape() bool {
	return len(d.pending) == 1 && d.pending[0] == 0x1b
}

func (d *Decoder) Feed(data []byte) []Action {
	d.pending = append(d.pending, data...)
	actions := make([]Action, 0, len(d.pending))
	for len(d.pending) > 0 {
		if d.pending[0] == 0x1b {
			consumed, action, complete := decodeEscape(d.pending)
			if !complete {
				break
			}
			d.pending = d.pending[consumed:]
			if action != nil {
				actions = append(actions, *action)
			}
			continue
		}
		action, consumed, complete := decodeOne(d.pending)
		if !complete {
			break
		}
		d.pending = d.pending[consumed:]
		if action != nil {
			actions = append(actions, *action)
		}
	}
	return actions
}

func (d *Decoder) Flush() []Action {
	if len(d.pending) == 0 {
		return nil
	}
	if d.pending[0] == 0x1b {
		d.pending = nil
		return []Action{{Input: tui.Input{Kind: tui.InputCancel}}}
	}
	actions := Decode(d.pending)
	d.pending = nil
	return actions
}

func Decode(data []byte) []Action {
	var decoder Decoder
	actions := decoder.Feed(data)
	return append(actions, decoder.Flush()...)
}

func decodeOne(data []byte) (*Action, int, bool) {
	if len(data) == 0 {
		return nil, 0, false
	}
	r, size := utf8.DecodeRune(data)
	if r == utf8.RuneError && size == 1 {
		return nil, 1, true
	}
	action, ok := decodeRune(r)
	if !ok {
		return nil, size, true
	}
	return &action, size, true
}

func decodeEscape(data []byte) (int, *Action, bool) {
	if len(data) == 1 {
		return 0, nil, false
	}
	if data[1] == 'O' {
		if len(data) < 3 {
			return 0, nil, false
		}
		if data[2] == 'P' {
			return 3, &Action{Input: tui.Input{Kind: tui.InputOpenHelp}}, true
		}
		return 3, nil, true
	}
	if data[1] != '[' {
		return 1, &Action{Input: tui.Input{Kind: tui.InputCancel}}, true
	}
	for index := 2; index < len(data); index++ {
		if data[index] >= 0x40 && data[index] <= 0x7e {
			return index + 1, decodeCSI(data[2 : index+1]), true
		}
	}
	return 0, nil, false
}

func decodeCSI(sequence []byte) *Action {
	input := tui.Input{}
	switch string(sequence) {
	case "A":
		input.Kind = tui.InputHistoryPrevious
	case "B":
		input.Kind = tui.InputHistoryNext
	case "C":
		input.Kind = tui.InputMoveRight
	case "D":
		input.Kind = tui.InputMoveLeft
	case "H", "1~", "7~":
		input.Kind = tui.InputHome
	case "F", "4~", "8~":
		input.Kind = tui.InputEnd
	case "3~":
		input.Kind = tui.InputDelete
	case "5~":
		input.Kind = tui.InputPageUp
	case "6~":
		input.Kind = tui.InputPageDown
	case "11~":
		input.Kind = tui.InputOpenHelp
	default:
		sequenceText := string(sequence)
		switch {
		case strings.HasPrefix(sequenceText, "<64;") && strings.HasSuffix(sequenceText, "M"):
			input.Kind = tui.InputScrollUp
		case strings.HasPrefix(sequenceText, "<65;") && strings.HasSuffix(sequenceText, "M"):
			input.Kind = tui.InputScrollDown
		default:
			return nil
		}
	}
	return &Action{Input: input}
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
		return Action{}, false
	case '?':
		return Action{Input: tui.Input{Kind: tui.InputOpenHelp}}, true
	}
	if !unicode.IsPrint(r) {
		return Action{}, false
	}
	return Action{Input: tui.Input{Kind: tui.InputText, Text: string(r)}}, true
}
