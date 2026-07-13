package main

import "testing"

func TestParseArgsUsesDefaultLineMode(t *testing.T) {
	config := parseArgs([]string{"mudclient"})

	if config.address != defaultAddress {
		t.Fatalf("address = %q, want %q", config.address, defaultAddress)
	}
	if config.tui {
		t.Fatalf("tui = true, want false")
	}
}

func TestParseArgsUsesAddressOverride(t *testing.T) {
	config := parseArgs([]string{"mudclient", "127.0.0.1:5000"})

	if config.address != "127.0.0.1:5000" {
		t.Fatalf("address = %q, want override", config.address)
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
