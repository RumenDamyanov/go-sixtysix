package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// Action is a generic instruction sent by a client/actor.
type Action struct {
	Type    string         `json:"type"`
	Actor   string         `json:"actor,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
	// Client-provided idempotency key to safely retry requests
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// Game defines the logic for a particular game.
// Implementations must be deterministic and pure: given the same input state and action, they must return the same output state.
type Game interface {
	// Name returns a stable, unique name for the game (e.g., "tictactoe").
	Name() string
	// InitialState creates the starting state. The seed can be used for deterministic randomness.
	InitialState(seed int64) any
	// Validate checks whether an action is valid given the state. Should not mutate state.
	Validate(state any, action Action) error
	// Apply returns a new state after applying action. Must not mutate the input state.
	Apply(state any, action Action) (any, error)
}

// Session represents a single instance of a game.
type Session struct {
	ID        string    `json:"id"`
	GameName  string    `json:"gameName"`
	State     any       `json:"state"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store abstracts persistence for sessions.
type Store interface {
	Create(ctx context.Context, s Session) error
	Get(ctx context.Context, id string) (Session, bool, error)
	Update(ctx context.Context, s Session) error
	List(ctx context.Context, gameName string, offset, limit int) ([]Session, error)
	Delete(ctx context.Context, id string) error
}

var (
	ErrGameNotFound    = errors.New("engine: game not found")
	ErrSessionNotFound = errors.New("engine: session not found")
	ErrConflict        = errors.New("engine: conflict")
)

// Engine wires games with storage and provides a simple API to manipulate sessions.
type Engine struct {
	store Store
	mu    sync.RWMutex
	games map[string]Game
}

func New(store Store) *Engine {
	return &Engine{store: store, games: make(map[string]Game)}
}

// Register adds a game. Panics if a game with the same name already exists.
func (e *Engine) Register(g Game) {
	e.mu.Lock()
	defer e.mu.Unlock()
	name := g.Name()
	if name == "" {
		panic("engine: game must have a non-empty Name")
	}
	if _, exists := e.games[name]; exists {
		panic("engine: duplicate game name: " + name)
	}
	e.games[name] = g
}

// Games returns a snapshot of registered game names.
func (e *Engine) Games() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.games))
	for k := range e.games {
		out = append(out, k)
	}
	return out
}

// CreateSession creates a new session for the named game.
func (e *Engine) CreateSession(ctx context.Context, gameName string, seed int64) (Session, error) {
	e.mu.RLock()
	g, ok := e.games[gameName]
	e.mu.RUnlock()
	if !ok {
		return Session{}, ErrGameNotFound
	}
	id := randomID()
	now := time.Now().UTC()
	s := Session{
		ID:        id,
		GameName:  gameName,
		State:     g.InitialState(seed),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := e.store.Create(ctx, s); err != nil {
		return Session{}, err
	}
	return s, nil
}

// GetSession returns a session by id.
func (e *Engine) GetSession(ctx context.Context, id string) (Session, error) {
	s, ok, err := e.store.Get(ctx, id)
	if err != nil {
		return Session{}, err
	}
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return s, nil
}

// ApplyAction validates and applies an action to the session state.
func (e *Engine) ApplyAction(ctx context.Context, id string, action Action) (Session, error) {
	s, ok, err := e.store.Get(ctx, id)
	if err != nil {
		return Session{}, err
	}
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	e.mu.RLock()
	g, ok := e.games[s.GameName]
	e.mu.RUnlock()
	if !ok {
		return Session{}, ErrGameNotFound
	}
	if err := g.Validate(s.State, action); err != nil {
		return Session{}, err
	}
	newState, err := g.Apply(s.State, action)
	if err != nil {
		return Session{}, err
	}
	s.State = newState
	s.Version++
	s.UpdatedAt = time.Now().UTC()
	if err := e.store.Update(ctx, s); err != nil {
		return Session{}, err
	}
	return s, nil
}

// ListSessions returns sessions for a given game.
func (e *Engine) ListSessions(ctx context.Context, gameName string, offset, limit int) ([]Session, error) {
	return e.store.List(ctx, gameName, offset, limit)
}

func (e *Engine) DeleteSession(ctx context.Context, id string) error {
	return e.store.Delete(ctx, id)
}

func randomID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
