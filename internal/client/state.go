package client

import (
	"PMud/internal/command"
	"PMud/internal/content"
	"PMud/internal/protocol"
	"slices"
	"strings"
	"sync"
)

type State struct {
	catalog    content.ClientCatalog
	itemIDs    map[string][]string
	itemIDsMux sync.RWMutex
}

type CommandResolution struct {
	Command        string
	Send           bool
	AmbiguousItems []string
	LocalEvent     protocol.Event
}

func NewState(catalog content.ClientCatalog) *State {
	state := &State{
		catalog: catalog,
		itemIDs: make(map[string][]string),
	}
	for itemID, nameKey := range catalog.ItemDisplayNames {
		name, ok := catalog.Text[nameKey]
		if !ok {
			continue
		}
		state.addItemMatch(name, string(itemID))
		for _, aliasKey := range catalog.ItemAliases[itemID] {
			alias, ok := catalog.Text[aliasKey]
			if !ok {
				continue
			}
			state.addItemMatch(alias, string(itemID))
		}
	}
	return state
}

func (s *State) Observe(event protocol.Event) {
	items := event.Fields["items"]
	if items == "" {
		return
	}

	s.itemIDsMux.Lock()
	defer s.itemIDsMux.Unlock()
	for itemID := range strings.SplitSeq(items, ",") {
		nameKey, ok := s.catalog.ItemDisplayNames[content.ItemID(itemID)]
		if !ok {
			continue
		}
		name, ok := s.catalog.Text[nameKey]
		if !ok {
			continue
		}
		s.addItemMatch(name, itemID)
	}
}

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
		return s.resolveItemCommand(input, string(parsedCommand.Verb), parsedCommand.Target)
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

func (s *State) AmbiguousItemEvent(itemIDs []string) protocol.Event {
	names := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		nameKey, ok := s.catalog.ItemDisplayNames[content.ItemID(itemID)]
		if !ok {
			names = append(names, itemID)
			continue
		}
		name, ok := s.catalog.Text[nameKey]
		if !ok {
			names = append(names, itemID)
			continue
		}
		names = append(names, name)
	}
	return protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message": "名字不明确: " + strings.Join(names, ", "),
		},
	}
}

func (s *State) resolveItemCommand(command string, verb string, item string) CommandResolution {
	resolved := s.resolveItem(item)
	if len(resolved.AmbiguousItems) > 0 {
		resolved.Command = command
		resolved.Send = false
		return resolved
	}
	resolved.Command = verb + " " + resolved.Command
	resolved.Send = true
	return resolved
}

func (s *State) resolveItem(item string) CommandResolution {
	if strings.HasPrefix(item, "item.") {
		return CommandResolution{Command: item}
	}

	s.itemIDsMux.RLock()
	defer s.itemIDsMux.RUnlock()
	matches := s.itemIDs[item]
	if len(matches) == 0 {
		return CommandResolution{Command: item}
	}
	if len(matches) > 1 {
		candidates := append([]string(nil), matches...)
		slices.Sort(candidates)
		return CommandResolution{Command: item, AmbiguousItems: candidates}
	}
	return CommandResolution{Command: matches[0]}
}

func (s *State) addItemMatch(name string, itemID string) {
	if slices.Contains(s.itemIDs[name], itemID) {
		return
	}
	s.itemIDs[name] = append(s.itemIDs[name], itemID)
}
