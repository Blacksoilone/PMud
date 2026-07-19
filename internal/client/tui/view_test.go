package tui

import (
	"strings"
	"testing"

	"PMud/internal/client/termwidth"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

func TestViewIncludesRoomEventAndPrompt(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(3)
	model.Input = "get 旧油灯"
	model = ApplyEvent(model, tutorialRoomEvent())

	got := View(model, catalog, 128).String()

	assertContains(t, got, "小地图")
	assertContains(t, got, "教学大厅")
	assertContains(t, got, "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。")
	assertContains(t, got, "> get 旧油灯")
}

func TestMinimapNeighborsUsesPlanarRoomProjection(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(3)
	model.Regions.Room.Neighbors = "north=room.tutorial.item_yard,up=room.tutorial.item_yard"

	neighbors := minimapNeighbors(model, catalog)

	if got := neighbors[MapNorth].Label; got != "物品庭院" {
		t.Fatalf("north label = %q, want 物品庭院", got)
	}
	if _, exists := neighbors[MapDirection("up")]; exists {
		t.Fatal("height-changing direction entered planar minimap")
	}
}

func TestTutorialMinimapShowsTriangleAsCurrentRoomToOneHopNeighborsOnly(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	model := NewModel(3)
	model.Regions.Room.Room = "room.tutorial.hall"
	model.Regions.Room.Neighbors = "north=room.tutorial.item_yard,east=room.tutorial.lock_hall"

	lines := renderMinimapGrid(MinimapRegion{
		Current:   MinimapRoom{Label: minimapLabel(model, compiled.Client)},
		Neighbors: minimapNeighbors(model, compiled.Client),
	})

	grid := termwidth.StripANSI(lines[0] + " " + lines[1] + " " + lines[2])
	if !strings.Contains(grid, "物品庭院") || !strings.Contains(grid, "锁钥厅") {
		t.Fatalf("minimap grid should contain both one-hop neighbors: %q", grid)
	}
	if strings.Contains(grid, "--") {
		t.Fatalf("minimap connected neighboring rooms to each other: %q", grid)
	}
}

func TestViewIncludesMultipleEventsInOrder(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(3)
	model = ApplyEvent(model, protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.help",
		},
	})
	model = ApplyEvent(model, protocol.Event{
		Name: "inventory",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern",
		},
	})

	got := View(model, catalog, 128).String()
	helpIndex := strings.Index(got, "可用命令")
	inventoryIndex := strings.Index(got, "你带着: 旧油灯")

	if helpIndex == -1 {
		t.Fatalf("view does not include help text:\n%s", got)
	}
	if inventoryIndex == -1 {
		t.Fatalf("view does not include inventory text:\n%s", got)
	}
	if helpIndex > inventoryIndex {
		t.Fatalf("events are out of order:\n%s", got)
	}
}

func TestViewKeepsCJKPromptVisible(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(1)
	model.Input = "drop 旧油灯"

	got := View(model, catalog, 128).String()

	assertContains(t, got, "> drop 旧油灯")
}

func TestViewRendersRightHUDPermanentPanes(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model.Input = "look"
	model = ApplyEvent(model, tutorialRoomEvent())
	model = ApplyEvent(model, protocol.Event{
		Name: "quest",
		Fields: map[string]string{
			"quest_id":   "quest.tutorial.first_steps",
			"quest_name": "初入练习场",
			"stage_id":   "quest.tutorial.first_steps.stage.enter_yard",
			"stage_text": "走进院子",
			"conditions": "查看练习木剑",
			"state":      "active",
		},
	})

	got := View(model, catalog, 128).String()

	for _, want := range []string{
		"房间 / 可见物",
		"日志",
		"小地图",
		"状态",
		"当前任务",
		"状态系统未开放",
		"初入练习场",
		"阶段: 走进院子",
		"目标: 查看练习木剑",
		"教学大厅",
		"> look",
	} {
		assertContains(t, got, want)
	}
}

func TestViewWithSizeUsesRequestedHeight(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model = ApplyEvent(model, tutorialRoomEvent())

	got := ViewWithSize(model, catalog, 128, 32)

	if len(got.Lines) != 32 {
		t.Fatalf("view height = %d, want 32", len(got.Lines))
	}
}

func TestViewWithSizeDoesNotExceedRequestedWidth(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model = ApplyEvent(model, tutorialRoomEvent())

	got := ViewWithSize(model, catalog, 166, 43)

	for index, line := range got.Lines {
		if width := termwidth.Width(line); width > 166 {
			t.Fatalf("line %d width = %d, want <= 166:\n%s", index, width, line)
		}
	}
}

func TestViewWithSizeRendersSharedFrameSeparators(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model = ApplyEvent(model, tutorialRoomEvent())

	got := ViewWithSize(model, catalog, 166, 43).String()

	assertContains(t, got, "房间 / 可见物")
	assertContains(t, got, "小地图")
	assertContains(t, got, "状态")
	assertContains(t, got, "当前任务")
	if strings.Count(got, "┌") != 1 || strings.Count(got, "┐") != 1 || strings.Count(got, "└") != 1 || strings.Count(got, "┘") != 1 {
		t.Fatalf("frame should have one outer border:\n%s", got)
	}
	if strings.Count(got, "├") < 3 || strings.Count(got, "┤") < 3 {
		t.Fatalf("frame should have shared internal separators:\n%s", got)
	}
}

func TestViewWithSizeInputSeparatorJoinsMainDivider(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)

	got := ViewWithSize(model, catalog, 166, 43)
	separator := got.Lines[len(got.Lines)-3]

	byteIndex := strings.Index(separator, "┴")
	divider := termwidth.Width(separator[:byteIndex])
	if divider != 130 {
		t.Fatalf("input junction column = %d, want 130: %q", divider, separator)
	}
}

func TestViewLogOmitsPermanentRoomAndQuestEvents(t *testing.T) {
	catalog := testClientCatalog(t)
	model := NewModel(5)
	model = ApplyEvent(model, tutorialRoomEvent())
	model = ApplyEvent(model, protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.item.taken",
			"item":        "item.tutorial.old_lantern",
		},
	})
	model = ApplyEvent(model, protocol.Event{
		Name: "quest",
		Fields: map[string]string{
			"quest_id":   "quest.tutorial.first_steps",
			"quest_name": "初入练习场",
			"stage_text": "走进院子",
			"conditions": "查看练习木剑",
		},
	})

	got := ViewWithSize(model, catalog, 128, 26).String()

	if strings.Count(got, "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。") != 1 {
		t.Fatalf("room event should only appear in room pane:\n%s", got)
	}
	if strings.Count(got, "初入练习场") != 1 {
		t.Fatalf("quest event should only appear in quest pane:\n%s", got)
	}
	assertContains(t, got, "你拿起了旧油灯")
}

func testClientCatalog(t *testing.T) content.ClientCatalog {
	t.Helper()
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatalf("Compile TutorialSource: %v", err)
	}
	return compiled.Client
}

func tutorialRoomEvent() protocol.Event {
	return protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            "room.tutorial.hall",
			"name_key":        "room.tutorial.hall.name",
			"description_key": "room.tutorial.hall.description",
			"exits":           "north,east,portal",
			"neighbors":       "north=room.tutorial.item_yard,east=room.tutorial.lock_hall",
		},
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("view does not include %q:\n%s", want, got)
	}
}
