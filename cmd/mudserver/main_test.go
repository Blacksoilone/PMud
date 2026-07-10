package main

import "testing"

func TestBuildGame_usesTutorialContentSnapshot(t *testing.T) {
	// Given / When
	game := buildGame()

	// Then
	observation, ok := game.Look(game.StartRoom())
	if !ok {
		t.Fatal("expected start room to exist")
	}
	if observation.Name != "练习场入口" {
		t.Fatalf("expected tutorial start room name, got %q", observation.Name)
	}
	if len(observation.Items) != 1 || observation.Items[0] != "旧油灯" {
		t.Fatalf("expected old lantern in start room, got %v", observation.Items)
	}
}
