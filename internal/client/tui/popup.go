package tui

import (
	"fmt"
	"strings"
)

func OpenPopup(model Model, content PopupContent) Model {
	model.Popup = PopupState{Active: true, Content: content}
	return model
}

func ClosePopup(model Model) Model {
	model.Popup = PopupState{}
	return model
}

func ScrollPopup(model Model, delta int, viewportLines int) Model {
	if !model.Popup.Active {
		return model
	}
	model.Popup.ScrollOffset = clampPopupScrollOffset(
		model.Popup.ScrollOffset+delta,
		len(model.Popup.Content.Lines),
		viewportLines,
	)
	return model
}

func clampPopupScrollOffset(offset int, lineCount int, viewportLines int) int {
	if offset < 0 {
		return 0
	}
	maxOffset := max(0, lineCount-viewportLines)
	if offset > maxOffset {
		return maxOffset
	}
	return offset
}

func MoveCursorPopup(model Model, delta int) Model {
	if !model.Popup.Active {
		return model
	}
	c := model.Popup.Content.Cursor + delta
	if c < 0 {
		c = 0
	}
	if c >= len(model.Popup.Content.QuestIDs) {
		c = len(model.Popup.Content.QuestIDs) - 1
	}
	model.Popup.Content.Cursor = c
	// Keep cursor visible by adjusting scroll
	lineIndex := c // each quest is one line
	if lineIndex < model.Popup.ScrollOffset {
		model.Popup.ScrollOffset = lineIndex
	}
	maxVisible := 1 // will be adjusted by caller if needed
	if lineIndex >= model.Popup.ScrollOffset+maxVisible {
		model.Popup.ScrollOffset = lineIndex - maxVisible + 1
	}
	return model
}

func submitQuestSelection(model Model) (Model, Command) {
	if !model.Popup.Active || model.Popup.Content.Kind != PopupQuestList {
		return model, Command{}
	}
	cursor := model.Popup.Content.Cursor
	if cursor < 0 || cursor >= len(model.Popup.Content.QuestIDs) {
		return model, Command{}
	}
	questID := model.Popup.Content.QuestIDs[cursor]
	model = ClosePopup(model)
	return model, Command{Line: "quest " + questID, Submitted: true}
}

// QuestListPopupContent builds popup content from server quest_list event fields.
func QuestListPopupContent(fields map[string]string) PopupContent {
	items := fields["items"]
	tracked := fields["tracked"]

	questIDs := make([]string, 0)
	lines := make([]string, 0)

	// items format: id|name|stage|state,id2|name2|stage2|state2
	parts := strings.Split(items, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}
		segments := strings.SplitN(part, "|", 4)
		if len(segments) < 4 {
			continue
		}
		id := segments[0]
		name := segments[1]
		state := segments[3]

		status := ""
		mark := " "
		if id == tracked {
			mark = ">"
			status = " (跟踪中)"
		} else if state == "completed" {
			status = " (已完成)"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", mark, name, status))
		questIDs = append(questIDs, id)
	}

	return PopupContent{
		Kind:      PopupQuestList,
		Title:     fmt.Sprintf("任务列表 (%d)", len(questIDs)),
		Lines:     lines,
		QuestIDs:  questIDs,
		TrackedID: tracked,
	}
}


