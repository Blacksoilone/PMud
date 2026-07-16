package main

import "testing"

func TestParseArgsUsesDefaultTUI(t *testing.T) {
	config := parseArgs([]string{"mudclient"})

	if config.address != defaultAddress {
		t.Fatalf("address = %q, want %q", config.address, defaultAddress)
	}
	if !config.tui {
		t.Fatalf("tui = false, want true")
	}
}

func TestParseArgsUsesAddressOverride(t *testing.T) {
	config := parseArgs([]string{"mudclient", "127.0.0.1:5000"})

	if config.address != "127.0.0.1:5000" {
		t.Fatalf("address = %q, want override", config.address)
	}
	if !config.tui {
		t.Fatalf("tui = false, want true")
	}
}

func TestParseArgsDisablesTUIWithLineMode(t *testing.T) {
	config := parseArgs([]string{"mudclient", "--line"})

	if config.address != defaultAddress {
		t.Fatalf("address = %q, want %q", config.address, defaultAddress)
	}
	if config.tui {
		t.Fatalf("tui = true, want false")
	}
}

func TestParseArgsEnablesTUIWithDefaultAddress(t *testing.T) {
	config := parseArgs([]string{"mudclient", "--tui"})

	if config.address != defaultAddress {
		t.Fatalf("address = %q, want %q", config.address, defaultAddress)
	}
	if !config.tui {
		t.Fatalf("tui = false, want true")
	}
}

func TestParseArgsEnablesTUIWithAddressOverride(t *testing.T) {
	config := parseArgs([]string{"mudclient", "--tui", "127.0.0.1:5000"})

	if config.address != "127.0.0.1:5000" {
		t.Fatalf("address = %q, want override", config.address)
	}
	if !config.tui {
		t.Fatalf("tui = false, want true")
	}
}

func TestLoadClientCatalogLoadsTutorialData(t *testing.T) {
	catalog, err := loadClientCatalog("../../data/tutorial/source.json")
	if err != nil {
		t.Fatalf("loadClientCatalog: %v", err)
	}

	startNameKey := catalog.RoomNames["room.tutorial.start"]
	if catalog.Text[startNameKey] != "练习场入口" {
		t.Fatalf("start room name = %q, want 练习场入口", catalog.Text[startNameKey])
	}
	oldLanternKey := catalog.ItemDisplayNames["item.tutorial.old_lantern"]
	if catalog.Text[oldLanternKey] != "旧油灯" {
		t.Fatalf("old lantern name = %q, want 旧油灯", catalog.Text[oldLanternKey])
	}
	oldLanternInnerKey := catalog.ItemInnerNames["item.tutorial.old_lantern"]
	if catalog.Text[oldLanternInnerKey] != "old lantern" {
		t.Fatalf("old lantern inner name = %q, want old lantern", catalog.Text[oldLanternInnerKey])
	}
}
