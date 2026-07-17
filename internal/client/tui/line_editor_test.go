package tui

import "testing"

func TestLineEditorInsertsAtCursor(t *testing.T) {
	editor := NewLineEditor(100)
	editor = editor.Insert("get 灯")
	editor = editor.MoveLeft()
	editor = editor.Insert("旧油")

	if got := editor.String(); got != "get 旧油灯" {
		t.Fatalf("String() = %q, want get 旧油灯", got)
	}
}

func TestLineEditorDeletesBeforeAndAtCursor(t *testing.T) {
	editor := NewLineEditor(100).Insert("abc")
	editor = editor.MoveLeft().Backspace()
	if got := editor.String(); got != "ac" {
		t.Fatalf("after Backspace = %q, want ac", got)
	}
	editor = editor.Delete()
	if got := editor.String(); got != "a" {
		t.Fatalf("after Delete = %q, want a", got)
	}
}

func TestLineEditorHistoryPreservesDraft(t *testing.T) {
	editor := NewLineEditor(100)
	editor = editor.Submit("look")
	editor = editor.Submit("inventory")
	editor = editor.Insert("draft")
	editor = editor.HistoryPrevious()

	if got := editor.String(); got != "inventory" {
		t.Fatalf("previous history = %q, want inventory", got)
	}
	editor = editor.HistoryPrevious()
	if got := editor.String(); got != "look" {
		t.Fatalf("older history = %q, want look", got)
	}
	editor = editor.HistoryNext()
	if got := editor.String(); got != "inventory" {
		t.Fatalf("next history = %q, want inventory", got)
	}
	editor = editor.HistoryNext()
	if got := editor.String(); got != "draft" {
		t.Fatalf("draft restore = %q, want draft", got)
	}
}

func TestLineEditorLimitsHistory(t *testing.T) {
	editor := NewLineEditor(2)
	editor = editor.Submit("one").Submit("two").Submit("three")
	editor = editor.HistoryPrevious()

	if got := editor.String(); got != "three" {
		t.Fatalf("latest history = %q, want three", got)
	}
	editor = editor.HistoryPrevious()
	if got := editor.String(); got != "two" {
		t.Fatalf("oldest retained history = %q, want two", got)
	}
}
