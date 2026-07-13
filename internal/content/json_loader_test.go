package content

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSourceLoadsValidJSON(t *testing.T) {
	path := writeSourceJSON(t, `{
		"StartRoomID": "room.test.start",
		"Rooms": [
			{
				"ID": "room.test.start",
				"NameKey": "room.test.start.name",
				"DescriptionKey": "room.test.start.description",
				"Exits": {"north": "room.test.yard"}
			}
		],
		"Items": [
			{
				"ID": "item.test.lantern",
				"NameKey": "item.test.lantern.name",
				"InitialRoom": "room.test.start"
			}
		],
		"Text": {
			"room.test.start.name": "入口",
			"room.test.start.description": "一处入口。",
			"item.test.lantern.name": "油灯"
		}
	}`)

	source, err := LoadSource(path)

	if err != nil {
		t.Fatalf("LoadSource: %v", err)
	}
	if source.StartRoomID != "room.test.start" {
		t.Fatalf("StartRoomID = %q, want room.test.start", source.StartRoomID)
	}
	if len(source.Rooms) != 1 || source.Rooms[0].Exits["north"] != "room.test.yard" {
		t.Fatalf("Rooms = %#v, want start room with north exit", source.Rooms)
	}
	if len(source.Items) != 1 || source.Items[0].InitialRoom != "room.test.start" {
		t.Fatalf("Items = %#v, want lantern in start room", source.Items)
	}
	compiled, err := Compile(source)
	if err != nil {
		t.Fatalf("Compile loaded source: %v", err)
	}
	if compiled.Client.Text["item.test.lantern.name"] != "油灯" {
		t.Fatalf("loaded item text = %q, want 油灯", compiled.Client.Text["item.test.lantern.name"])
	}
}

func TestLoadSourceReturnsMalformedJSONError(t *testing.T) {
	path := writeSourceJSON(t, `{`)

	_, err := LoadSource(path)

	if err == nil {
		t.Fatalf("LoadSource error = nil, want malformed JSON error")
	}
}

func TestLoadSourceReturnsMissingFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	_, err := LoadSource(path)

	if err == nil {
		t.Fatalf("LoadSource error = nil, want missing file error")
	}
}

func TestLoadTutorialSourceJSONMatchesFixture(t *testing.T) {
	loaded, err := LoadSource(filepath.Join("..", "..", "data", "tutorial", "source.json"))
	if err != nil {
		t.Fatalf("LoadSource tutorial JSON: %v", err)
	}
	loadedCompiled, err := Compile(loaded)
	if err != nil {
		t.Fatalf("Compile loaded tutorial: %v", err)
	}
	fixtureCompiled, err := Compile(TutorialSource())
	if err != nil {
		t.Fatalf("Compile fixture tutorial: %v", err)
	}

	if !reflect.DeepEqual(loadedCompiled, fixtureCompiled) {
		t.Fatalf("compiled tutorial JSON differs from fixture\nloaded: %#v\nfixture: %#v", loadedCompiled, fixtureCompiled)
	}
}

func writeSourceJSON(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.json")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
