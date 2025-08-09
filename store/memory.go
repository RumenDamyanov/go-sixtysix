package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.rumenx.com/sixtysix/engine"
)

// Memory is a threadsafe in-memory store useful for tests and small deployments.
type Memory struct {
	mu       sync.RWMutex
	sessions map[string]engine.Session
}

func NewMemory() *Memory {
	return &Memory{sessions: make(map[string]engine.Session)}
}

func (m *Memory) Create(ctx context.Context, s engine.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[s.ID]; ok {
		return errors.New("store: duplicate id")
	}
	// shallow copy
	m.sessions[s.ID] = s
	return nil
}

func (m *Memory) Get(ctx context.Context, id string) (engine.Session, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok, nil
}

func (m *Memory) Update(ctx context.Context, s engine.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[s.ID]; !ok {
		return engine.ErrSessionNotFound
	}
	s.UpdatedAt = time.Now().UTC()
	m.sessions[s.ID] = s
	return nil
}

func (m *Memory) List(ctx context.Context, gameName string, offset, limit int) ([]engine.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]engine.Session, 0)
	for _, s := range m.sessions {
		if gameName == "" || s.GameName == gameName {
			all = append(all, s)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.Before(all[j].CreatedAt) })
	if offset > len(all) {
		return []engine.Session{}, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *Memory) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[id]; !ok {
		return engine.ErrSessionNotFound
	}
	delete(m.sessions, id)
	return nil
}
