package world

import "fmt"

type TagID string

type TagScope string

const (
	TagScopeItem TagScope = "item"
	TagScopeRoom TagScope = "room"
)

type TagFieldType string

const (
	TagFieldString TagFieldType = "string"
	TagFieldBool   TagFieldType = "bool"
	TagFieldInt    TagFieldType = "int"
	TagFieldRef    TagFieldType = "ref"
)

type TagField struct {
	Name     string
	Type     TagFieldType
	Required bool
	Default  any
}

type HookPhase int

const (
	HookPreAction  HookPhase = iota // 执行前检查，可阻断
	HookPostAction                   // 执行后附加事件
)

type TagHook struct {
	Phase   HookPhase
	Verbs   []string // 空 = 所有动词
	Handler func(ctx *AttemptContext, params map[string]any)
}

type TagDefinition struct {
	ID          TagID
	Description string
	Scopes      []TagScope
	Fields      []TagField
	Hooks       []TagHook
	Observable  bool // 设为 true 时，该 tag 会在 examine 中展示
}

type TagInstance struct {
	DefinitionID TagID
	Params       map[string]any
}

func (i Item) tagParams(tagID TagID) (map[string]any, bool) {
	for _, inst := range i.Tags {
		if inst.DefinitionID == tagID {
			return inst.Params, true
		}
	}
	return nil, false
}

// PartTag 检查指定 part 上是否存在指定 tag。
func (i Item) PartTag(partID string, tagID TagID) bool {
	part, ok := i.Parts[partID]
	if !ok {
		return false
	}
	for _, inst := range part.Tags {
		if inst.DefinitionID == tagID {
			return true
		}
	}
	return false
}

// AnyPartTag 检查任意 part 上是否存在指定 tag。
func (i Item) AnyPartTag(tagID TagID) bool {
	for _, part := range i.Parts {
		for _, inst := range part.Tags {
			if inst.DefinitionID == tagID {
				return true
			}
		}
	}
	return false
}

// ObservableTagDescriptions 收集 item 根和所有 part 上 Observable=true 的 tag 描述文本。
func (i Item) ObservableTagDescriptions(w *World) (rootTags []string, partTags map[string][]string) {
	for _, inst := range i.Tags {
		def, ok := w.TagDefinition(inst.DefinitionID)
		if ok && def.Observable {
			rootTags = append(rootTags, def.Description)
		}
	}
	for partID, part := range i.Parts {
		for _, inst := range part.Tags {
			def, ok := w.TagDefinition(inst.DefinitionID)
			if ok && def.Observable {
				if partTags == nil {
					partTags = make(map[string][]string)
				}
				partTags[partID] = append(partTags[partID], def.Description)
			}
		}
	}
	return
}

func (w *World) RegisterTag(def TagDefinition) error {
	if _, exists := w.tagDefinitions[def.ID]; exists {
		return fmt.Errorf("tag %q already registered", def.ID)
	}
	seen := make(map[string]bool, len(def.Fields))
	for _, f := range def.Fields {
		if seen[f.Name] {
			return fmt.Errorf("tag %q: duplicate field %q", def.ID, f.Name)
		}
		seen[f.Name] = true
	}
	w.tagDefinitions[def.ID] = def
	return nil
}

func (w *World) TagDefinition(id TagID) (TagDefinition, bool) {
	def, ok := w.tagDefinitions[id]
	return def, ok
}

func (w *World) TagDefinitions() map[TagID]TagDefinition {
	result := make(map[TagID]TagDefinition, len(w.tagDefinitions))
	for id, def := range w.tagDefinitions {
		result[id] = def
	}
	return result
}

func builtinTagDefs() []TagDefinition {
	return []TagDefinition{
		{
			ID:          "tag.exit",
			Description: "Exit to another room",
			Scopes:      []TagScope{TagScopeItem},
			Fields: []TagField{
				{Name: "direction", Type: TagFieldString, Required: false, Default: ""},
				{Name: "target", Type: TagFieldRef, Required: true},
			},
		},
		{
			ID:          "tag.carryable",
			Description: "Can be picked up",
			Scopes:      []TagScope{TagScopeItem},
			Fields:      nil,
		},
		{
			ID:          "tag.container",
			Description: "Can hold other items",
			Scopes:      []TagScope{TagScopeItem},
			Fields: []TagField{
				{Name: "capacity", Type: TagFieldInt, Required: false, Default: int(1)},
				{Name: "openable", Type: TagFieldBool, Required: false, Default: true},
			},
		},
		{
			ID:          "tag.lightable",
			Description: "Can be lit to provide light",
			Scopes:      []TagScope{TagScopeItem},
			Fields: []TagField{
				{Name: "lit_verb", Type: TagFieldString, Required: false, Default: "burns brightly"},
			},
		},
		{
			ID:          "tag.lockable",
			Description: "Requires a key item to pass through",
			Scopes:      []TagScope{TagScopeItem},
			Fields: []TagField{
				{Name: "key_item_id", Type: TagFieldRef, Required: true},
			},
			Hooks: []TagHook{
				{
					Phase: HookPreAction,
					Verbs: []string{"move"},
					Handler: func(ctx *AttemptContext, params map[string]any) {
						keyID, _ := params["key_item_id"].(string)
						if keyID == "" {
							return
						}
						if !ctx.World.PlayerHasItem(ctx.PlayerID, ItemID(keyID)) {
							ctx.Blocked = true
							ctx.BlockReason = "locked"
						}
					},
				},
			},
		},
	}
}
