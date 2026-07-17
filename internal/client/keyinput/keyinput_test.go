package keyinput_test

import (
	"testing"

	"PMud/internal/client/keyinput"
	"PMud/internal/client/tui"
)

func TestDecodeASCIIText(t *testing.T) {
	actions := keyinput.Decode([]byte("get"))

	assertInputText(t, actions, []string{"g", "e", "t"})
}

func TestDecodeCJKText(t *testing.T) {
	actions := keyinput.Decode([]byte("旧灯"))

	assertInputText(t, actions, []string{"旧", "灯"})
}

func TestDecodeEnterSubmits(t *testing.T) {
	actions := keyinput.Decode([]byte{'\n', '\r'})

	if len(actions) != 2 {
		t.Fatalf("actions length = %d, want 2", len(actions))
	}
	for index, action := range actions {
		if action.Input.Kind != tui.InputSubmit {
			t.Fatalf("actions[%d].Input.Kind = %v, want InputSubmit", index, action.Input.Kind)
		}
	}
}

func TestDecodeBackspace(t *testing.T) {
	actions := keyinput.Decode([]byte{0x7f, '\b'})

	if len(actions) != 2 {
		t.Fatalf("actions length = %d, want 2", len(actions))
	}
	for index, action := range actions {
		if action.Input.Kind != tui.InputBackspace {
			t.Fatalf("actions[%d].Input.Kind = %v, want InputBackspace", index, action.Input.Kind)
		}
	}
}

func TestDecodeCtrlUClear(t *testing.T) {
	actions := keyinput.Decode([]byte{0x15})

	if len(actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(actions))
	}
	if actions[0].Input.Kind != tui.InputClear {
		t.Fatalf("Input.Kind = %v, want InputClear", actions[0].Input.Kind)
	}
}

func TestDecodeCtrlRForceRedraw(t *testing.T) {
	actions := keyinput.Decode([]byte{0x12})

	if len(actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(actions))
	}
	if actions[0].Input.Kind != tui.InputForceRedraw {
		t.Fatalf("Input.Kind = %v, want InputForceRedraw", actions[0].Input.Kind)
	}
}

func TestDecodeCtrlCIgnoresInterrupt(t *testing.T) {
	if actions := keyinput.Decode([]byte{0x03}); len(actions) != 0 {
		t.Fatalf("actions = %#v, want none", actions)
	}
}

func TestDecodeEscapeSequencesDoNotEnterInput(t *testing.T) {
	data := []byte("\x1b[<64;12;4M")

	if actions := keyinput.Decode(data); len(actions) != 1 || actions[0].Input.Kind != tui.InputScrollUp {
		t.Fatalf("actions = %#v, want one mouse-scroll action", actions)
	}
}

func TestDecodeEditingKeys(t *testing.T) {
	tests := []struct {
		name string
		data string
		want tui.InputKind
	}{
		{name: "up history", data: "\x1b[A", want: tui.InputHistoryPrevious},
		{name: "down history", data: "\x1b[B", want: tui.InputHistoryNext},
		{name: "right", data: "\x1b[C", want: tui.InputMoveRight},
		{name: "left", data: "\x1b[D", want: tui.InputMoveLeft},
		{name: "home", data: "\x1b[H", want: tui.InputHome},
		{name: "end", data: "\x1b[F", want: tui.InputEnd},
		{name: "delete", data: "\x1b[3~", want: tui.InputDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := keyinput.Decode([]byte(tt.data))
			if len(actions) != 1 || actions[0].Input.Kind != tt.want {
				t.Fatalf("actions = %#v, want one input kind %v", actions, tt.want)
			}
		})
	}
}

func TestDecodePopupControlKeys(t *testing.T) {
	tests := []struct {
		name string
		data string
		kind tui.InputKind
	}{
		{name: "question mark", data: "?", kind: tui.InputOpenHelp},
		{name: "F1 CSI", data: "\x1b[11~", kind: tui.InputOpenHelp},
		{name: "F1 SS3", data: "\x1bOP", kind: tui.InputOpenHelp},
		{name: "page up", data: "\x1b[5~", kind: tui.InputPageUp},
		{name: "page down", data: "\x1b[6~", kind: tui.InputPageDown},
		{name: "wheel up", data: "\x1b[<64;10;4M", kind: tui.InputScrollUp},
		{name: "wheel down", data: "\x1b[<65;10;4M", kind: tui.InputScrollDown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actions := keyinput.Decode([]byte(test.data))
			if len(actions) != 1 || actions[0].Input.Kind != test.kind {
				t.Fatalf("Decode(%q) = %#v, want one action kind %v", test.data, actions, test.kind)
			}
		})
	}
}

func TestDecodeStandaloneEscapeCancelsInput(t *testing.T) {
	actions := keyinput.Decode([]byte{0x1b})

	if len(actions) != 1 || actions[0].Input.Kind != tui.InputCancel {
		t.Fatalf("actions = %#v, want one InputCancel", actions)
	}
}

func TestDecoderBuffersSplitCSISequence(t *testing.T) {
	var decoder keyinput.Decoder

	if actions := decoder.Feed([]byte{0x1b}); len(actions) != 0 {
		t.Fatalf("first actions = %#v, want none while sequence is incomplete", actions)
	}
	if actions := decoder.Feed([]byte("[A")); len(actions) != 1 || actions[0].Input.Kind != tui.InputHistoryPrevious {
		t.Fatalf("second actions = %#v, want one history previous action", actions)
	}
}

func TestDecodeCtrlCIgnoresInterruptWithoutQuitAction(t *testing.T) {
	if actions := keyinput.Decode([]byte{0x03}); len(actions) != 0 {
		t.Fatalf("actions = %#v, want none", actions)
	}
}

func TestDecodeIgnoresInvalidUTF8(t *testing.T) {
	actions := keyinput.Decode([]byte{0xff, 'a'})

	assertInputText(t, actions, []string{"a"})
}

func assertInputText(t *testing.T, actions []keyinput.Action, want []string) {
	t.Helper()
	if len(actions) != len(want) {
		t.Fatalf("actions length = %d, want %d", len(actions), len(want))
	}
	for index, text := range want {
		action := actions[index]
		if action.Quit {
			t.Fatalf("actions[%d].Quit = true, want false", index)
		}
		if action.Input.Kind != tui.InputText {
			t.Fatalf("actions[%d].Input.Kind = %v, want InputText", index, action.Input.Kind)
		}
		if action.Input.Text != text {
			t.Fatalf("actions[%d].Input.Text = %q, want %q", index, action.Input.Text, text)
		}
	}
}
