package client

import (
	"PMud/internal/command"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

type State struct {
	catalog content.ClientCatalog
}

type CommandResolution struct {
	Command    string
	Send       bool
	LocalEvent protocol.Event
}

func NewState(catalog content.ClientCatalog) *State {
	return &State{catalog: catalog}
}

func (s *State) Observe(event protocol.Event) {}

func (s *State) ResolveCommand(command string) string {
	return s.ResolveCommandInput(command).Command
}

func (s *State) ResolveCommandInput(input string) CommandResolution {
	parsed := command.ParseClientInput(input)
	switch parsedCommand := parsed.(type) {
	case command.LookCommand:
		return CommandResolution{Command: "look", Send: true}
	case command.InventoryCommand:
		return CommandResolution{Command: "inventory", Send: true}
	case command.HelpCommand:
		return CommandResolution{LocalEvent: systemMessageEvent("system.help")}
	case command.EmptyCommand:
		return CommandResolution{LocalEvent: systemMessageEvent("system.empty_input")}
	case command.MoveCommand:
		return CommandResolution{Command: "go " + parsedCommand.Direction, Send: true}
	case command.ItemCommand:
		return CommandResolution{Command: string(parsedCommand.Verb) + " " + parsedCommand.Target, Send: true}
	case command.UnknownCommand:
		return CommandResolution{Command: parsedCommand.Input, Send: true}
	default:
		return CommandResolution{Command: input, Send: true}
	}
}

func systemMessageEvent(messageKey string) protocol.Event {
	return protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": messageKey,
		},
	}
}
