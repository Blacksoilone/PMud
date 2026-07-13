package client

import (
	"PMud/internal/content"
	"PMud/internal/protocol"
	"strings"
	"sync"
)

type State struct {
	catalog    content.ClientCatalog
	itemIDs    map[string]string
	itemIDsMux sync.RWMutex
}

func NewState(catalog content.ClientCatalog) *State {
	state := &State{
		catalog: catalog,
		itemIDs: make(map[string]string),
	}
	for itemID, nameKey := range catalog.ItemNames {
		name, ok := catalog.Text[nameKey]
		if !ok {
			continue
		}
		state.itemIDs[name] = string(itemID)
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
		nameKey, ok := s.catalog.ItemNames[content.ItemID(itemID)]
		if !ok {
			continue
		}
		name, ok := s.catalog.Text[nameKey]
		if !ok {
			continue
		}
		s.itemIDs[name] = itemID
	}
}

func (s *State) ResolveCommand(command string) string {
	trimmed := strings.TrimSpace(command)
	if remainder, ok := strings.CutPrefix(trimmed, "get "); ok {
		return "get " + s.resolveItem(strings.TrimSpace(remainder))
	}
	if remainder, ok := strings.CutPrefix(trimmed, "drop "); ok {
		return "drop " + s.resolveItem(strings.TrimSpace(remainder))
	}
	if remainder, ok := strings.CutPrefix(trimmed, "examine "); ok {
		return "examine " + s.resolveItem(strings.TrimSpace(remainder))
	}
	return command
}

func (s *State) resolveItem(item string) string {
	if strings.HasPrefix(item, "item.") {
		return item
	}

	s.itemIDsMux.RLock()
	defer s.itemIDsMux.RUnlock()
	itemID, ok := s.itemIDs[item]
	if !ok {
		return item
	}
	return itemID
}
