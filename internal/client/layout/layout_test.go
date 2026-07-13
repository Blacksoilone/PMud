package layout

import "testing"

func TestBlockString_rendersLinesWithFinalNewline(t *testing.T) {
	block := NewBlock([]string{"旧油灯", "HP: 10"})

	got := block.String()
	want := "旧油灯\nHP: 10\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBox_wrapsBlockInPanel(t *testing.T) {
	block := NewBlock([]string{"旧油灯"})

	got := Box(block, 6).String()
	want := "+--------+\n| 旧油灯 |\n+--------+\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestVBox_stacksBlocksVertically(t *testing.T) {
	top := NewBlock([]string{"旧油灯"})
	bottom := NewBlock([]string{"HP: 10"})

	got := VBox(top, bottom).String()
	want := "旧油灯\nHP: 10\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestHBox_joinsBlocksHorizontally(t *testing.T) {
	left := Box(NewBlock([]string{"旧油灯"}), 6)
	right := Box(NewBlock([]string{"HP"}), 4)

	got := HBox(2, left, right).String()
	want := "+--------+  +------+\n| 旧油灯 |  | HP   |\n+--------+  +------+\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
