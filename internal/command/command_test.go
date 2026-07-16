package command

import "testing"

func TestParseClientInput_mapsCommandAliases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ClientCommand
	}{
		{name: "take item", input: "take jiuyoudeng", want: ItemCommand{Verb: ItemVerbGet, Target: "jiuyoudeng"}},
		{name: "get item", input: "get jiuyoudeng", want: ItemCommand{Verb: ItemVerbGet, Target: "jiuyoudeng"}},
		{name: "drop item", input: "drop jiuyoudeng", want: ItemCommand{Verb: ItemVerbDrop, Target: "jiuyoudeng"}},
		{name: "x item", input: "x jiuyoudeng", want: ItemCommand{Verb: ItemVerbExamine, Target: "jiuyoudeng"}},
		{name: "inspect item", input: "inspect jiuyoudeng", want: ItemCommand{Verb: ItemVerbExamine, Target: "jiuyoudeng"}},
		{name: "examine item", input: "examine jiuyoudeng", want: ItemCommand{Verb: ItemVerbExamine, Target: "jiuyoudeng"}},
		{name: "inventory alias", input: "i", want: InventoryCommand{}},
		{name: "inventory", input: "inventory", want: InventoryCommand{}},
		{name: "look alias", input: "l", want: LookCommand{}},
		{name: "look", input: "look", want: LookCommand{}},
		{name: "quest", input: "quest", want: QuestCommand{}},
		{name: "help", input: "help", want: HelpCommand{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClientInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseClientInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseClientInput_mapsCommandAliasesCaseInsensitively(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ClientCommand
	}{
		{name: "take uppercase", input: "TAKE Old_Lantern", want: ItemCommand{Verb: ItemVerbGet, Target: "Old_Lantern"}},
		{name: "examine mixed case", input: "Examine Old_Lantern", want: ItemCommand{Verb: ItemVerbExamine, Target: "Old_Lantern"}},
		{name: "inventory uppercase", input: "I", want: InventoryCommand{}},
		{name: "look uppercase", input: "L", want: LookCommand{}},
		{name: "quest uppercase", input: "QUEST", want: QuestCommand{}},
		{name: "help uppercase", input: "HELP", want: HelpCommand{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClientInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseClientInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseClientInput_mapsStandardDirections(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  MoveCommand
	}{
		{name: "bare north", input: "north", want: MoveCommand{Direction: "north"}},
		{name: "bare n", input: "n", want: MoveCommand{Direction: "north"}},
		{name: "go n", input: "go n", want: MoveCommand{Direction: "north"}},
		{name: "bare south", input: "south", want: MoveCommand{Direction: "south"}},
		{name: "bare east", input: "east", want: MoveCommand{Direction: "east"}},
		{name: "bare west", input: "west", want: MoveCommand{Direction: "west"}},
		{name: "bare up", input: "up", want: MoveCommand{Direction: "up"}},
		{name: "bare down", input: "down", want: MoveCommand{Direction: "down"}},
		{name: "bare northeast", input: "northeast", want: MoveCommand{Direction: "northeast"}},
		{name: "bare ne", input: "ne", want: MoveCommand{Direction: "northeast"}},
		{name: "go nw", input: "go nw", want: MoveCommand{Direction: "northwest"}},
		{name: "bare southeast", input: "southeast", want: MoveCommand{Direction: "southeast"}},
		{name: "bare sw", input: "sw", want: MoveCommand{Direction: "southwest"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClientInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseClientInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseClientInput_mapsStandardDirectionsCaseInsensitively(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  MoveCommand
	}{
		{name: "bare uppercase", input: "N", want: MoveCommand{Direction: "north"}},
		{name: "go uppercase alias", input: "GO NW", want: MoveCommand{Direction: "northwest"}},
		{name: "full mixed case", input: "NorthEast", want: MoveCommand{Direction: "northeast"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClientInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseClientInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCanonicalDirection_mapsOnlyStandardDirections(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		want      string
		wantOK    bool
	}{
		{name: "north", direction: "n", want: "north", wantOK: true},
		{name: "north chinese", direction: "北", want: "north", wantOK: true},
		{name: "south", direction: "s", want: "south", wantOK: true},
		{name: "south chinese", direction: "南", want: "south", wantOK: true},
		{name: "east", direction: "e", want: "east", wantOK: true},
		{name: "west", direction: "w", want: "west", wantOK: true},
		{name: "up", direction: "u", want: "up", wantOK: true},
		{name: "down", direction: "d", want: "down", wantOK: true},
		{name: "northeast", direction: "ne", want: "northeast", wantOK: true},
		{name: "northwest", direction: "nw", want: "northwest", wantOK: true},
		{name: "southeast", direction: "se", want: "southeast", wantOK: true},
		{name: "southwest", direction: "sw", want: "southwest", wantOK: true},
		{name: "uppercase alias", direction: "NW", want: "northwest", wantOK: true},
		{name: "mixed full name", direction: "SouthEast", want: "southeast", wantOK: true},
		{name: "keeps special exits out", direction: "trapdoor", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CanonicalDirection(tt.direction)

			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("direction = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseClientInput_preservesUnknownInput(t *testing.T) {
	got := ParseClientInput("dance wildly")

	want := UnknownCommand{Input: "dance wildly"}
	if got != want {
		t.Fatalf("ParseClientInput() = %#v, want %#v", got, want)
	}
}

func TestParseClientInput_preservesCommandsMissingRequiredTargets(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "go missing direction", input: "go"},
		{name: "get missing item", input: "get"},
		{name: "take missing item", input: "take"},
		{name: "drop missing item", input: "drop"},
		{name: "examine missing item", input: "examine"},
		{name: "x missing item", input: "x"},
		{name: "inspect missing item", input: "inspect"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseClientInput(tt.input)

			want := UnknownCommand{Input: tt.input}
			if got != want {
				t.Fatalf("ParseClientInput(%q) = %#v, want %#v", tt.input, got, want)
			}
		})
	}
}

func TestParseClientInput_mapsEmptyInput(t *testing.T) {
	got := ParseClientInput("   ")

	want := EmptyCommand{}
	if got != want {
		t.Fatalf("ParseClientInput() = %#v, want %#v", got, want)
	}
}

func TestParseServerInput_mapsCanonicalCommands(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ServerCommand
	}{
		{name: "look", input: "look", want: LookCommand{}},
		{name: "help", input: "help", want: HelpCommand{}},
		{name: "inventory", input: "inventory", want: InventoryCommand{}},
		{name: "quest", input: "quest", want: QuestCommand{}},
		{name: "go direction", input: "go north", want: MoveCommand{Direction: "north"}},
		{name: "go direction alias", input: "go nw", want: MoveCommand{Direction: "northwest"}},
		{name: "get item id", input: "get item.tutorial.old_lantern", want: ItemCommand{Verb: ItemVerbGet, Target: "item.tutorial.old_lantern"}},
		{name: "drop item id", input: "drop item.tutorial.old_lantern", want: ItemCommand{Verb: ItemVerbDrop, Target: "item.tutorial.old_lantern"}},
		{name: "examine item id", input: "examine item.tutorial.old_lantern", want: ItemCommand{Verb: ItemVerbExamine, Target: "item.tutorial.old_lantern"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServerInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseServerInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseServerInput_mapsCanonicalCommandsCaseInsensitively(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ServerCommand
	}{
		{name: "look uppercase", input: "LOOK", want: LookCommand{}},
		{name: "inventory mixed case", input: "Inventory", want: InventoryCommand{}},
		{name: "quest uppercase", input: "QUEST", want: QuestCommand{}},
		{name: "go uppercase alias", input: "GO NW", want: MoveCommand{Direction: "northwest"}},
		{name: "get uppercase verb preserves target", input: "GET Item.Tutorial.Old_Lantern", want: ItemCommand{Verb: ItemVerbGet, Target: "Item.Tutorial.Old_Lantern"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServerInput(tt.input)

			if got != tt.want {
				t.Fatalf("ParseServerInput(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseServerInput_preservesUnknownAndClientOnlyAliases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "unknown", input: "dance wildly"},
		{name: "client take alias", input: "take item.tutorial.old_lantern"},
		{name: "client examine alias", input: "x item.tutorial.old_lantern"},
		{name: "client inventory alias", input: "i"},
		{name: "bare direction", input: "nw"},
		{name: "special exit", input: "go trapdoor"},
		{name: "go missing direction", input: "go"},
		{name: "get missing item", input: "get"},
		{name: "drop missing item", input: "drop"},
		{name: "examine missing item", input: "examine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServerInput(tt.input)

			want := UnknownCommand{Input: tt.input}
			if got != want {
				t.Fatalf("ParseServerInput(%q) = %#v, want %#v", tt.input, got, want)
			}
		})
	}
}
