package termwidth

import "testing"

func TestWidth_countsASCIIAsOneCell(t *testing.T) {
	got := Width("hello")
	want := 5
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestWidth_countsCJKAsTwoCells(t *testing.T) {
	got := Width("旧油灯")
	want := 6
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestWidth_countsMixedText(t *testing.T) {
	got := Width("HP:旧油灯")
	want := 9
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestLineWidth_returnsLongestRenderedLine(t *testing.T) {
	got := LineWidth("abc\n旧油灯")
	want := 6
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}
