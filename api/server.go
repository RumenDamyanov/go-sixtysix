package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go.rumenx.com/sixtysix/engine"
)

// Server is a minimal HTTP server exposing the engine.
type Server struct {
	Engine *engine.Engine
	mux    *http.ServeMux
}

func New(e *engine.Engine) *Server {
	s := &Server{Engine: e, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// GET /games -> list
	s.mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"games": s.Engine.Games()})
	})

	// POST /sessions?game=NAME&seed=0
	s.mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			game := r.URL.Query().Get("game")
			if game == "" {
				http.Error(w, "missing game", http.StatusBadRequest)
				return
			}
			seedStr := r.URL.Query().Get("seed")
			var seed int64
			if seedStr != "" {
				if v, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
					seed = v
				}
			}
			sess, err := s.Engine.CreateSession(r.Context(), game, seed)
			if err != nil {
				handleEngineError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, sess)
		case http.MethodGet:
			game := r.URL.Query().Get("game")
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			list, err := s.Engine.ListSessions(r.Context(), game, offset, limit)
			if err != nil {
				handleEngineError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"sessions": list})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET/POST/DELETE /sessions/{id}
	s.mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/sessions/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sess, err := s.Engine.GetSession(r.Context(), id)
			if err != nil {
				handleEngineError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, sess)
		case http.MethodPost: // apply action
			var a engine.Action
			if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			sess, err := s.Engine.ApplyAction(context.Background(), id, a)
			if err != nil {
				handleEngineError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, sess)
		case http.MethodDelete:
			if err := s.Engine.DeleteSession(r.Context(), id); err != nil {
				handleEngineError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func handleEngineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, engine.ErrGameNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, engine.ErrSessionNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, engine.ErrConflict):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
