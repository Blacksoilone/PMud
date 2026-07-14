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
	canonical, ok := standardDirections[strings.TrimSpace(direction)]
	return canonical, ok
}

func ParseClientInput(input string) ClientCommand {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return EmptyCommand{}
	}
	if trimmed == "look" || trimmed == "l" {
		return LookCommand{}
	}
	if trimmed == "inventory" || trimmed == "i" {
		return InventoryCommand{}
	}
	if trimmed == "help" {
		return HelpCommand{}
	}
	if direction, ok := CanonicalDirection(trimmed); ok {
		return MoveCommand{Direction: direction}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "go "); ok {
		if direction, ok := CanonicalDirection(remainder); ok {
			return MoveCommand{Direction: direction}
		}
		return UnknownCommand{Input: input}
	}
	if item, ok := strings.CutPrefix(trimmed, "get "); ok {
		return ItemCommand{Verb: ItemVerbGet, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "take "); ok {
		return ItemCommand{Verb: ItemVerbGet, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "drop "); ok {
		return ItemCommand{Verb: ItemVerbDrop, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "examine "); ok {
		return ItemCommand{Verb: ItemVerbExamine, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "x "); ok {
		return ItemCommand{Verb: ItemVerbExamine, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "inspect "); ok {
		return ItemCommand{Verb: ItemVerbExamine, Target: strings.TrimSpace(item)}
	}
	return UnknownCommand{Input: input}
}

func ParseServerInput(input string) ServerCommand {
	trimmed := strings.TrimSpace(input)
	if trimmed == "look" {
		return LookCommand{}
	}
	if trimmed == "inventory" {
		return InventoryCommand{}
	}
	if trimmed == "help" {
		return HelpCommand{}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "go "); ok {
		if direction, ok := CanonicalDirection(remainder); ok {
			return MoveCommand{Direction: direction}
		}
		return UnknownCommand{Input: input}
	}
	if item, ok := strings.CutPrefix(trimmed, "get "); ok {
		return ItemCommand{Verb: ItemVerbGet, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "drop "); ok {
		return ItemCommand{Verb: ItemVerbDrop, Target: strings.TrimSpace(item)}
	}
	if item, ok := strings.CutPrefix(trimmed, "examine "); ok {
		return ItemCommand{Verb: ItemVerbExamine, Target: strings.TrimSpace(item)}
	}
	return UnknownCommand{Input: input}
}
