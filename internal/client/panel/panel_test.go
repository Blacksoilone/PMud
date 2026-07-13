package panel

import (
	"PMud/internal/client/termwidth"
	"testing"
)

func TestBoxLines_drawsASCIIBox(t *testing.T) {
	got := BoxLines([]string{"HP: 10", "旧油灯"}, 8)
	want := []string{
		"+----------+",
		"| HP: 10   |",
		"| 旧油灯   |",
		"+----------+",
	}
	assertLines(t, got, want)
}

func TestBoxLines_returnsEqualCellWidthLines(t *testing.T) {
	got := BoxLines([]string{"HP: 10", "旧油灯"}, 8)
	wantWidth := termwidth.Width(got[0])
	for _, line := range got {
		if width := termwidth.Width(line); width != wantWidth {
			t.Fatalf("expected width %d for %q, got %d", wantWidth, line, width)
		}
	}
}

func TestBoxLines_usesLongestLineWhenWidthIsSmaller(t *testing.T) {
	got := BoxLines([]string{"旧油灯"}, 2)
	want := []string{
		"+--------+",
		"| 旧油灯 |",
		"+--------+",
	}
	assertLines(t, got, want)
}

func TestRenderLines_joinsLinesWithFinalNewline(t *testing.T) {
	got := RenderLines([]string{"+--+", "|  |", "+--+"})
	want := "+--+\n|  |\n+--+\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEqualWidths_reportsConsistentCellWidths(t *testing.T) {
	lines := []string{
		"+----------+",
		"| HP: 10   |",
		"| 旧油灯   |",
		"+----------+",
	}
	if !EqualWidths(lines) {
		t.Fatal("expected equal cell widths")
	}
}

func TestEqualWidths_reportsMismatchedCellWidths(t *testing.T) {
	lines := []string{"abc", "旧油灯"}
	if EqualWidths(lines) {
		t.Fatal("expected mismatched cell widths")
	}
}

func assertLines(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d: %v", len(want), len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("line %d: expected %q, got %q", index, want[index], got[index])
		}
	}
}
