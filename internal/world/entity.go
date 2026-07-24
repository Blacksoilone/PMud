package world

// EntityID 是所有实体在 EntityStore 中的统一标识。
// 命名惯例：room:tutorial.hall  item:practice_sword  player:alice  exit:room.hall.north
type EntityID = string

// RoomData 是房间类型实体的专属数据。
type RoomData struct {
	Dark bool
}

// ItemData 是物品类型实体的专属数据。
type ItemData struct {
	Weight int
	Volume int
}

// PlayerData 是玩家类型实体的专属数据。
type PlayerData struct {
	MaxWeight int
	MaxVolume int
}

// ExitData 是出口类型实体的专属数据。
// 出口是一等实体，不再寄生在物品的 tag.exit 上。
type ExitData struct {
	Direction    string   // "north"、"portal"、"trapdoor" 等
	TargetRoomID EntityID // 目标房间 ID
}

// Entity 是所有世界实体的统一表示。
// 四个类型指针字段最多一个非 nil：Entity 不是 Room 就是 Item 就是 Player 就是 Exit。
type Entity struct {
	ID             EntityID
	InnerName      string
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Tags           []TagInstance
	Aliases        []string
	Parts          map[string]ItemPart

	Room   *RoomData
	Item   *ItemData
	Player *PlayerData
	Exit   *ExitData
}
