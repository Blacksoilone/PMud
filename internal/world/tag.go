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

type TagDefinition struct {
	ID          TagID
	Description string
	Scopes      []TagScope
	Fields      []TagField
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
		},
	}
}
