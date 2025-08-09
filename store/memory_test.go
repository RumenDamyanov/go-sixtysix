package store_test

import (
	"context"
	"testing"
	"time"

	"go.rumenx.com/sixtysix/engine"
	"go.rumenx.com/sixtysix/store"
)

func TestMemory_CRUD(t *testing.T) {
	m := store.NewMemory()
	now := time.Now().UTC()
	s := engine.Session{ID: "a", GameName: "g", Version: 1, CreatedAt: now, UpdatedAt: now}

	// create
	if err := m.Create(context.Background(), s); err != nil {
		t.Fatalf("create: %v", err)
	}
	// duplicate
	if err := m.Create(context.Background(), s); err == nil {
		t.Fatalf("expected duplicate error")
	}

	// get
	got, ok, err := m.Get(context.Background(), "a")
	if err != nil || !ok || got.ID != "a" {
		t.Fatalf("get: %v ok=%v got=%+v", err, ok, got)
	}

	// update
	got.Version = 2
	if err := m.Update(context.Background(), got); err != nil {
		t.Fatalf("update: %v", err)
	}

	// list
	list, err := m.List(context.Background(), "g", 0, 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v len=%d", err, len(list))
	}

	// delete
	if err := m.Delete(context.Background(), "a"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
