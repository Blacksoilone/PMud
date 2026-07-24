package world

import (
	"testing"

	"PMud/internal/content"
	"PMud/internal/presentation"
)

func TestLoop_recoverPanicFromVerbHandler(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	w := NewFromSnapshot(compiled.Server, compiled.Client)
	l := NewLoop(w)
	l.RegisterVerb("panic", func(l2 *Loop, ctx *AttemptContext) {
		panic("test panic from verb handler")
	})
	l.Start()

	resp := make(chan ActionResult, 1)
	l.Submit(Action{
		PlayerID: "player.test",
		Verb:     "panic",
		Resp:     resp,
	})

	result := <-resp
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
	sysMsg, ok := result.Events[0].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected SystemMessageEvent, got %T", result.Events[0])
	}
	if sysMsg.MessageKey != "system.internal_error" {
		t.Fatalf("message key = %q, want system.internal_error", sysMsg.MessageKey)
	}
}

func TestLoop_customItemResolver_overridesFallback(t *testing.T) {
	w := New()
	l := NewLoop(w)

	const testBlockReason = "custom resolver blocked"

	l.RegisterItemResolver("custom", func(l2 *Loop, ctx *AttemptContext) []*Entity {
		ent := l2.world.store.Get("item.tutorial.old_lantern")
		if ent != nil {
			return []*Entity{ent}
		}
		return nil
	})

	// 在 tag.lightable 上加一个 hook，用于检查 resolver 是否生效
	origDef, _ := w.TagDefinition("tag.lightable")
	origDef.Hooks = append(origDef.Hooks, TagHook{
		Phase: HookPreAction,
		Verbs: []string{"custom"},
		Handler: func(ctx *AttemptContext, params map[string]any) {
			ctx.Blocked = true
			ctx.BlockReason = testBlockReason
			ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "custom.blocked"})
		},
	})
	w.tagDefinitions["tag.lightable"] = origDef

	// 注册一个不阻塞的 verb handler
	l.RegisterVerb("custom", func(l2 *Loop, ctx *AttemptContext) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "custom.executed"})
	})
	l.Start()

	// 把玩家放入锁钥厅（旧油灯在那里）
	l.world.store.Add(&Entity{
		ID: "player.test",
		Player: &PlayerData{MaxWeight: 20, MaxVolume: 10},
	})
	l.world.store.PlaceInRoom("player.test", "room.tutorial.lock_hall")

	resp := make(chan ActionResult, 1)
	l.Submit(Action{PlayerID: "player.test", Verb: "custom", Resp: resp})

	result := <-resp
	if len(result.Events) == 0 {
		t.Fatal("expected at least one event")
	}
	sysMsg, ok := result.Events[0].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected SystemMessageEvent, got %T", result.Events[0])
	}
	if sysMsg.MessageKey != "custom.blocked" {
		t.Fatalf("expected pre-hook to block via custom resolver, got message key = %q", sysMsg.MessageKey)
	}
}

func TestLoop_relevantItemsFallback_visibleItemsWithMatchingHooks(t *testing.T) {
	w := New()
	l := NewLoop(w)

	const testBlockReason = "fallback hook blocked"

	origDef, _ := w.TagDefinition("tag.lightable")
	origDef.Hooks = append(origDef.Hooks, TagHook{
		Phase: HookPreAction,
		Verbs: []string{"testfallback"},
		Handler: func(ctx *AttemptContext, params map[string]any) {
			ctx.Blocked = true
			ctx.BlockReason = testBlockReason
			ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "light.blocked"})
		},
	})
	w.tagDefinitions["tag.lightable"] = origDef

	l.RegisterVerb("testfallback", func(l2 *Loop, ctx *AttemptContext) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "light.executed"})
	})
	l.Start()

	l.world.store.Add(&Entity{
		ID: "player.test",
		Player: &PlayerData{MaxWeight: 20, MaxVolume: 10},
	})
	l.world.store.PlaceInRoom("player.test", "room.tutorial.lock_hall")

	resp := make(chan ActionResult, 1)
	l.Submit(Action{PlayerID: "player.test", Verb: "testfallback", Resp: resp})

	result := <-resp
	if len(result.Events) == 0 {
		t.Fatal("expected at least one event")
	}
	sysMsg, ok := result.Events[0].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected SystemMessageEvent, got %T", result.Events[0])
	}
	if sysMsg.MessageKey != "light.blocked" {
		t.Fatalf("expected hook to block via fallback, got message key = %q", sysMsg.MessageKey)
	}
}
