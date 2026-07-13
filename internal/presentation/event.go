package presentation

type Event interface {
	EventKind() string
}

type SystemMessageEvent struct {
	MessageKey string
	Fields     map[string]string
	Message    string
}

func (e SystemMessageEvent) EventKind() string {
	return "SystemMessageEvent"
}

type RoomObservationEvent struct {
	Room           string
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Exits          []string
	Items          []string
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

type ItemObservationEvent struct {
	Item           string
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
}

func (e ItemObservationEvent) EventKind() string {
	return "ItemObservationEvent"
}
