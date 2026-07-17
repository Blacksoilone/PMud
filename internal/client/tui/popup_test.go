package tui

import (
	"strings"
	"testing"

	"PMud/internal/client/termwidth"
	"PMud/internal/protocol"
)

func TestOpenPopupReplacesActivePopup(t *testing.T) {
	// Given
	model := NewModel(3)
	model = OpenPopup(model, PopupContent{Title: "Help", Lines: []string{"look", "inventory"}})
	model = ScrollPopup(model, 1, 1)

	// When
	model = OpenPopup(model, PopupContent{Title: "Inventory", Lines: []string{"old lantern"}})

	// Then
	if !model.Popup.Active {
		t.Fatal("Popup.Active = false, want true")
	}
	if model.Popup.Content.Title != "Inventory" {
		t.Fatalf("Popup.Content.Title = %q, want Inventory", model.Popup.Content.Title)
	}
	if len(model.Popup.Content.Lines) != 1 || model.Popup.Content.Lines[0] != "old lantern" {
		t.Fatalf("Popup.Content.Lines = %#v, want inventory content", model.Popup.Content.Lines)
	}
	if model.Popup.ScrollOffset != 0 {
		t.Fatalf("Popup.ScrollOffset = %d, want replacement to reset scroll", model.Popup.ScrollOffset)
	}
}

func TestClosePopupPreservesBaseModel(t *testing.T) {
	// Given
	model := NewModel(3)
	model.Input = "look"
	model = ApplyEvent(model, protocol.Event{Name: "room", Fields: map[string]string{"room": "room.tutorial.start"}})
	model = OpenPopup(model, PopupContent{Title: "Help", Lines: []string{"look"}})

	// When
	model = ClosePopup(model)

	// Then
	if model.Popup.Active {
		t.Fatal("Popup.Active = true, want false")
	}
	if model.Input != "look" {
		t.Fatalf("Input = %q, want look", model.Input)
	}
	if model.Regions.Room.Room != "room.tutorial.start" {
		t.Fatalf("Regions.Room.Room = %q, want room.tutorial.start", model.Regions.Room.Room)
	}
	if len(model.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(model.Events))
	}
}

func TestScrollPopupClampsToLowerBound(t *testing.T) {
	// Given
	model := NewModel(3)
	model = OpenPopup(model, PopupContent{Title: "Help", Lines: []string{"one", "two", "three"}})
	model = ScrollPopup(model, 2, 1)

	// When
	model = ScrollPopup(model, -5, 1)

	// Then
	if model.Popup.ScrollOffset != 0 {
		t.Fatalf("Popup.ScrollOffset = %d, want 0", model.Popup.ScrollOffset)
	}
}

func TestScrollPopupClampsToUpperBound(t *testing.T) {
	// Given
	model := NewModel(3)
	model = OpenPopup(model, PopupContent{Title: "Help", Lines: []string{"one", "two", "three", "four"}})

	// When
	model = ScrollPopup(model, 10, 2)

	// Then
	if model.Popup.ScrollOffset != 2 {
		t.Fatalf("Popup.ScrollOffset = %d, want 2", model.Popup.ScrollOffset)
	}
}

func TestScrollPopupEmptyContentStaysAtZero(t *testing.T) {
	// Given
	model := NewModel(3)
	model = OpenPopup(model, PopupContent{Title: "Empty"})

	// When
	model = ScrollPopup(model, 3, 5)

	// Then
	if !model.Popup.Active {
		t.Fatal("Popup.Active = false, want true")
	}
	if model.Popup.ScrollOffset != 0 {
		t.Fatalf("Popup.ScrollOffset = %d, want 0", model.Popup.ScrollOffset)
	}
}

func TestViewWithSizeRendersPopupOverContentAndKeepsInputVisible(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model.Input = "look"
	model = OpenPopup(model, PopupContent{
		Kind:  PopupHelp,
		Title: "帮助",
		Lines: []string{"基础命令", "look - 查看周围"},
	})

	got := ViewWithSize(model, catalog, 128, 26)

	if !strings.Contains(got.String(), "╔") || !strings.Contains(got.String(), "帮助") {
		t.Fatalf("view does not contain popup frame:\n%s", got.String())
	}
	if !strings.Contains(got.String(), "\x1b[2m") {
		t.Fatalf("view does not dim popup background:\n%s", got.String())
	}
	if !strings.HasPrefix(got.Lines[0], "\x1b[2m") {
		t.Fatalf("top border is not dimmed while popup is active: %q", got.Lines[0])
	}
	if !strings.Contains(got.Lines[len(got.Lines)-2], "> look") {
		t.Fatalf("popup covered input row: %q", got.Lines[len(got.Lines)-2])
	}
	for index, line := range got.Lines {
		if width := termwidth.Width(line); width != 128 {
			t.Fatalf("line %d width = %d, want 128: %q", index, width, line)
		}
	}
}

func TestViewWithSizePopupScrollsBody(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	lines := make([]string, 30)
	for index := range lines {
		lines[index] = string(rune('A' + index))
	}
	model = OpenPopup(model, PopupContent{Kind: PopupHelp, Title: "帮助", Lines: lines})
	model.Popup.ScrollOffset = 2

	got := ViewWithSize(model, catalog, 128, 26).String()

	if strings.Contains(got, "│ A") || strings.Contains(got, "│ B") {
		t.Fatalf("popup still shows skipped content after scrolling:\n%s", got)
	}
	if !strings.Contains(got, "[Esc] 关闭") || !strings.Contains(got, "[↑↓/滚轮] 滚动") {
		t.Fatalf("popup footer missing controls:\n%s", got)
	}
}
