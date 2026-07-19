package command

import "strings"

type ClientCommand interface {
	clientCommand()
}

type ServerCommand interface {
	serverCommand()
}

type ItemVerb string

const (
	ItemVerbGet     ItemVerb = "get"
	ItemVerbDrop    ItemVerb = "drop"
	ItemVerbExamine ItemVerb = "examine"
	ItemVerbLook    ItemVerb = "look"
)

type ItemCommand struct {
	Verb   ItemVerb
	Target string
}

func (ItemCommand) clientCommand() {}

func (ItemCommand) serverCommand() {}

type MoveCommand struct {
	Direction string
}

func (MoveCommand) clientCommand() {}

func (MoveCommand) serverCommand() {}

type LookCommand struct{}

func (LookCommand) clientCommand() {}

func (LookCommand) serverCommand() {}

type InventoryCommand struct{}

func (InventoryCommand) clientCommand() {}

func (InventoryCommand) serverCommand() {}

type QuestCommand struct{}

func (QuestCommand) clientCommand() {}

func (QuestCommand) serverCommand() {}

type VerbCommand struct{}

func (VerbCommand) clientCommand() {}

func (VerbCommand) serverCommand() {}

type HelpCommand struct{}

func (HelpCommand) clientCommand() {}

func (HelpCommand) serverCommand() {}

type EmptyCommand struct{}

func (EmptyCommand) clientCommand() {}

type UnknownCommand struct {
	Input string
}

func (UnknownCommand) clientCommand() {}

func (UnknownCommand) serverCommand() {}

var standardDirections = map[string]string{
	"north":     "north",
	"n":         "north",
	"北":         "north",
	"south":     "south",
	"s":         "south",
	"南":         "south",
	"east":      "east",
	"e":         "east",
	"west":      "west",
	"w":         "west",
	"up":        "up",
	"u":         "up",
	"down":      "down",
	"d":         "down",
	"northeast": "northeast",
	"ne":        "northeast",
	"northwest": "northwest",
	"nw":        "northwest",
	"southeast": "southeast",
	"se":        "southeast",
	"southwest": "southwest",
	"sw":        "southwest",
}

func CanonicalDirection(direction string) (string, bool) {
	canonical, ok := standardDirections[strings.ToLower(strings.TrimSpace(direction))]
	return canonical, ok
}

func ParseClientInput(input string) ClientCommand {
	trimmed := strings.TrimSpace(input)
	verb, target := splitVerbTarget(trimmed)
	if trimmed == "" {
		return EmptyCommand{}
	}
	if verb == "look" || verb == "l" {
		if target != "" {
			return clientItemCommand(input, target, ItemVerbLook)
		}
		return LookCommand{}
	}
	if verb == "inventory" || verb == "i" {
		return InventoryCommand{}
	}
	if verb == "quest" {
		return QuestCommand{}
	}
	if verb == "verb" || verb == "verbs" {
		return VerbCommand{}
	}
	if verb == "help" {
		return HelpCommand{}
	}
	if direction, ok := CanonicalDirection(trimmed); ok {
		return MoveCommand{Direction: direction}
	}
	if verb == "go" {
		if direction, ok := CanonicalDirection(target); ok {
			return MoveCommand{Direction: direction}
		}
		return UnknownCommand{Input: input}
	}
	if verb == "get" {
		return clientItemCommand(input, target, ItemVerbGet)
	}
	if verb == "take" {
		return clientItemCommand(input, target, ItemVerbGet)
	}
	if verb == "drop" {
		return clientItemCommand(input, target, ItemVerbDrop)
	}
	if verb == "examine" {
		return clientItemCommand(input, target, ItemVerbExamine)
	}
	if verb == "x" {
		return clientItemCommand(input, target, ItemVerbExamine)
	}
	if verb == "inspect" {
		return clientItemCommand(input, target, ItemVerbExamine)
	}
	return UnknownCommand{Input: input}
}

func ParseServerInput(input string) ServerCommand {
	trimmed := strings.TrimSpace(input)
	verb, target := splitVerbTarget(trimmed)
	if verb == "look" {
		if target != "" {
			return serverItemCommand(input, target, ItemVerbLook)
		}
		return LookCommand{}
	}
	if verb == "inventory" {
		return InventoryCommand{}
	}
	if verb == "quest" {
		return QuestCommand{}
	}
	if verb == "verb" || verb == "verbs" {
		return VerbCommand{}
	}
	if verb == "help" {
		return HelpCommand{}
	}
	if verb == "go" {
		if direction, ok := CanonicalDirection(target); ok {
			return MoveCommand{Direction: direction}
		}
		return UnknownCommand{Input: input}
	}
	if verb == "get" {
		return serverItemCommand(input, target, ItemVerbGet)
	}
	if verb == "drop" {
		return serverItemCommand(input, target, ItemVerbDrop)
	}
	if verb == "examine" {
		return serverItemCommand(input, target, ItemVerbExamine)
	}
	return UnknownCommand{Input: input}
}

func clientItemCommand(input string, target string, verb ItemVerb) ClientCommand {
	if target == "" {
		return UnknownCommand{Input: input}
	}
	return ItemCommand{Verb: verb, Target: target}
}

func serverItemCommand(input string, target string, verb ItemVerb) ServerCommand {
	if target == "" {
		return UnknownCommand{Input: input}
	}
	return ItemCommand{Verb: verb, Target: target}
}

func splitVerbTarget(input string) (string, string) {
	verb, target, ok := strings.Cut(input, " ")
	if !ok {
		return strings.ToLower(input), ""
	}
	return strings.ToLower(verb), strings.TrimSpace(target)
}
