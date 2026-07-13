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

func TestDecodeCtrlCQuit(t *testing.T) {
	actions := keyinput.Decode([]byte{0x03})

	if len(actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(actions))
	}
	if !actions[0].Quit {
		t.Fatalf("Quit = false, want true")
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
