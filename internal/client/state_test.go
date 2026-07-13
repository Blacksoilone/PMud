package client

import (
	"PMud/internal/content"
	"PMud/internal/protocol"
	"testing"
)

func TestState_ResolveCommand_mapsObservedItemNameToID(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern",
		},
	})

	got := state.ResolveCommand("get 旧油灯")

	want := "get item.tutorial.old_lantern"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_mapsCatalogItemNameBeforeObservation(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	got := state.ResolveCommand("get 旧油灯")

	want := "get item.tutorial.old_lantern"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestState_ResolveCommand_mapsInventoryItemNameToID(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{
		Name: "inventory",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern",
		},
	})

	got := state.ResolveCommand("drop 旧油灯")

	want := "drop item.tutorial.old_lantern"
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
