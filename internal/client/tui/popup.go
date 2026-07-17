package tui

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
