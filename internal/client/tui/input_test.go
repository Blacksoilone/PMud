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

func TestApplyInputForceRedrawDoesNotChangeInputOrSubmitCommand(t *testing.T) {
	model := NewModel(3)
	model.Input = "look"

	model, command := ApplyInput(model, Input{Kind: InputForceRedraw})

	if command.Submitted {
		t.Fatal("Command.Submitted = true, want false")
	}
	if model.Input != "look" {
		t.Fatalf("Input = %q, want look", model.Input)
	}
}

func TestApplyInputQuitRequestsConfirmation(t *testing.T) {
	model := NewModel(3)
	model.Input = "quit"

	model, command := ApplyInput(model, Input{Kind: InputSubmit})

	if command.Submitted || command.ExitRequested {
		t.Fatalf("command = %#v, want confirmation without submission", command)
	}
	if !model.ExitConfirmation {
		t.Fatal("ExitConfirmation = false, want true")
	}
	if model.Input != "" {
		t.Fatalf("Input = %q, want empty", model.Input)
	}
}

func TestApplyInputExitConfirmationAcceptsYes(t *testing.T) {
	model := NewModel(3)
	model.ExitConfirmation = true
	model.Input = "yes"

	model, command := ApplyInput(model, Input{Kind: InputSubmit})

	if !command.ExitRequested {
		t.Fatalf("command = %#v, want exit request", command)
	}
	if model.ExitConfirmation {
		t.Fatal("ExitConfirmation = true, want false")
	}
}

func TestApplyInputExitConfirmationCancelsWithEscape(t *testing.T) {
	model := NewModel(3)
	model.ExitConfirmation = true
	model.Input = "maybe"

	model, command := ApplyInput(model, Input{Kind: InputCancel})

	if command.ExitRequested || model.ExitConfirmation || model.Input != "" {
		t.Fatalf("model = %#v command = %#v, want cancelled confirmation", model, command)
	}
}
