package tui

import "PMud/internal/protocol"

type Model struct {
	Events       []protocol.Event
	Input        string
	HistoryLimit int
}

type Command struct {
	Line string
}

func NewModel(historyLimit int) Model {
	return Model{HistoryLimit: normalizeHistoryLimit(historyLimit)}
}

func ApplyEvent(model Model, event protocol.Event) Model {
	limit := normalizeHistoryLimit(model.HistoryLimit)
	events := append([]protocol.Event(nil), model.Events...)
	events = append(events, event)
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	model.Events = events
	model.HistoryLimit = limit
	return model
}

func normalizeHistoryLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	return limit
}
