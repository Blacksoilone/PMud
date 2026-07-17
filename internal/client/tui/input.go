package tui

import (
	"strings"
)

type InputKind int

const (
	InputText InputKind = iota
	InputBackspace
	InputClear
	InputSubmit
	InputForceRedraw
	InputCancel
	InputMoveLeft
	InputMoveRight
	InputDelete
	InputHome
	InputEnd
	InputHistoryPrevious
	InputHistoryNext
)

type Input struct {
	Kind InputKind
	Text string
}

func ApplyInput(model Model, input Input) (Model, Command) {
	model = syncEditorFromInput(model)
	switch input.Kind {
	case InputText:
		model.Editor = model.Editor.Insert(input.Text)
	case InputBackspace:
		model.Editor = model.Editor.Backspace()
	case InputClear:
		model.Editor = model.Editor.Clear()
	case InputSubmit:
		return submitInput(model)
	case InputForceRedraw:
		return model, Command{}
	case InputCancel:
		model.Editor = model.Editor.Clear()
		model.ExitConfirmation = false
	case InputMoveLeft:
		model.Editor = model.Editor.MoveLeft()
	case InputMoveRight:
		model.Editor = model.Editor.MoveRight()
	case InputDelete:
		model.Editor = model.Editor.Delete()
	case InputHome:
		model.Editor = model.Editor.Home()
	case InputEnd:
		model.Editor = model.Editor.End()
	case InputHistoryPrevious:
		model.Editor = model.Editor.HistoryPrevious()
	case InputHistoryNext:
		model.Editor = model.Editor.HistoryNext()
	}
	model.Input = model.Editor.String()
	return model, Command{}
}

func submitInput(model Model) (Model, Command) {
	line := strings.TrimSpace(model.Editor.String())
	model.Editor = model.Editor.Clear()
	model.Input = ""
	if model.ExitConfirmation {
		model.ExitConfirmation = false
		if strings.EqualFold(line, "y") || strings.EqualFold(line, "yes") {
			return model, Command{ExitRequested: true}
		}
		return model, Command{}
	}
	if strings.EqualFold(line, "quit") || strings.EqualFold(line, "exit") {
		model.ExitConfirmation = true
		return model, Command{}
	}
	model.Editor = model.Editor.Submit(line)
	return model, Command{Line: line, Submitted: true}
}

func syncEditorFromInput(model Model) Model {
	if model.Editor.historyLimit == 0 {
		model.Editor = NewLineEditor(defaultHistoryLimit)
	}
	if model.Editor.String() != model.Input {
		model.Editor = model.Editor.Replace(model.Input)
	}
	return model
}
