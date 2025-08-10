package engine_test

import (
	"context"
	"testing"

	"go.rumenx.com/sixtysix"
	"go.rumenx.com/sixtysix/engine"
	"go.rumenx.com/sixtysix/store"
)

func TestEngine_CreateGetApplyListDelete(t *testing.T) {
	mem := store.NewMemory()
	e := engine.New(mem)
	e.Register(sixtysix.Game{})

	// create session
	s, err := e.CreateSession(context.Background(), "sixtysix", 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if s.GameName != "sixtysix" || s.Version != 1 {
		t.Fatalf("unexpected session: %+v", s)
	}

	// get session
	gs, err := e.GetSession(context.Background(), s.ID)
	if err != nil || gs.ID != s.ID {
		t.Fatalf("get: %v", err)
	}

	// apply an action that requires no payload to validate engine flow
	if _, err := e.ApplyAction(context.Background(), s.ID, engine.Action{Type: sixtysix.ActionCloseStock}); err != nil {
		t.Fatalf("apply closeStock: %v", err)
	}

	// list
	list, err := e.ListSessions(context.Background(), "sixtysix", 0, 10)
	if err != nil || len(list) == 0 {
		t.Fatalf("list: %v len=%d", err, len(list))
	}

	// delete
	if err := e.DeleteSession(context.Background(), s.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
