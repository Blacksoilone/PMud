package tui

import (
	"strings"
	"unicode/utf8"
)

type InputKind int

const (
	InputText InputKind = iota
	InputBackspace
	InputClear
	InputSubmit
	InputForceRedraw
)

type Input struct {
	Kind InputKind
	Text string
}

func ApplyInput(model Model, input Input) (Model, Command) {
	switch input.Kind {
	case InputText:
		model.Input += input.Text
	case InputBackspace:
		model.Input = removeLastRune(model.Input)
	case InputClear:
		model.Input = ""
	case InputSubmit:
		return submitInput(model)
	case InputForceRedraw:
		return model, Command{}
	}
	return model, Command{}
}

func removeLastRune(value string) string {
	if value == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(value)
	return value[:len(value)-size]
}

func submitInput(model Model) (Model, Command) {
	line := strings.TrimSpace(model.Input)
	model.Input = ""
	return model, Command{Line: line, Submitted: true}
}
