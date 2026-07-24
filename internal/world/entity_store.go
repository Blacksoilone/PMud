package world

// EntityStore 是世界中所有实体的统一存储。
// 它替代了 World 当前的三张分散 map（rooms、items、players）。
//
// 核心能力：
//   - 统一的 Add/Get/Has
//   - 跨类型的 tag 索引（TagIndex）
//   - 位置索引（atRoom、container）
//   - 便捷的类型数据访问器（Room、Item、Player、Exit）
type EntityStore struct {
	entities map[EntityID]*Entity
	tagIndex map[TagID]map[EntityID]struct{}
	atRoom   map[EntityID][]EntityID // roomID → 在该房间中的 entityID
}

// NewEntityStore 创建并返回一个空的 EntityStore。
func NewEntityStore() *EntityStore {
	return &EntityStore{
		entities: make(map[EntityID]*Entity),
		tagIndex: make(map[TagID]map[EntityID]struct{}),
		atRoom:   make(map[EntityID][]EntityID),
	}
}

// ---------------------------------------------------------------------------
// 基本 CRUD
// ---------------------------------------------------------------------------

// Add 将实体加入存储，并构建 tag 索引。
// 如果已存在同名实体则覆盖。
func (s *EntityStore) Add(entity *Entity) {
	s.entities[entity.ID] = entity
	s.indexTags(entity)
}

func (s *EntityStore) Remove(id EntityID) {
	entity := s.entities[id]
	if entity == nil {
		return
	}
	for _, t := range entity.Tags {
		if set, ok := s.tagIndex[t.DefinitionID]; ok {
			delete(set, id)
			if len(set) == 0 {
				delete(s.tagIndex, t.DefinitionID)
			}
		}
	}
	s.RemoveFromRoom(id)
	delete(s.entities, id)
}

// Get 返回指定 ID 的实体指针；不存在时返回 nil。
func (s *EntityStore) Get(id EntityID) *Entity {
	return s.entities[id]
}

// Has 返回指定 ID 的实体是否存在。
func (s *EntityStore) Has(id EntityID) bool {
	_, ok := s.entities[id]
	return ok
}

// ---------------------------------------------------------------------------
// Tag 索引
// ---------------------------------------------------------------------------

// indexTags 将实体的所有 tag 加入反向索引。
func (s *EntityStore) indexTags(entity *Entity) {
	for _, t := range entity.Tags {
		set, ok := s.tagIndex[t.DefinitionID]
		if !ok {
			set = make(map[EntityID]struct{})
			s.tagIndex[t.DefinitionID] = set
		}
		set[entity.ID] = struct{}{}
	}
}

// FindByTag 返回所有持有指定 tag 的实体切片。
func (s *EntityStore) FindByTag(tagID TagID) []*Entity {
	set, ok := s.tagIndex[tagID]
	if !ok || len(set) == 0 {
		return nil
	}
	result := make([]*Entity, 0, len(set))
	for id := range set {
		if e := s.entities[id]; e != nil {
			result = append(result, e)
		}
	}
	return result
}

// FindByTagInRoom 返回指定房间内持有指定 tag 的所有实体。
// 同时检查房间自身的 tag。
func (s *EntityStore) FindByTagInRoom(tagID TagID, roomID EntityID) []*Entity {
	tagSet, ok := s.tagIndex[tagID]
	if !ok || len(tagSet) == 0 {
		return nil
	}

	seen := make(map[EntityID]struct{})
	var result []*Entity

	// 检查房间自身的 tag
	if _, tagged := tagSet[roomID]; tagged {
		if e := s.entities[roomID]; e != nil {
			result = append(result, e)
			seen[roomID] = struct{}{}
		}
	}

	// 检查房间内实体的 tag
	for _, eid := range s.atRoom[roomID] {
		if _, tagged := tagSet[eid]; tagged {
			if _, already := seen[eid]; already {
				continue
			}
			if e := s.entities[eid]; e != nil {
				result = append(result, e)
				seen[eid] = struct{}{}
			}
		}
	}

	return result
}

// ---------------------------------------------------------------------------
// 位置
// ---------------------------------------------------------------------------

// PlaceInRoom 将 entityID 放置在 roomID 中。
// 如果该 entity 此前已在另一房间，先移除旧记录。
func (s *EntityStore) PlaceInRoom(entityID, roomID EntityID) {
	s.removeFromAnyRoom(entityID)
	s.atRoom[roomID] = append(s.atRoom[roomID], entityID)
}

// EntitiesInRoom 返回指定房间中的所有实体 ID。
// 不包含房间自身。
func (s *EntityStore) EntitiesInRoom(roomID EntityID) []EntityID {
	ids := s.atRoom[roomID]
	if len(ids) == 0 {
		return nil
	}
	cp := make([]EntityID, len(ids))
	copy(cp, ids)
	return cp
}

// RemoveFromRoom 将 entityID 从房间中移除（公开版本）。
func (s *EntityStore) RemoveFromRoom(entityID EntityID) {
	s.removeFromAnyRoom(entityID)
}

// IsInRoom 返回 entityID 所在的房间 ID；如果不在任何房间中则返回空字符串。
func (s *EntityStore) IsInRoom(entityID EntityID) EntityID {
	for roomID, entities := range s.atRoom {
		for _, eid := range entities {
			if eid == entityID {
				return roomID
			}
		}
	}
	return ""
}

// removeFromAnyRoom 将 entityID 从当前所在的房间中移除。
func (s *EntityStore) removeFromAnyRoom(entityID EntityID) {
	for roomID, entities := range s.atRoom {
		for i, eid := range entities {
			if eid == entityID {
				s.atRoom[roomID] = append(entities[:i], entities[i+1:]...)
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 便捷类型数据访问器
// ---------------------------------------------------------------------------

// Room 返回实体的 RoomData；如果实体不是房间类型则返回 nil。
func (s *EntityStore) Room(id EntityID) *RoomData {
	if e := s.entities[id]; e != nil {
		return e.Room
	}
	return nil
}

// Item 返回实体的 ItemData；如果实体不是物品类型则返回 nil。
func (s *EntityStore) Item(id EntityID) *ItemData {
	if e := s.entities[id]; e != nil {
		return e.Item
	}
	return nil
}

// Player 返回实体的 PlayerData；如果实体不是玩家类型则返回 nil。
func (s *EntityStore) Player(id EntityID) *PlayerData {
	if e := s.entities[id]; e != nil {
		return e.Player
	}
	return nil
}

// Exit 返回实体的 ExitData；如果实体不是出口类型则返回 nil。
func (s *EntityStore) Exit(id EntityID) *ExitData {
	if e := s.entities[id]; e != nil {
		return e.Exit
	}
	return nil
}

// Tag 检查指定实体是否持有指定 tag。
func (s *EntityStore) Tag(id EntityID, tagID TagID) bool {
	set, ok := s.tagIndex[tagID]
	if !ok {
		return false
	}
	_, tagged := set[id]
	return tagged
}

// Entities 返回全部实体的 EntityID 切片。
func (s *EntityStore) Entities() []EntityID {
	ids := make([]EntityID, 0, len(s.entities))
	for id := range s.entities {
		ids = append(ids, id)
	}
	return ids
}
