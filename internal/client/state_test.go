package client

import (
	"PMud/internal/content"
	"testing"
)

func TestState_ResolveCommand_forwardsObservedItemPhrase(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get 旧油灯")

	want := "get 旧油灯"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_forwardsCatalogItemPhrase(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get 旧油灯")

	want := "get 旧油灯"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_forwardsInventoryItemPhrase(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("drop 旧油灯")

	want := "drop 旧油灯"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_forwardsExamineItemPhrase(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("examine 旧油灯")

	want := "examine 旧油灯"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_keepsDirectItemID(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get item.tutorial.old_lantern")

	want := "get item.tutorial.old_lantern"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_keepsUnknownItemName(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get 不存在的东西")

	want := "get 不存在的东西"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_forwardsAliasItemPhrase(t *testing.T) {
	compiled, err := content.Compile(aliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get jiuyoudeng")

	want := "get jiuyoudeng"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_keepsAmbiguousAliasName(t *testing.T) {
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get shared")

	want := "get shared"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_keepsAmbiguousDisplayName(t *testing.T) {
	source := content.TutorialSource()
	source.Text["item.tutorial.old_lantern.name"] = "shared"
	source.Text["item.tutorial.practice_sword.name"] = "shared"
	compiled, err := content.Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("examine shared")

	want := "examine shared"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommandInput_forwardsAmbiguousItemPhrase(t *testing.T) {
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommandInput("get shared")

	if !got.Send {
		t.Fatalf("expected ambiguous phrase to send")
	}
	if got.Command != "get shared" {
		t.Fatalf("command = %q, want original command", got.Command)
	}
}

func TestState_ResolveCommandInput_mapsCommandAliases(t *testing.T) {
	compiled, err := content.Compile(aliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{name: "take item", command: "take jiuyoudeng", want: "get jiuyoudeng"},
		{name: "x item", command: "x jiuyoudeng", want: "examine jiuyoudeng"},
		{name: "inspect item", command: "inspect jiuyoudeng", want: "examine jiuyoudeng"},
		{name: "inventory", command: "i", want: "inventory"},
		{name: "look", command: "l", want: "look"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.ResolveCommandInput(tt.command)

			if !got.Send {
				t.Fatalf("expected command to send")
			}
			if got.Command != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got.Command)
			}
		})
	}
}

func TestState_ResolveCommandInput_mapsStandardDirections(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{name: "bare north", command: "north", want: "go north"},
		{name: "bare n", command: "n", want: "go north"},
		{name: "go n", command: "go n", want: "go north"},
		{name: "bare south", command: "south", want: "go south"},
		{name: "bare east", command: "east", want: "go east"},
		{name: "bare west", command: "west", want: "go west"},
		{name: "bare up", command: "up", want: "go up"},
		{name: "bare down", command: "down", want: "go down"},
		{name: "bare northeast", command: "northeast", want: "go northeast"},
		{name: "bare ne", command: "ne", want: "go northeast"},
		{name: "go nw", command: "go nw", want: "go northwest"},
		{name: "bare southeast", command: "southeast", want: "go southeast"},
		{name: "bare sw", command: "sw", want: "go southwest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.ResolveCommandInput(tt.command)

			if !got.Send {
				t.Fatalf("expected command to send")
			}
			if got.Command != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got.Command)
			}
		})
	}
}

func aliasContentSource() content.ContentSource {
	source := content.TutorialSource()
	source.Items[0].Aliases = []content.TextKey{
		"item.tutorial.old_lantern.alias.jiuyoudeng",
	}
	source.Text["item.tutorial.old_lantern.alias.jiuyoudeng"] = "jiuyoudeng"
	return source
}

func ambiguousAliasContentSource() content.ContentSource {
	source := content.TutorialSource()
	source.Items[0].Aliases = []content.TextKey{
		"item.tutorial.old_lantern.alias.shared",
	}
	source.Items[1].Aliases = []content.TextKey{
		"item.tutorial.practice_sword.alias.shared",
	}
	source.Text["item.tutorial.old_lantern.alias.shared"] = "shared"
	source.Text["item.tutorial.practice_sword.alias.shared"] = "shared"
	return source
}
