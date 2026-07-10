package presentation

type Event interface {
	EventKind() string
}

type SystemMessageEvent struct {
	Message string
}

func (e SystemMessageEvent) EventKind() string {
	return "SystemMessageEvent"
}

type RoomObservationEvent struct {
	Name        string
	Description string
	Exits       []string
	Items       []string
}

func (e RoomObservationEvent) EventKind() string {
	return "RoomObservationEvent"
}

type InventoryEvent struct {
	Items []string
}

func (e InventoryEvent) EventKind() string {
	return "InventoryEvent"
}
