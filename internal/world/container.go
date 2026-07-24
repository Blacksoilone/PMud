package world

import "fmt"

func (w *World) itemIsContainer(itemID EntityID) bool {
	return w.store.Tag(itemID, "tag.container")
}

func (w *World) containerCapacity(itemID EntityID) int {
	ent := w.store.Get(itemID)
	if ent == nil {
		return 0
	}
	for _, inst := range ent.Tags {
		if inst.DefinitionID == "tag.container" {
			cap, _ := inst.Params["capacity"].(int)
			if cap <= 0 {
				return 1
			}
			return cap
		}
	}
	return 0
}

func (w *World) containerItemCount(containerID string) int {
	return len(w.containerContents[containerID])
}

func (w *World) RemainingCapacity(itemID EntityID) int {
	maxCap := w.containerCapacity(itemID)
	current := w.containerItemCount(ItemContainerID(itemID))
	if current >= maxCap {
		return 0
	}
	return maxCap - current
}

func (w *World) ContainerIsOpen(itemID EntityID) bool {
	return w.containerOpen[itemID]
}

func (w *World) OpenContainer(itemID EntityID) bool {
	if !w.itemIsContainer(itemID) {
		return false
	}
	if w.ContainerIsOpen(itemID) {
		return true
	}
	if w.ItemIsLocked(itemID) {
		return false
	}
	w.containerOpen[itemID] = true
	return true
}

func (w *World) CloseContainer(itemID EntityID) bool {
	if !w.itemIsContainer(itemID) {
		return false
	}
	if !w.ContainerIsOpen(itemID) {
		return true
	}
	w.containerOpen[itemID] = false
	return true
}

func (w *World) ContainerContents(itemID EntityID) []EntityID {
	if !w.ContainerIsOpen(itemID) {
		return nil
	}
	return w.containerContents[ItemContainerID(itemID)]
}

func (w *World) canHoldContainer(targetContainerID EntityID) bool {
	return !w.itemIsCarryable(targetContainerID)
}

func (w *World) PutItemInContainer(itemID, containerID, playerID EntityID) error {
	if !w.itemIsContainer(containerID) {
		return fmt.Errorf("%s 不是容器", w.store.Get(containerID).Name)
	}
	if !w.ContainerIsOpen(containerID) {
		return fmt.Errorf("%s 是关闭的", w.store.Get(containerID).Name)
	}
	playerContainer := PlayerContainerID(playerID)
	inPlayer := false
	for _, id := range w.containerContents[playerContainer] {
		if id == itemID {
			inPlayer = true
			break
		}
	}
	if !inPlayer {
		return fmt.Errorf("你没有那个物品")
	}
	if w.RemainingCapacity(containerID) <= 0 {
		return fmt.Errorf("%s 装不下了", w.store.Get(containerID).Name)
	}
	if w.itemIsContainer(itemID) && !w.canHoldContainer(containerID) {
		return fmt.Errorf("不能把容器放进便携容器中")
	}
	w.removeFromContainer(itemID)
	containerLoc := ItemContainerID(containerID)
	w.containerContents[containerLoc] = append(w.containerContents[containerLoc], itemID)
	return nil
}

func (w *World) GetItemFromContainer(containerID, itemID, playerID EntityID) error {
	if !w.itemIsContainer(containerID) {
		return fmt.Errorf("%s 不是容器", w.store.Get(containerID).Name)
	}
	if !w.ContainerIsOpen(containerID) {
		return fmt.Errorf("%s 是关闭的", w.store.Get(containerID).Name)
	}
	containerLoc := ItemContainerID(containerID)
	inContainer := false
	for _, id := range w.containerContents[containerLoc] {
		if id == itemID {
			inContainer = true
			break
		}
	}
	if !inContainer {
		return fmt.Errorf("%s 里没有那个物品", w.store.Get(containerID).Name)
	}
	volumeOK, _ := w.CanAddItem(playerID, itemID)
	if !volumeOK {
		return fmt.Errorf("背包空间不足")
	}
	w.removeFromContainer(itemID)
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], itemID)
	return nil
}

func (w *World) ItemIsLocked(itemID EntityID) bool {
	ent := w.store.Get(itemID)
	if ent == nil {
		return false
	}
	for _, inst := range ent.Tags {
		if inst.DefinitionID == "tag.lockable" {
			keyID, _ := inst.Params["key_item_id"].(string)
			return keyID != ""
		}
	}
	return false
}
