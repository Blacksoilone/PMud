package tui

import "testing"

func TestApplyInputAppendsText(t *testing.T) {
	model := NewModel(3)

	model, _ = ApplyInput(model, Input{Kind: InputText, Text: "get "})
	model, command := ApplyInput(model, Input{Kind: InputText, Text: "旧油灯"})

	if command.Line != "" {
		t.Fatalf("Command.Line = %q, want empty", command.Line)
	}
	if model.Input != "get 旧油灯" {
		t.Fatalf("Input = %q, want get 旧油灯", model.Input)
	}
}

func TestApplyInputBackspaceRemovesRune(t *testing.T) {
	model := NewModel(3)
	model.Input = "get 灯"

	model, command := ApplyInput(model, Input{Kind: InputBackspace})

	if command.Line != "" {
		t.Fatalf("Command.Line = %q, want empty", command.Line)
	}
	if model.Input != "get " {
		t.Fatalf("Input = %q, want get followed by space", model.Input)
	}
}

func TestApplyInputClearEmptiesBuffer(t *testing.T) {
	model := NewModel(3)
	model.Input = "inventory"

	model, command := ApplyInput(model, Input{Kind: InputClear})

	if command.Line != "" {
		t.Fatalf("Command.Line = %q, want empty", command.Line)
	}
	if model.Input != "" {
		t.Fatalf("Input = %q, want empty", model.Input)
	}
}

func TestApplyInputSubmitReturnsCommandAndClearsBuffer(t *testing.T) {
	model := NewModel(3)
	model.Input = "get 旧油灯"

	model, command := ApplyInput(model, Input{Kind: InputSubmit})

	if command.Line != "get 旧油灯" {
		t.Fatalf("Command.Line = %q, want get 旧油灯", command.Line)
	}
	if model.Input != "" {
		t.Fatalf("Input = %q, want empty", model.Input)
	}
}

func TestApplyInputSubmitWhitespaceReturnsEmptyCommand(t *testing.T) {
	model := NewModel(3)
	model.Input = "  \t "

	model, command := ApplyInput(model, Input{Kind: InputSubmit})

	if command.Line != "" {
		t.Fatalf("Command.Line = %q, want empty", command.Line)
	}
	if model.Input != "" {
		t.Fatalf("Input = %q, want empty", model.Input)
	}
}
