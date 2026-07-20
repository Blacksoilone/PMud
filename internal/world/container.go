package world

import "fmt"

func (w *World) itemIsContainer(itemID ItemID) bool {
	item, ok := w.items[itemID]
	if !ok {
		return false
	}
	_, ok = item.tagParams("tag.container")
	return ok
}

func (w *World) containerCapacity(itemID ItemID) int {
	item, ok := w.items[itemID]
	if !ok {
		return 0
	}
	params, ok := item.tagParams("tag.container")
	if !ok {
		return 0
	}
	cap, _ := params["capacity"].(int)
	if cap <= 0 {
		return 1
	}
	return cap
}

func (w *World) containerItemCount(containerID string) int {
	return len(w.itemsInContainer(containerID))
}

func (w *World) RemainingCapacity(itemID ItemID) int {
	maxCap := w.containerCapacity(itemID)
	current := w.containerItemCount(ItemContainerID(itemID))
	if current >= maxCap {
		return 0
	}
	return maxCap - current
}

func (w *World) ContainerIsOpen(itemID ItemID) bool {
	return w.containerOpen[itemID]
}

func (w *World) OpenContainer(itemID ItemID) bool {
	if !w.itemIsContainer(itemID) {
		return false
	}
	if w.ContainerIsOpen(itemID) {
		return true
	}
	// 检查是否上锁
	if w.ItemIsLocked(itemID) {
		return false
	}
	w.containerOpen[itemID] = true
	return true
}

func (w *World) CloseContainer(itemID ItemID) bool {
	if !w.itemIsContainer(itemID) {
		return false
	}
	if !w.ContainerIsOpen(itemID) {
		return true
	}
	w.containerOpen[itemID] = false
	return true
}

// ContainerContents 返回容器内的物品列表，关闭时返回空
func (w *World) ContainerContents(itemID ItemID) []ItemID {
	if !w.ContainerIsOpen(itemID) {
		return nil
	}
	return w.itemsInContainer(ItemContainerID(itemID))
}

// canHoldContainer 检查 targetContainer 是否可以装另一个容器物品
func (w *World) canHoldContainer(targetContainerID ItemID) bool {
	return !w.itemIsCarryable(targetContainerID)
}

// PutItemInContainer 将玩家背包中的物品放入容器
func (w *World) PutItemInContainer(itemID ItemID, containerID ItemID, playerID PlayerID) error {
	if !w.itemIsContainer(containerID) {
		return fmt.Errorf("%s 不是容器", w.items[containerID].Name)
	}
	if !w.ContainerIsOpen(containerID) {
		return fmt.Errorf("%s 是关闭的", w.items[containerID].Name)
	}
	// 检查物品是否在玩家背包中
	playerContainer := PlayerContainerID(playerID)
	inPlayer := false
	for _, id := range w.itemsInContainer(playerContainer) {
		if id == itemID {
			inPlayer = true
			break
		}
	}
	if !inPlayer {
		return fmt.Errorf("你没有那个物品")
	}
	// 容量检查
	if w.RemainingCapacity(containerID) <= 0 {
		return fmt.Errorf("%s 装不下了", w.items[containerID].Name)
	}
	// 容器嵌套检查
	if w.itemIsContainer(itemID) && !w.canHoldContainer(containerID) {
		return fmt.Errorf("不能把容器放进便携容器中")
	}
	w.itemLocations[itemID] = ContainerItemLocation{ContainerID: ItemContainerID(containerID)}
	return nil
}

// GetItemFromContainer 从容器的内容中取出物品到玩家背包
func (w *World) GetItemFromContainer(containerID ItemID, itemID ItemID, playerID PlayerID) error {
	if !w.itemIsContainer(containerID) {
		return fmt.Errorf("%s 不是容器", w.items[containerID].Name)
	}
	if !w.ContainerIsOpen(containerID) {
		return fmt.Errorf("%s 是关闭的", w.items[containerID].Name)
	}
	// 检查物品是否在容器中
	containerLocation := ItemContainerID(containerID)
	inContainer := false
	for _, id := range w.itemsInContainer(containerLocation) {
		if id == itemID {
			inContainer = true
			break
		}
	}
	if !inContainer {
		return fmt.Errorf("%s 里没有那个物品", w.items[containerID].Name)
	}
	volumeOK, _ := w.CanAddItem(playerID, itemID)
	if !volumeOK {
		return fmt.Errorf("背包空间不足")
	}
	w.itemLocations[itemID] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
	return nil
}

// ItemIsLocked 检查一个物品是否被锁定（tag.lockable + key_item_id 已设置）
func (w *World) ItemIsLocked(itemID ItemID) bool {
	item, ok := w.items[itemID]
	if !ok {
		return false
	}
	params, ok := item.tagParams("tag.lockable")
	if !ok {
		return false
	}
	keyID, ok := params["key_item_id"].(string)
	return ok && keyID != ""
}
